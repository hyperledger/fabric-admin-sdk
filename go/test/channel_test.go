package test

import (
	"crypto/tls"
	"crypto/x509"
	"fabric-admin-sdk/pkg/channel"
	"fmt"
	"os"

	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("channel", func() {
	Context("get channel list", func() {
		It("should work", func() {
			_, err := os.Stat("../" + testNetworkPath)
			if err != nil {
				Skip("skip for unit test")
			}
			var caFile, clientCert, clientKey, osnURL string
			osnURL = "https://orderer.example.com:7053"
			caFile = testNetworkPath + "/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
			clientCert = testNetworkPath + "/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt"
			clientKey = testNetworkPath + "/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key"
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
			_, err := os.Stat("../" + testNetworkPath)
			if err != nil {
				Skip("skip for unit test")
			}
			var peerAddr = "peer0.org1.example.com:7051"
			var TLSCACert = testNetworkPath + "/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
			var PrivKeyPath = testNetworkPath + "/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk"
			var SignCert = testNetworkPath + "/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem"
			var MSPID = "Org1MSP"
			var channelID = "mychannel"

			logger := log.New()
			peer1 := basic.Node{
				Addr:      peerAddr,
				TLSCACert: TLSCACert,
			}
			err = peer1.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			connection, err := basic.CreateEndorserClient(peer1, logger)
			Expect(err).NotTo(HaveOccurred())
			configBlock, err := channel.GetConfigBlock(SignCert, PrivKeyPath, MSPID, channelID, connection)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("config block", configBlock)
		})
	})

	Context("get block chain info", func() {
		It("should work", func() {
			_, err := os.Stat("../" + testNetworkPath)
			if err != nil {
				Skip("skip for unit test")
			}
			var peerAddr = "peer0.org1.example.com:7051"
			var TLSCACert = testNetworkPath + "/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
			var PrivKeyPath = testNetworkPath + "/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk"
			var SignCert = testNetworkPath + "/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem"
			var MSPID = "Org1MSP"
			var channelID = "mychannel"

			logger := log.New()
			peer1 := basic.Node{
				Addr:      peerAddr,
				TLSCACert: TLSCACert,
			}
			err = peer1.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			connection, err := basic.CreateEndorserClient(peer1, logger)
			Expect(err).NotTo(HaveOccurred())
			blockChainInfo, err := channel.GetBlockChainInfo(SignCert, PrivKeyPath, MSPID, channelID, connection)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("blockchain info", blockChainInfo)
		})
	})
})
