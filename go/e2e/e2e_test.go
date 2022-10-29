package e2e_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fabric-admin-sdk/chaincode"
	"fabric-admin-sdk/channel"
	"fabric-admin-sdk/tools"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("e2e", func() {
	Context("the e2e test with test network", func() {
		It("should work", func() {
			//genesis block
			_, err := os.Stat("../../fabric-samples/test-network")
			if err != nil {
				ginkgo.Skip("skip for unit test")
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
			caFilePEM, err := ioutil.ReadFile(caFile)
			caCertPool.AppendCertsFromPEM(caFilePEM)
			Expect(err).NotTo(HaveOccurred())
			tlsClientCert, err := tls.LoadX509KeyPair(clientCert, clientKey)
			Expect(err).NotTo(HaveOccurred())
			resp, err := channel.CreateChannel(osnURL, block, caCertPool, tlsClientCert)
			Expect(err).NotTo(HaveOccurred())
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("my Http error is ", err)
			}
			fmt.Println("response statuscode is ", resp.StatusCode,
				"\nhead[name]=", resp.Header["Name"],
				"\nbody is ", string(body))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusCreated))

			//join peer1
			peer_addr := "localhost:7051"
			TLSCACert := "../../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
			PrivKeyPath := "../../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk"
			SignCert := "../../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem"
			MSPID := "Org1MSP"

			logger := log.New()
			peer1 := basic.Node{
				Addr:      peer_addr,
				TLSCACert: TLSCACert,
			}
			err = peer1.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			connection1, err := basic.CreateEndorserClient(peer1, logger)
			Expect(err).NotTo(HaveOccurred())
			org1MSP, err := tools.CreateSigner(PrivKeyPath, SignCert, MSPID)
			Expect(err).NotTo(HaveOccurred())
			err = channel.JoinChannel(
				block, *org1MSP, connection1,
			)
			Expect(err).NotTo(HaveOccurred())

			//join peer2
			peer_addr = "localhost:9051"
			TLSCACert = "../../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
			PrivKeyPath = "../../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/keystore/priv_sk"
			SignCert = "../../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/signcerts/Admin@org2.example.com-cert.pem"
			MSPID = "Org2MSP"

			peer2 := basic.Node{
				Addr:      peer_addr,
				TLSCACert: TLSCACert,
			}
			err = peer2.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			connection2, err := basic.CreateEndorserClient(peer2, logger)
			Expect(err).NotTo(HaveOccurred())
			org2MSP, err := tools.CreateSigner(PrivKeyPath, SignCert, MSPID)
			Expect(err).NotTo(HaveOccurred())
			err = channel.JoinChannel(
				block, *org2MSP, connection2,
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
			err = chaincode.PackageCCAAS(dummyConnection, dummyMeta, tmpDir, "basic-asset.tar.gz")
			Expect(err).NotTo(HaveOccurred())
			// install as CCAAS at peer1
			err = chaincode.InstallChainCode("", tmpDir+"/basic-asset.tar.gz", "basic", "1.0", *org1MSP, connection1)
			//err = chaincode.InstallChainCode("", "./basicj.tar.gz", "basic-asset", "1.0", *org1MSP, connection1)
			Expect(err).NotTo(HaveOccurred())
			// install as CCAAS at peer2
			err = chaincode.InstallChainCode("", tmpDir+"/basic-asset.tar.gz", "basic", "1.0", *org2MSP, connection2)
			//err = chaincode.InstallChainCode("", "./basicj.tar.gz", "basic-asset", "1.0", *org2MSP, connection2)
			Expect(err).NotTo(HaveOccurred())

			// approve from org1
			// orderer
			orderer_addr := "localhost:7050"
			orderer_TLSCACert := "../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem"
			orderer_node := basic.Node{
				Addr:      orderer_addr,
				TLSCACert: orderer_TLSCACert,
			}
			err = orderer_node.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			connection3, err := basic.CreateBroadcastClient(context.Background(), orderer_node, logger)
			Expect(err).NotTo(HaveOccurred())
			endorsement_org1_group := make([]pb.EndorserClient, 1)
			endorsement_org1_group[0] = connection1
			//endorsement_org1_group[1] = connection2
			err = chaincode.Approve(*org1MSP, "mychannel", "", "", "basic", "1.0", "", "", 1, nil, false, nil, endorsement_org1_group, connection3)
			Expect(err).NotTo(HaveOccurred())
			// approve from org2
			time.Sleep(time.Duration(15) * time.Second)
			connection4, err := basic.CreateBroadcastClient(context.Background(), orderer_node, logger)
			endorsement_org2_group := make([]pb.EndorserClient, 1)
			endorsement_org2_group[0] = connection2
			//endorsement_org1_group[1] = connection1
			err = chaincode.Approve(*org2MSP, "mychannel", "", "", "basic", "1.0", "", "", 1, nil, false, nil, endorsement_org2_group, connection4)
			Expect(err).NotTo(HaveOccurred())
			// commit from org1
			time.Sleep(time.Duration(15) * time.Second)
			connection5, err := basic.CreateBroadcastClient(context.Background(), orderer_node, logger)
			err = chaincode.Commit("mychannel", "", "", "basic", "1.0", "", "", 1, nil, false, nil, *org1MSP, endorsement_org1_group, connection5)
			Expect(err).NotTo(HaveOccurred())
			// commit from org2
			time.Sleep(time.Duration(15) * time.Second)
			connection6, err := basic.CreateBroadcastClient(context.Background(), orderer_node, logger)
			err = chaincode.Commit("mychannel", "", "", "basic", "1.0", "", "", 1, nil, false, nil, *org2MSP, endorsement_org2_group, connection6)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
