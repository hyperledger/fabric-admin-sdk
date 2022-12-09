package test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fabric-admin-sdk/internal/network"
	"fabric-admin-sdk/pkg/chaincode"
	"fabric-admin-sdk/pkg/channel"
	"fabric-admin-sdk/pkg/identity"
	"fabric-admin-sdk/pkg/tools"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
)

const (
	org1PeerAddress = "localhost:7051"
	org2PeerAddress = "localhost:9051"
)

type ConnectionDetails struct {
	id         identity.SigningIdentity
	connection grpc.ClientConnInterface
	client     peer.EndorserClient
}

var _ = Describe("e2e", func() {
	Context("the e2e test with test network", func() {
		It("should work", func(specCtx SpecContext) {
			//genesis block
			_, err := os.Stat("../../fabric-samples/test-network")
			if err != nil {
				Skip("skip for unit test")
			}
			profile, err := tools.LoadProfile("TwoOrgsApplicationGenesis", "./")
			Expect(err).NotTo(HaveOccurred())
			Expect(profile).ToNot(BeNil())
			Expect(profile.Orderer.BatchSize.MaxMessageCount).To(Equal(uint32(10)))
			block, err := tools.ConfigTxGen(profile, "mychannel")
			Expect(err).NotTo(HaveOccurred())
			Expect(block).ToNot(BeNil())

			//create channel
			var caFile, clientCert, clientKey, osnURL string
			osnURL = "https://localhost:7053"
			caFile = "../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
			clientCert = "../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt"
			clientKey = "../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key"
			caCertPool := x509.NewCertPool()
			caFilePEM, err := os.ReadFile(caFile)
			caCertPool.AppendCertsFromPEM(caFilePEM)
			Expect(err).NotTo(HaveOccurred())
			tlsClientCert, err := tls.LoadX509KeyPair(clientCert, clientKey)
			Expect(err).NotTo(HaveOccurred())
			resp, err := channel.CreateChannel(osnURL, block, caCertPool, tlsClientCert)
			Expect(err).NotTo(HaveOccurred())
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("my Http error is ", err)
			}
			fmt.Println("response statuscode is ", resp.StatusCode,
				"\nhead[name]=", resp.Header["Name"],
				"\nbody is ", string(body))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusCreated))

			//join peer1
			TLSCACert := "../../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
			PrivKeyPath := "../../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk"
			SignCert := "../../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem"
			MSPID := "Org1MSP"

			peer1 := network.Node{
				Addr:      org1PeerAddress,
				TLSCACert: TLSCACert,
			}
			err = peer1.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			n_conn1, err := network.DialConnection(peer1)
			Expect(err).NotTo(HaveOccurred())

			connection1 := peer.NewEndorserClient(n_conn1)
			Expect(err).NotTo(HaveOccurred())
			org1MSP, err := tools.CreateSigner(PrivKeyPath, SignCert, MSPID)
			Expect(err).NotTo(HaveOccurred())
			err = channel.JoinChannel(
				block, org1MSP, connection1,
			)
			Expect(err).NotTo(HaveOccurred())

			//join peer2
			TLSCACert = "../../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
			PrivKeyPath = "../../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/keystore/priv_sk"
			SignCert = "../../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/signcerts/Admin@org2.example.com-cert.pem"
			MSPID = "Org2MSP"

			peer2 := network.Node{
				Addr:      org2PeerAddress,
				TLSCACert: TLSCACert,
			}
			err = peer2.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			n_conn2, err := network.DialConnection(peer2)
			Expect(err).NotTo(HaveOccurred())
			connection2 := peer.NewEndorserClient(n_conn2)
			Expect(err).NotTo(HaveOccurred())
			org2MSP, err := tools.CreateSigner(PrivKeyPath, SignCert, MSPID)
			Expect(err).NotTo(HaveOccurred())
			err = channel.JoinChannel(
				block, org2MSP, connection2,
			)
			Expect(err).NotTo(HaveOccurred())

			// package chaincode as CCAAS
			dummyConnection := chaincode.Connection{
				Address:      "{{.peername}}_basic:9999",
				Dial_timeout: "10s",
				Tls_required: false,
			}
			dummyMeta := chaincode.Metadata{
				Type:  "ccaas",
				Label: "basic_1.0",
			}
			packageFileName := "basic-asset.tar.gz"
			err = chaincode.PackageCCAAS(dummyConnection, dummyMeta, tmpDir, packageFileName)
			Expect(err).NotTo(HaveOccurred())

			packageFilePath := path.Join(tmpDir, packageFileName)
			packageID, err := chaincode.PackageID(packageFilePath)
			Expect(err).NotTo(HaveOccurred(), "get chaincode package ID")
			fmt.Println(packageID)

			peerConnections := []*ConnectionDetails{
				{
					client:     connection1,
					connection: n_conn1,
					id:         org1MSP,
				},
				{
					client:     connection2,
					connection: n_conn2,
					id:         org2MSP,
				},
			}

			var installWaitGroup sync.WaitGroup
			for _, installTarget := range peerConnections {
				installWaitGroup.Add(1)

				go func(target *ConnectionDetails) {
					defer installWaitGroup.Done()

					packageFile, err := os.Open(packageFilePath)
					Expect(err).NotTo(HaveOccurred(), "open chaincode package file")

					ctx, cancel := context.WithTimeout(specCtx, 2*time.Minute)
					defer cancel()

					err = chaincode.Install(ctx, target.connection, target.id, packageFile)
					Expect(err).NotTo(HaveOccurred(), "chaincode install")
				}(installTarget)
			}

			installWaitGroup.Wait()

			var queryInstalledWaitGroup sync.WaitGroup
			for _, queryInstalledTarget := range peerConnections {
				queryInstalledWaitGroup.Add(1)

				go func(target *ConnectionDetails) {
					defer queryInstalledWaitGroup.Done()

					ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
					defer cancel()

					result, err := chaincode.QueryInstalled(ctx, target.connection, target.id)
					Expect(err).NotTo(HaveOccurred(), "query installed chaincode")

					installedChaincodes := result.GetInstalledChaincodes()
					Expect(installedChaincodes).To(HaveLen(1), "number of installed chaincodes")
					Expect(installedChaincodes[0].GetPackageId()).To(Equal(packageID), "installed chaincode package ID")
					Expect(installedChaincodes[0].GetLabel()).To(Equal(dummyMeta.Label), "installed chaincode label")
				}(queryInstalledTarget)
			}

			queryInstalledWaitGroup.Wait()

			//cc define
			CCDefine := chaincode.CCDefine{
				ChannelID:                "mychannel",
				InputTxID:                "",
				PackageID:                "",
				Name:                     "basic",
				Version:                  "1.0",
				EndorsementPlugin:        "",
				ValidationPlugin:         "",
				Sequence:                 1,
				ValidationParameterBytes: nil,
				InitRequired:             false,
				CollectionConfigPackage:  nil,
			}
			// orderer
			orderer_addr := "localhost:7050"
			orderer_TLSCACert := "../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem"
			orderer_node := network.Node{
				Addr:      orderer_addr,
				TLSCACert: orderer_TLSCACert,
			}
			err = orderer_node.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			// approve from org2
			time.Sleep(time.Duration(15) * time.Second)
			endorsement_org2_group := make([]peer.EndorserClient, 1)
			endorsement_org2_group[0] = connection2
			n_conn3, err := network.DialConnection(orderer_node)
			Expect(err).NotTo(HaveOccurred())
			connection3, err := orderer.NewAtomicBroadcastClient(n_conn3).Broadcast(context.Background())
			Expect(err).NotTo(HaveOccurred())
			err = chaincode.Approve(CCDefine, org2MSP, endorsement_org2_group, connection3)
			Expect(err).NotTo(HaveOccurred())
			// ReadinessCheck from org2
			time.Sleep(time.Duration(15) * time.Second)
			err = chaincode.ReadinessCheck(CCDefine, org2MSP, connection2)
			Expect(err).NotTo(HaveOccurred())

			// approve from org1
			endorsement_org1_group := make([]peer.EndorserClient, 1)
			endorsement_org1_group[0] = connection1
			err = chaincode.Approve(CCDefine, org1MSP, endorsement_org1_group, connection3)
			Expect(err).NotTo(HaveOccurred())
			// ReadinessCheck from org1
			time.Sleep(time.Duration(15) * time.Second)
			err = chaincode.ReadinessCheck(CCDefine, org1MSP, connection1)
			Expect(err).NotTo(HaveOccurred())

			// commit from org1
			n_conn3, err = network.DialConnection(orderer_node)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Duration(15) * time.Second)
			connection3, err = orderer.NewAtomicBroadcastClient(n_conn3).Broadcast(context.Background())
			Expect(err).NotTo(HaveOccurred())
			err = chaincode.Commit(CCDefine, org1MSP, endorsement_org1_group, connection3)
			Expect(err).NotTo(HaveOccurred())

			// commit from org2
			time.Sleep(time.Duration(15) * time.Second)
			connection3, err = orderer.NewAtomicBroadcastClient(n_conn3).Broadcast(context.Background())
			Expect(err).NotTo(HaveOccurred())
			err = chaincode.Commit(CCDefine, org2MSP, endorsement_org2_group, connection3)
			Expect(err).NotTo(HaveOccurred())

			f, _ := os.Create("PackageID")
			_, err = io.WriteString(f, packageID)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
