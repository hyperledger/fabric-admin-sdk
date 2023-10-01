package test

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-admin-sdk/pkg/discovery"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/network"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("discovery", func() {
	Context("query peer membership", func() {
		It("should work", func(specCtx SpecContext) {
			var peerAddr = "localhost:7051"
			var TLSCACert = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
			var PrivKeyPath = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk"
			var SignCert = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem"
			var MSPID = "Org1MSP"
			var channelID = "mychannel"

			peer0 := network.Node{
				Addr:      peerAddr,
				TLSCACert: TLSCACert,
			}
			err := peer0.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			peerConnection, err := network.DialConnection(peer0)
			Expect(err).NotTo(HaveOccurred())

			cert, _, err := identity.GetCertificate(SignCert)
			Expect(err).NotTo(HaveOccurred())

			priv, err := identity.GetecdsaPrivateKey(PrivKeyPath)
			Expect(err).NotTo(HaveOccurred())

			id, err := identity.NewPrivateKeySigningIdentity(MSPID, cert, priv)
			Expect(err).NotTo(HaveOccurred())

			ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
			defer cancel()
			peerMembershipResult, err := discovery.PeerMembershipQuery(ctx, peerConnection, id, channelID, nil)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("peerMembershipResult", peerMembershipResult)
		})
	})
})
