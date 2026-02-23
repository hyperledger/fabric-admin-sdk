package test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	"github.com/hyperledger/fabric-admin-sdk/pkg/channel"
	"github.com/hyperledger/fabric-admin-sdk/pkg/discovery"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/network"
	"github.com/hyperledger/fabric-admin-sdk/pkg/snapshot"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	gatewaypb "github.com/hyperledger/fabric-protos-go-apiv2/gateway"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/status"
)

const (
	org1PeerAddress = "localhost:7051"
	org2PeerAddress = "localhost:9051"
	channelName     = "mychannel"
	org1MspID       = "Org1MSP"
	org2MspID       = "Org2MSP"

	snapshotBlockNumber uint64 = 10
)

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

	fmt.Printf("Received error type %T: %s\n", err, err)

	var endorseErr *client.EndorseError
	var submitErr *client.SubmitError
	var commitStatusErr *client.CommitStatusError
	var commitErr *client.CommitError

	if errors.As(err, &endorseErr) {
		fmt.Printf("Endorse error for transaction %s with gRPC status %v: %s\n", endorseErr.TransactionID, status.Code(err), endorseErr)
	} else if errors.As(err, &submitErr) {
		fmt.Printf("Submit error for transaction %s with gRPC status %v: %s\n", submitErr.TransactionID, status.Code(err), submitErr)
	} else if errors.As(err, &commitStatusErr) {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Timeout waiting for transaction %s commit status: %s", commitStatusErr.TransactionID, commitStatusErr)
		} else {
			fmt.Printf("Error obtaining commit status for transaction %s with gRPC status %v: %s\n", commitStatusErr.TransactionID, status.Code(commitStatusErr), commitStatusErr)
		}
	} else if errors.As(err, &commitErr) {
		fmt.Printf("Transaction %s failed to commit with status %d: %s\n", commitErr.TransactionID, int32(commitErr.Code), commitErr)
	}

	statusErr := status.Convert(err)

	details := statusErr.Details()
	if len(details) > 0 {
		fmt.Println("Error Details:")

		for _, detail := range details {
			switch detail := detail.(type) {
			case *gatewaypb.ErrorDetail:
				fmt.Printf("- address: %s, mspId: %s, message: %s\n", detail.GetAddress(), detail.GetMspId(), detail.GetMessage())
			}
		}
	}
}

