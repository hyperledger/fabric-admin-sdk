/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
)

var _ = Describe("QueryCommitted", func() {
	It("Endorser client called with supplied context", func(specCtx SpecContext) {
		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) error {
				CopyProto(NewProposalResponse(common.Status_SUCCESS, ""), out)
				return ctx.Err()
			})

		mockSigner := NewMockSigner(controller, "", nil, nil)

		ctx, cancel := context.WithCancel(specCtx)
		cancel()

		_, err := QueryCommitted(ctx, mockConnection, mockSigner, "")

		Expect(err).To(MatchError(context.Canceled))
	})

	It("Endorser client errors returned", func(specCtx SpecContext) {
		expectedErr := errors.New("EXPECTED_ERROR")

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr)

		mockSigner := NewMockSigner(controller, "", nil, nil)

		_, err := QueryCommitted(specCtx, mockConnection, mockSigner, "mockchannel")

		Expect(err).To(MatchError(expectedErr))
	})

	It("Unsuccessful proposal response gives error", func(specCtx SpecContext) {
		expectedStatus := common.Status_BAD_REQUEST
		expectedMessage := "EXPECTED_ERROR"

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				CopyProto(NewProposalResponse(expectedStatus, expectedMessage), out)
			})

		mockSigner := NewMockSigner(controller, "", nil, nil)

		_, err := QueryCommitted(specCtx, mockConnection, mockSigner, "mockchannel")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(And(
			ContainSubstring("%d", expectedStatus),
			ContainSubstring(expectedStatus.String()),
			ContainSubstring(expectedMessage),
		))
	})

	It("Uses signer", func(specCtx SpecContext) {
		expected := []byte("SIGNATURE")

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		var signedProposal *peer.SignedProposal
		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				signedProposal = in
				CopyProto(NewProposalResponse(common.Status_SUCCESS, ""), out)
			}).
			Times(1)

		mockSigner := NewMockSigner(controller, "", nil, expected)

		_, err := QueryCommitted(specCtx, mockConnection, mockSigner, "mockchannel")
		Expect(err).NotTo(HaveOccurred())

		actual := signedProposal.GetSignature()
		Expect(actual).To(BeEquivalentTo(expected))
	})

	It("Proposal includes creator", func(specCtx SpecContext) {
		expected := &msp.SerializedIdentity{
			Mspid:   "MSP_ID",
			IdBytes: []byte("CREDENTIALS"),
		}

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		var signedProposal *peer.SignedProposal
		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				signedProposal = in
				CopyProto(NewProposalResponse(common.Status_SUCCESS, ""), out)
			}).
			Times(1)

		mockSigner := NewMockSigner(controller, expected.Mspid, expected.IdBytes, nil)

		_, err := QueryCommitted(specCtx, mockConnection, mockSigner, "mockchannel")
		Expect(err).NotTo(HaveOccurred())

		signatureHeader := AssertUnmarshalSignatureHeader(signedProposal)

		actual := &msp.SerializedIdentity{}
		AssertUnmarshal(signatureHeader.GetCreator(), actual)

		AssertProtoEqual(expected, actual)
	})

	It("Committed chaincodes returned on successful proposal response", func(specCtx SpecContext) {
		expected := &lifecycle.QueryChaincodeDefinitionsResult{
			ChaincodeDefinitions: []*lifecycle.QueryChaincodeDefinitionsResult_ChaincodeDefinition{
				{
					Name:                "CHAINCODE_NAME",
					Version:             "CHAINCODE_VERSION",
					Sequence:            1,
					EndorsementPlugin:   "ENDORSEMENT_PLUGIN",
					ValidationPlugin:    "VALIDATION_PLUGIN",
					ValidationParameter: []byte("VALIDATION_PARAMETER"),
					InitRequired:        true,
					Collections:         &peer.CollectionConfigPackage{},
				},
			},
		}

		response := NewProposalResponse(common.Status_SUCCESS, "")
		response.Response.Payload = AssertMarshal(expected)

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				CopyProto(response, out)
			})

		mockSigner := NewMockSigner(controller, "", nil, nil)

		actual, err := QueryCommitted(specCtx, mockConnection, mockSigner, "mockchannel")
		Expect(err).NotTo(HaveOccurred())

		AssertProtoEqual(expected, actual)
	})
})
