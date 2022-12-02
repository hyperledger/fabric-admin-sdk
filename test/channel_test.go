package test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-admin-sdk/internal/network"
	"github.com/hyperledger/fabric-admin-sdk/pkg/channel"
	"github.com/hyperledger/fabric-admin-sdk/pkg/tools"

	npb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("channel", func() {
	Context("get channel list", func() {
		It("should work", func() {
			_, err := os.Stat("../fabric-samples/test-network")
			if err != nil {
				Skip("skip for unit test")
			}
			var caFile, clientCert, clientKey, osnURL string
			osnURL = "https://localhost:7053"
			caFile = "../fabric-samples/test-network/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
			clientCert = "../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt"
			clientKey = "../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key"
			caCertPool := x509.NewCertPool()
			caFilePEM, err := os.ReadFile(caFile)
			caCertPool.AppendCertsFromPEM(caFilePEM)
			Expect(err).NotTo(HaveOccurred())
			tlsClientCert, err := tls.LoadX509KeyPair(clientCert, clientKey)
			Expect(err).NotTo(HaveOccurred())
			channels, err := channel.ListChannel(osnURL, caCertPool, tlsClientCert)
			Expect(err).NotTo(HaveOccurred())
			for _, v := range channels.Channels {
				fmt.Println("channel name: ", v.Name)
			}
		})
	})

	Context("get config block", func() {
		It("should work", func() {
			_, err := os.Stat("../fabric-samples/test-network")
			if err != nil {
				Skip("skip for unit test")
			}
			var peerAddr = "localhost:7051"
			var TLSCACert = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
			var PrivKeyPath = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk"
			var SignCert = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem"
			var MSPID = "Org1MSP"
			var channelID = "mychannel"

			peer1 := network.Node{
				Addr:      peerAddr,
				TLSCACert: TLSCACert,
			}
			err = peer1.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			n_conn1, err := network.DialConnection(peer1)
			Expect(err).NotTo(HaveOccurred())

			connection := npb.NewEndorserClient(n_conn1)

			id, err := tools.CreateSigner(PrivKeyPath, SignCert, MSPID)
			Expect(err).NotTo(HaveOccurred())

			configBlock, err := channel.GetConfigBlock(id, channelID, connection)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("config block", configBlock)
		})
	})

	Context("get block chain info", func() {
		It("should work", func() {
			_, err := os.Stat("../fabric-samples/test-network")
			if err != nil {
				Skip("skip for unit test")
			}
			var peerAddr = "localhost:7051"
			var TLSCACert = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
			var PrivKeyPath = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk"
			var SignCert = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem"
			var MSPID = "Org1MSP"
			var channelID = "mychannel"

			peer1 := network.Node{
				Addr:      peerAddr,
				TLSCACert: TLSCACert,
			}
			err = peer1.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			n_conn1, err := network.DialConnection(peer1)
			Expect(err).NotTo(HaveOccurred())
			connection := npb.NewEndorserClient(n_conn1)

			id, err := tools.CreateSigner(PrivKeyPath, SignCert, MSPID)
			Expect(err).NotTo(HaveOccurred())

			blockChainInfo, err := channel.GetBlockChainInfo(id, channelID, connection)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("blockchain info", blockChainInfo)
		})
	})
})