var _ = Describe("e2e", func() {
	Context("the e2e test with test network", func() {
		It("should work", func(specCtx SpecContext) {
			_, err := os.Stat("../fabric-samples/test-network")
			if err != nil {
				Skip("skip for unit test")
			}
			TLSCACert := "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
			PrivKeyPath := "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/priv_sk"
			SignCert := "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem"

			peer1 := network.Node{
				Addr:      org1PeerAddress,
				TLSCACert: TLSCACert,
			}
			err = peer1.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			peer1Connection, err := network.DialConnection(peer1)
			Expect(err).NotTo(HaveOccurred())

			cert, err := identity.ReadCertificate(SignCert)
			Expect(err).NotTo(HaveOccurred())

			priv, err := identity.ReadPrivateKey(PrivKeyPath)
			Expect(err).NotTo(HaveOccurred())

			org1MSP, err := identity.NewPrivateKeySigningIdentity(org1MspID, cert, priv)
			Expect(err).NotTo(HaveOccurred())

			TLSCACert = "../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
			PrivKeyPath = "../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/keystore/priv_sk"
			SignCert = "../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/signcerts/Admin@org2.example.com-cert.pem"

			peer2 := network.Node{
				Addr:      org2PeerAddress,
				TLSCACert: TLSCACert,
			}
			err = peer2.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			peer2Connection, err := network.DialConnection(peer2)
			Expect(err).NotTo(HaveOccurred())

			cert2, err := identity.ReadCertificate(SignCert)
			Expect(err).NotTo(HaveOccurred())

			priv2, err := identity.ReadPrivateKey(PrivKeyPath)
			Expect(err).NotTo(HaveOccurred())

			org2MSP, err := identity.NewPrivateKeySigningIdentity(org2MspID, cert2, priv2)
			Expect(err).NotTo(HaveOccurred())

			//genesis block
			createChannel, ok := os.LookupEnv("CREATE_CHANNEL")
			if createChannel == "create_channel" && ok {
				IsBFT, check := os.LookupEnv("CONSENSUS")
				var block *cb.Block
				if IsBFT == "BFT" && check {
					block, err = channel.CreateGenesisBlock("ChannelUsingBFT", "./bft", channelName)
					Expect(err).NotTo(HaveOccurred())
				} else {
					block, err = channel.CreateGenesisBlock("ChannelUsingRaft", "./", channelName)
					Expect(err).NotTo(HaveOccurred())
				}
				Expect(block).ToNot(BeNil())

				//create channel
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
				resp, err := channel.CreateChannel(osnURL, block, caCertPool, tlsClientCert)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).Should(Equal(http.StatusCreated))
				if IsBFT == "BFT" && check {
					osnURLs := []string{"https://localhost:7055", "https://localhost:7057", "https://localhost:7059"}
					clientCerts := []string{
						"../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer2.example.com/tls/server.crt",
						"../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer3.example.com/tls/server.crt",
						"../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer4.example.com/tls/server.crt",
					}
					clientKeys := []string{
						"../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer2.example.com/tls/server.key",
						"../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer3.example.com/tls/server.key",
						"../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer4.example.com/tls/server.key",
					}
					for i := 0; i < 3; i++ {
						osnURL = osnURLs[i]
						caFile = "../fabric-samples/test-network/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
						clientCert = clientCerts[i]
						clientKey = clientKeys[i]
						caCertPool := x509.NewCertPool()
						caFilePEM, err := os.ReadFile(caFile)
						caCertPool.AppendCertsFromPEM(caFilePEM)
						Expect(err).NotTo(HaveOccurred())
						tlsClientCert, err := tls.LoadX509KeyPair(clientCert, clientKey)
						Expect(err).NotTo(HaveOccurred())
						resp, err := channel.CreateChannel(osnURL, block, caCertPool, tlsClientCert)
						Expect(err).NotTo(HaveOccurred())
						Expect(resp.StatusCode).Should(Equal(http.StatusCreated))
					}
				}
				//osnURL
				order := network.Node{
					Addr:      "localhost:7050",
					TLSCACert: clientCert,
				}
				err = order.LoadConfig()
				Expect(err).NotTo(HaveOccurred())
				ordererConnection, err := network.DialConnection(order)
				Expect(err).NotTo(HaveOccurred())
				ctx, cancel := context.WithTimeout(specCtx, 2*time.Minute)
				defer cancel()
				ordererBlock, err := channel.GetConfigBlockFromOrderer(ctx, ordererConnection, org1MSP, channelName, tlsClientCert)
				Expect(err).NotTo(HaveOccurred())
				Expect(ordererBlock).NotTo(BeNil())

				//join peer1
				err = channel.JoinChannel(specCtx, peer1Connection, org1MSP, block)
				Expect(err).NotTo(HaveOccurred())

				//join peer2
				err = channel.JoinChannel(specCtx, peer2Connection, org2MSP, block)
				Expect(err).NotTo(HaveOccurred())
			}
			// check peer join channel
			peerChannelInfo, err := channel.ListChannelOnPeer(specCtx, peer1Connection, org1MSP)
			Expect(err).NotTo(HaveOccurred())
			Expect(peerChannelInfo[0].GetChannelId()).To(Equal(channelName))
			// package chaincode as CCAAS
			dummyConnection := chaincode.Connection{
				Address:     "{{.peername}}_basic:9999",
				DialTimeout: "10s",
				TLSRequired: false,
			}
			dummyMeta := chaincode.Metadata{
				Type:  "ccaas",
				Label: "basic_1.0",
			}
			packageFileName := "basic-asset.tar.gz"
			err = chaincode.PackageCCAAS(dummyConnection, dummyMeta, tmpDir, packageFileName)
			Expect(err).NotTo(HaveOccurred())

			packageFilePath := path.Join(tmpDir, packageFileName)
			packageReader, err := os.Open(packageFilePath)
			Expect(err).NotTo(HaveOccurred(), "open chaincode package file")
			chaincodePackage, err := io.ReadAll(packageReader)
			Expect(err).NotTo(HaveOccurred(), "read chaincode package")

			packageID, err := chaincode.PackageID(bytes.NewReader(chaincodePackage))
			Expect(err).NotTo(HaveOccurred(), "get chaincode package ID")
			fmt.Println(packageID)

			networkPeers := []*chaincode.Peer{
				chaincode.NewPeer(peer1Connection, org1MSP),
				chaincode.NewPeer(peer2Connection, org2MSP),
			}

			org1Gateway := chaincode.NewGateway(peer1Connection, org1MSP)
			org2Gateway := chaincode.NewGateway(peer2Connection, org2MSP)
			allOrgGateways := []*chaincode.Gateway{org1Gateway, org2Gateway}

			// Install chaincode on each peer
			runParallel(networkPeers, func(peer *chaincode.Peer) {
				ctx, cancel := context.WithTimeout(specCtx, 2*time.Minute)
				defer cancel()
				result, err := peer.Install(ctx, bytes.NewReader(chaincodePackage))
				printGrpcError(err)
				Expect(err).NotTo(HaveOccurred(), "chaincode install")
				Expect(result.GetPackageId()).To(Equal(packageID), "install chaincode package ID")
				Expect(result.GetLabel()).To(Equal(dummyMeta.Label), "install chaincode label")
			})

			// Query installed chaincode on each peer
			runParallel(networkPeers, func(peer *chaincode.Peer) {
				ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
				defer cancel()
				result, err := peer.QueryInstalled(ctx)
				printGrpcError(err)
				Expect(err).NotTo(HaveOccurred(), "query installed chaincode")
				installedChaincodes := result.GetInstalledChaincodes()
				Expect(installedChaincodes).To(HaveLen(1), "number of installed chaincodes")
				Expect(installedChaincodes[0].GetPackageId()).To(Equal(packageID), "installed chaincode package ID")
				Expect(installedChaincodes[0].GetLabel()).To(Equal(dummyMeta.Label), "installed chaincode label")
			})

			// Get installed chaincode package from each peer
			runParallel(networkPeers, func(peer *chaincode.Peer) {
				ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
				defer cancel()
				result, err := peer.GetInstalled(ctx, packageID)
				printGrpcError(err)
				Expect(err).NotTo(HaveOccurred(), "get installed chaincode package")
				Expect(result).NotTo(BeEmpty())
			})

			time.Sleep(time.Duration(20) * time.Second)
			PolicyStr := "AND ('Org1MSP.peer','Org2MSP.peer')"
			applicationPolicy, err := chaincode.NewApplicationPolicy(PolicyStr, "")
			Expect(err).NotTo(HaveOccurred())
			chaincodeDef := &chaincode.Definition{
				ChannelName:       channelName,
				PackageID:         "",
				Name:              "basic",
				Version:           "1.0",
				EndorsementPlugin: "",
				ValidationPlugin:  "",
				Sequence:          1,
				ApplicationPolicy: applicationPolicy,
				InitRequired:      false,
				Collections:       nil,
			}
			Expect(err).NotTo(HaveOccurred())
			// Approve chaincode for each org
			runParallel(allOrgGateways, func(gateway *chaincode.Gateway) {
				ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
				defer cancel()
				err := gateway.Approve(ctx, chaincodeDef)
				printGrpcError(err)
				Expect(err).NotTo(HaveOccurred(), "approve chaincode for org %s", gateway.ClientIdentity().MspID())
			})

			// Query approved chaincode for each org
			runParallel(allOrgGateways, func(gateway *chaincode.Gateway) {
				ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
				defer cancel()
				result, err := gateway.QueryApproved(ctx, channelName, chaincodeDef.Name, chaincodeDef.Sequence)
				printGrpcError(err)
				Expect(err).NotTo(HaveOccurred(), "query approved chaincode for org %s", gateway.ClientIdentity().MspID())
				Expect(result.GetVersion()).To(Equal(chaincodeDef.Version))
			})

			// Check chaincode commit readiness
			readinessCtx, readinessCancel := context.WithTimeout(specCtx, 30*time.Second)
			defer readinessCancel()
			readinessResult, err := org1Gateway.CheckCommitReadiness(readinessCtx, chaincodeDef)
			printGrpcError(err)
			Expect(err).NotTo(HaveOccurred(), "check commit readiness")
			Expect(readinessResult.GetApprovals()[org1MspID]).To(BeTrue())
			Expect(readinessResult.GetApprovals()[org2MspID]).To(BeTrue())

			time.Sleep(time.Duration(20) * time.Second)

			// Commit chaincode
			commitCtx, commitCancel := context.WithTimeout(specCtx, 30*time.Second)
			defer commitCancel()
			err = org1Gateway.Commit(commitCtx, chaincodeDef)
			printGrpcError(err)
			Expect(err).NotTo(HaveOccurred(), "commit chaincode")

			// Query all committed chaincode
			committedCtx, committedCancel := context.WithTimeout(specCtx, 30*time.Second)
			defer committedCancel()
			committedResult, err := org1Gateway.QueryCommitted(committedCtx, channelName)
			printGrpcError(err)
			Expect(err).NotTo(HaveOccurred(), "query all committed chaincodes")
			committedChaincodes := committedResult.GetChaincodeDefinitions()
			Expect(committedChaincodes).To(HaveLen(1), "number of committed chaincodes")
			Expect(committedChaincodes[0].GetName()).To(Equal("basic"), "committed chaincode name")
			Expect(committedChaincodes[0].GetSequence()).To(Equal(int64(1)), "committed chaincode sequence")

			// Query named committed chaincode
			committedWithNameCtx, committedWithNameCancel := context.WithTimeout(specCtx, 30*time.Second)
			defer committedWithNameCancel()
			committedWithNameResult, err := org1Gateway.QueryCommittedWithName(committedWithNameCtx, channelName, chaincodeDef.Name)
			printGrpcError(err)
			Expect(err).NotTo(HaveOccurred(), "query committed chaincode with name")
			Expect(readinessResult.GetApprovals()[org1MspID]).To(BeTrue())
			Expect(readinessResult.GetApprovals()[org2MspID]).To(BeTrue())
			Expect(committedWithNameResult.GetSequence()).To(Equal(chaincodeDef.Sequence), "committed chaincode sequence")

			f, _ := os.Create("PackageID")
			_, err = io.WriteString(f, packageID)
			Expect(err).NotTo(HaveOccurred())

			// check discovery as query peer membership
			discoveryPeer := discovery.NewPeer(peer1Connection, org1MSP)
			peerMembershipCtx, cancel := context.WithTimeout(specCtx, 30*time.Second)
			defer cancel()
			peerMembershipResult, err := discoveryPeer.PeerMembershipQuery(peerMembershipCtx, channelName, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(peerMembershipResult.GetPeersByOrg()[org1MspID].GetPeers()).To(HaveLen(1))
			Expect(peerMembershipResult.GetPeersByOrg()[org2MspID].GetPeers()).To(HaveLen(1))

			// check snapshot
			snapshotPeer := snapshot.NewPeer(peer1Connection, org1MSP)
			snapshotCtx, snapshotCancel := context.WithTimeout(specCtx, 30*time.Second)
			defer snapshotCancel()

			// query snapshot when no snapshot request has been submitted
			queryPendingRes, err := snapshotPeer.QueryPending(snapshotCtx, channelName)
			Expect(err).NotTo(HaveOccurred())
			Expect(queryPendingRes.GetBlockNumbers()).To(BeEmpty(), "no pending snapshot requests before submission")

			// submit snapshot request
			err = snapshotPeer.SubmitRequest(snapshotCtx, channelName, snapshotBlockNumber)
			Expect(err).NotTo(HaveOccurred(), "submit snapshot request")

			// query pending snapshot request after submission
			queryPendingRes, err = snapshotPeer.QueryPending(snapshotCtx, channelName)
			Expect(err).NotTo(HaveOccurred())
			Expect(queryPendingRes.GetBlockNumbers()).To(ContainElement(snapshotBlockNumber), "pending snapshot request after submission")

			// cancel snapshot request
			err = snapshotPeer.CancelRequest(snapshotCtx, channelName, snapshotBlockNumber)
			Expect(err).NotTo(HaveOccurred(), "cancel snapshot request after submission")

			// query pending snapshot request after cancellation
			queryPendingRes, err = snapshotPeer.QueryPending(snapshotCtx, channelName)
			Expect(err).NotTo(HaveOccurred())
			Expect(queryPendingRes.GetBlockNumbers()).To(BeEmpty(), "no pending snapshot requests after cancellation")
		})
	})
})
