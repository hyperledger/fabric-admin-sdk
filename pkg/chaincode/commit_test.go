/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"errors"
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o endorserclient_mock_test.go github.com/hyperledger/fabric-protos-go-apiv2/peer.EndorserClient
//counterfeiter:generate -o broadcastclient_mock_test.go github.com/hyperledger/fabric-protos-go-apiv2/orderer.AtomicBroadcast_BroadcastClient

var _ = Describe("Commit", func() {

	var chaincodeDefinition Definition
	var endorsementClients []peer.EndorserClient
	BeforeEach(func() {
		chaincodeDefinition = Definition{
			Name:        "CHAINCODE",
			Version:     "1.0",
			Sequence:    1,
			ChannelName: "CHANNEL",
		}
	})

	Context("CreateCommitProposal", func() {
		It("Should work for function CreateCommitProposal", func() {
			controller := gomock.NewController(GinkgoT())
			defer controller.Finish()

			mockSigner := NewMockSigner(controller, "", nil, nil)
			_, err := CreateCommitProposal(chaincodeDefinition, mockSigner)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("when the channel name is not provided", func() {
			errorData := Definition{
				Name:        "CHAINCODE",
				Version:     "1.0",
				Sequence:    1,
				ChannelName: "",
			}
			_, err := CreateCommitProposal(errorData, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("when the chaincode name is not provided", func() {
			errorData := Definition{
				Name:        "",
				Version:     "1.0",
				Sequence:    1,
				ChannelName: "CHANNEL",
			}
			_, err := CreateCommitProposal(errorData, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("when the chaincode version is not provided", func() {
			errorData := Definition{
				Name:        "CHAINCODE",
				Version:     "",
				Sequence:    1,
				ChannelName: "CHANNEL",
			}
			_, err := CreateCommitProposal(errorData, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("when the sequence is not provided", func() {
			errorData := Definition{
				Name:        "CHAINCODE",
				Version:     "1.0",
				Sequence:    0,
				ChannelName: "CHANNEL",
			}
			_, err := CreateCommitProposal(errorData, nil)
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("Commit", func() {
		It("Should handle Sign error when Commit", func() {
			controller := gomock.NewController(GinkgoT())
			defer controller.Finish()

			mockSigner := NewMockSigningIdentity(controller)
			mockSigner.EXPECT().MspID().Return("").AnyTimes()
			mockSigner.EXPECT().Credentials().Return(nil).AnyTimes()
			mockSigner.EXPECT().Sign(gomock.Any()).Return(nil, fmt.Errorf("tea"))

			err := Commit(chaincodeDefinition, mockSigner, endorsementClients, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("Should handle Endorsement error when Commit", func() {
			controller := gomock.NewController(GinkgoT())
			defer controller.Finish()

			mockSigner := NewMockSigner(controller, "", nil, nil)
			mockEndorserClient := &FakeEndorserClient{}
			endorsementClients = make([]peer.EndorserClient, 0)
			endorsementClients = append(endorsementClients, mockEndorserClient)
			mockEndorserClient.ProcessProposalReturns(nil, errors.New("latte"))
			err := Commit(chaincodeDefinition, mockSigner, endorsementClients, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("Should handle BroadcastClient error when Commit", func() {
			controller := gomock.NewController(GinkgoT())
			defer controller.Finish()

			mockSigner := NewMockSigner(controller, "", nil, nil)
			mockEndorserClient := &FakeEndorserClient{}
			endorsementClients = make([]peer.EndorserClient, 0)
			endorsementClients = append(endorsementClients, mockEndorserClient)
			mockProposalResponse := &peer.ProposalResponse{
				Response: &peer.Response{
					Status: 200,
				},
				Endorsement: &peer.Endorsement{},
			}
			mockEndorserClient.ProcessProposalReturns(mockProposalResponse, nil)
			mockBroadcastClient := &FakeAtomicBroadcast_BroadcastClient{}
			mockBroadcastClient.SendReturns(errors.New("coffee"))
			err := Commit(chaincodeDefinition, mockSigner, endorsementClients, mockBroadcastClient)
			Expect(err).Should(HaveOccurred())
		})

		It("Should works when Commit", func() {
			controller := gomock.NewController(GinkgoT())
			defer controller.Finish()

			mockSigner := NewMockSigner(controller, "", nil, nil)
			mockEndorserClient := &FakeEndorserClient{}
			endorsementClients = make([]peer.EndorserClient, 0)
			endorsementClients = append(endorsementClients, mockEndorserClient)
			mockProposalResponse := &peer.ProposalResponse{
				Response: &peer.Response{
					Status: 200,
				},
				Endorsement: &peer.Endorsement{},
			}
			mockEndorserClient.ProcessProposalReturns(mockProposalResponse, nil)
			mockBroadcastClient := &FakeAtomicBroadcast_BroadcastClient{}
			err := Commit(chaincodeDefinition, mockSigner, endorsementClients, mockBroadcastClient)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
