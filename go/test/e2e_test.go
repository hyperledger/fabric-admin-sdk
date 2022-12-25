package test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
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

	"github.com/hyperledger/fabric-gateway/pkg/client"
	gatewaypb "github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	org1PeerAddress = "localhost:7051"
	org2PeerAddress = "localhost:9051"
	channelName     = "mychannel"
)

type ConnectionDetails struct {
	id         identity.SigningIdentity
	connection grpc.ClientConnInterface
	client     peer.EndorserClient
}

func runParallel[T any](args []T, f func(T)) {
	var wg sync.WaitGroup
	for _, arg := range args {
		wg.Add(1)

		go func(target T) {
			defer GinkgoRecover()
			defer wg.Done()
			f(target)
		}(arg)
	}

	wg.Wait()
}

func printGrpcError(err error) {
	if err == nil {
		return
	}

	switch err := err.(type) {
	case *client.EndorseError:
		fmt.Printf("Endorse error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
	case *client.SubmitError:
		fmt.Printf("Submit error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
	case *client.CommitStatusError:
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Timeout waiting for transaction %s commit status: %s", err.TransactionID, err)
		} else {
			fmt.Printf("Error obtaining commit status for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		}
	case *client.CommitError:
		fmt.Printf("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err)
	default:
		fmt.Printf("unexpected error type %T: %s", err, err)
	}

	statusErr := status.Convert(err)

	details := statusErr.Details()
	if len(details) > 0 {
		fmt.Println("Error Details:")

		for _, detail := range details {
			switch detail := detail.(type) {
			case *gatewaypb.ErrorDetail:
				fmt.Printf("- address: %s, mspId: %s, message: %s\n", detail.Address, detail.MspId, detail.Message)
			}
		}
	}
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
			block, err := tools.ConfigTxGen(profile, channelName)
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

			runParallel(peerConnections, func(target *ConnectionDetails) {
				packageFile, err := os.Open(packageFilePath)
				Expect(err).NotTo(HaveOccurred(), "open chaincode package file")

				ctx, cancel := context.WithTimeout(specCtx, 2*time.Minute)
				defer cancel()

				err = chaincode.Install(ctx, target.connection, target.id, packageFile)
				Expect(err).NotTo(HaveOccurred(), "chaincode install")
			})

			runParallel(peerConnections, func(target *ConnectionDetails) {
				ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
				defer cancel()

				result, err := chaincode.QueryInstalled(ctx, target.connection, target.id)
				Expect(err).NotTo(HaveOccurred(), "query installed chaincode")

				installedChaincodes := result.GetInstalledChaincodes()
				Expect(installedChaincodes).To(HaveLen(1), "number of installed chaincodes")
				Expect(installedChaincodes[0].GetPackageId()).To(Equal(packageID), "installed chaincode package ID")
				Expect(installedChaincodes[0].GetLabel()).To(Equal(dummyMeta.Label), "installed chaincode label")
			})

			time.Sleep(time.Duration(15) * time.Second)

			chaincodeDefinition := &chaincode.Definition{
				Name:     "basic",
				Version:  "1.0",
				Sequence: 1,
			}

			runParallel(peerConnections, func(target *ConnectionDetails) {
				ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
				defer cancel()

				err := chaincode.Approve(ctx, target.connection, target.id, channelName, chaincodeDefinition)
				printGrpcError(err)
				Expect(err).NotTo(HaveOccurred(), "approve chaincode for org %s", target.id.MspID())
			})

			CCDefine := chaincode.CCDefine{
				ChannelID:                channelName,
				InputTxID:                "",
				PackageID:                "",
				Name:                     chaincodeDefinition.Name,
				Version:                  chaincodeDefinition.Version,
				EndorsementPlugin:        "",
				ValidationPlugin:         "",
				Sequence:                 chaincodeDefinition.Sequence,
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

			// ReadinessCheck from org2
			time.Sleep(time.Duration(15) * time.Second)
			err = chaincode.ReadinessCheck(CCDefine, org2MSP, connection2)
			Expect(err).NotTo(HaveOccurred())

			// ReadinessCheck from org1
			time.Sleep(time.Duration(15) * time.Second)
			err = chaincode.ReadinessCheck(CCDefine, org1MSP, connection1)
			Expect(err).NotTo(HaveOccurred())

			ordererConnection, err := network.DialConnection(orderer_node)
			Expect(err).NotTo(HaveOccurred())
			ordererClient, err := orderer.NewAtomicBroadcastClient(ordererConnection).Broadcast(context.Background())
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Duration(15) * time.Second)

			// commit from org1
			endorsement_org1_group := make([]peer.EndorserClient, 1)
			endorsement_org1_group[0] = connection1
			err = chaincode.Commit(CCDefine, org1MSP, endorsement_org1_group, ordererClient)
			Expect(err).NotTo(HaveOccurred())

			// commit from org2
			endorsement_org2_group := make([]peer.EndorserClient, 1)
			endorsement_org2_group[0] = connection2
			err = chaincode.Commit(CCDefine, org2MSP, endorsement_org2_group, ordererClient)
			Expect(err).NotTo(HaveOccurred())

			f, _ := os.Create("PackageID")
			_, err = io.WriteString(f, packageID)
			Expect(err).NotTo(HaveOccurred())

			ctx_1, cancel := context.WithTimeout(specCtx, 30*time.Second)
			defer cancel()
			err = chaincode.QueryCommitted(ctx_1, n_conn1, org1MSP)
			Expect(err).NotTo(HaveOccurred(), "query installed chaincode "+org1MSP.MspID())

			ctx_2, cancel := context.WithTimeout(specCtx, 30*time.Second)
			defer cancel()
			err = chaincode.QueryCommitted(ctx_2, n_conn2, org2MSP)
			Expect(err).NotTo(HaveOccurred(), "query installed chaincode "+org2MSP.MspID())
		})
	})
})
