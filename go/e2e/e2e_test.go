package e2e_test

import (
	"crypto/tls"
	"crypto/x509"
	"fabric-admin-sdk/channel"
	"fabric-admin-sdk/tools"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("e2e", func() {
	Context("Create channel", func() {
		It("should create channel with test-network", func() {
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

			peer := basic.Node{
				Addr:      peer_addr,
				TLSCACert: TLSCACert,
			}

			err = peer.LoadConfig()
			Expect(err).NotTo(HaveOccurred())

			connection, err := basic.CreateEndorserClient(peer, logger)
			Expect(err).NotTo(HaveOccurred())

			err = channel.JoinChannel(
				block, PrivKeyPath, SignCert, MSPID, connection,
			)
			Expect(err).NotTo(HaveOccurred())

			//join peer2
			peer_addr = "localhost:9051"
			TLSCACert = "../../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
			PrivKeyPath = "../../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/keystore/priv_sk"
			SignCert = "../../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/signcerts/Admin@org2.example.com-cert.pem"
			MSPID = "Org2MSP"

			peer = basic.Node{
				Addr:      peer_addr,
				TLSCACert: TLSCACert,
			}

			err = peer.LoadConfig()
			Expect(err).NotTo(HaveOccurred())

			connection, err = basic.CreateEndorserClient(peer, logger)
			Expect(err).NotTo(HaveOccurred())

			err = channel.JoinChannel(
				block, PrivKeyPath, SignCert, MSPID, connection,
			)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
