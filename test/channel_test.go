package test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"

	"github.com/hyperledger/fabric-admin-sdk/pkg/channel"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/network"

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
		It("should work", func(specCtx SpecContext) {
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
			peerConnection, err := network.DialConnection(peer1)
			Expect(err).NotTo(HaveOccurred())

			cert, err := identity.ReadCertificate(SignCert)
			Expect(err).NotTo(HaveOccurred())

			priv, err := identity.ReadPrivateKey(PrivKeyPath)
			Expect(err).NotTo(HaveOccurred())

			id, err := identity.NewPrivateKeySigningIdentity(MSPID, cert, priv)
			Expect(err).NotTo(HaveOccurred())

			configBlock, err := channel.GetConfigBlock(specCtx, peerConnection, id, channelID)
			Expect(err).NotTo(HaveOccurred())
			Expect(configBlock).ShouldNot(BeNil())
			fmt.Println("config block", configBlock)
		})
	})

	Context("get block chain info", func() {
		It("should work", func(specCtx SpecContext) {
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
			peerConnection, err := network.DialConnection(peer1)
			Expect(err).NotTo(HaveOccurred())

			cert, err := identity.ReadCertificate(SignCert)
			Expect(err).NotTo(HaveOccurred())

			priv, err := identity.ReadPrivateKey(PrivKeyPath)
			Expect(err).NotTo(HaveOccurred())

			id, err := identity.NewPrivateKeySigningIdentity(MSPID, cert, priv)
			Expect(err).NotTo(HaveOccurred())

			blockChainInfo, err := channel.GetBlockChainInfo(specCtx, peerConnection, id, channelID)
			Expect(err).NotTo(HaveOccurred())
			Expect(blockChainInfo).ShouldNot(BeNil())
			fmt.Println("blockchain info", blockChainInfo)
		})
	})

	Context("update channel config", func() {
		It("should work", func() {
			_, err := os.Stat("../fabric-samples/test-network")
			if err != nil {
				Skip("skip for unit test")
			}
			var channelID = "mychannel"
			// Orderer
			var OrdererAddr = "localhost:7050"
			var OrdererTLSCACert = "../fabric-samples/test-network/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"

			// Peer1
			var MSPID = "Org1MSP"
			var PrivKeyPath = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk"
			var SignCertPath = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem"

			// Peer2
			var MSPID2 = "Org2MSP"
			var PrivKeyPath2 = "../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/keystore/priv_sk"
			var SignCertPath2 = "../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/signcerts/Admin@org2.example.com-cert.pem"

			cert, err := identity.ReadCertificate(SignCertPath)
			Expect(err).NotTo(HaveOccurred())

			priv, err := identity.ReadPrivateKey(PrivKeyPath)
			Expect(err).NotTo(HaveOccurred())

			signer, err := identity.NewPrivateKeySigningIdentity(MSPID, cert, priv)
			Expect(err).NotTo(HaveOccurred())

			cert2, err := identity.ReadCertificate(SignCertPath2)
			Expect(err).NotTo(HaveOccurred())

			priv2, err := identity.ReadPrivateKey(PrivKeyPath2)
			Expect(err).NotTo(HaveOccurred())

			signer2, err := identity.NewPrivateKeySigningIdentity(MSPID2, cert2, priv2)
			Expect(err).NotTo(HaveOccurred())

			// get update config file, see https://hyperledger-fabric.readthedocs.io/en/release-2.4/channel_update_tutorial.html#add-the-org3-crypto-material
			updateEnvelope, err := os.ReadFile("./org3_update_in_envelope.pb")
			Expect(err).NotTo(HaveOccurred())
			envelope, err := protoutil.UnmarshalEnvelope(updateEnvelope)
			Expect(err).NotTo(HaveOccurred())

			// Peer1 sign
			envelope, err = SignConfigTx(channelID, envelope, signer)
			Expect(err).NotTo(HaveOccurred())

			// Peer2 sign
			envelope, err = SignConfigTx(channelID, envelope, signer2)
			Expect(err).NotTo(HaveOccurred())

			ordererNode := network.Node{
				Addr:      OrdererAddr,
				TLSCACert: OrdererTLSCACert,
			}
			err = ordererNode.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			ordererConnection, err := network.DialConnection(ordererNode)
			Expect(err).NotTo(HaveOccurred())
			ordererClient, err := orderer.NewAtomicBroadcastClient(ordererConnection).Broadcast(context.Background())
			Expect(err).NotTo(HaveOccurred())
			defer func(ordererClient orderer.AtomicBroadcast_BroadcastClient) {
				_ = ordererClient.CloseSend()
			}(ordererClient)
			err = ordererClient.Send(envelope)
			Expect(err).NotTo(HaveOccurred())
			response, err := ordererClient.Recv()
			Expect(err).NotTo(HaveOccurred())
			log.Println("response: ", response.String())
		})
	})
})
