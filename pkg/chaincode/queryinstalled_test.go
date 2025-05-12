/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode_test

import (
	"context"
	"errors"

	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func AssertMarshal(m proto.Message) []byte {
	result, err := proto.Marshal(m)
	Expect(err).NotTo(HaveOccurred())
	return result
}

// AssertProtoEqual ensures an expected protobuf message matches an actual message
func AssertProtoEqual(expected proto.Message, actual proto.Message) {
	Expect(proto.Equal(expected, actual)).To(BeTrue(), "Expected %v, got %v", expected, actual)
}

var _ = Describe("QueryInstalled", func() {
	It("Endorser client called with supplied context", func(specCtx SpecContext) {
		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) error {
				proto.Merge(out, NewSuccessfulProposalResponse(nil))
				return ctx.Err()
			})

		mockSigner := NewMockSigner(controller, "", nil, nil)
		peer := chaincode.NewPeer(mockConnection, mockSigner)

		ctx, cancel := context.WithCancel(specCtx)
		cancel()

		_, err := peer.QueryInstalled(ctx)

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
		peer := chaincode.NewPeer(mockConnection, mockSigner)

		_, err := peer.QueryInstalled(specCtx)

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
				proto.Merge(out, NewErrorProposalResponse(expectedStatus, expectedMessage))
			})

		mockSigner := NewMockSigner(controller, "", nil, nil)
		peer := chaincode.NewPeer(mockConnection, mockSigner)

		_, err := peer.QueryInstalled(specCtx)

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
				proto.Merge(out, NewSuccessfulProposalResponse(nil))
			}).
			Times(1)

		mockSigner := NewMockSigner(controller, "", nil, expected)
		peer := chaincode.NewPeer(mockConnection, mockSigner)

		_, err := peer.QueryInstalled(specCtx)
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
				proto.Merge(out, NewSuccessfulProposalResponse(nil))
			}).
			Times(1)

		mockSigner := NewMockSigner(controller, expected.GetMspid(), expected.GetIdBytes(), nil)
		peer := chaincode.NewPeer(mockConnection, mockSigner)

		_, err := peer.QueryInstalled(specCtx)
		Expect(err).NotTo(HaveOccurred())

		signatureHeader := AssertUnmarshalSignatureHeader(signedProposal)

		actual := &msp.SerializedIdentity{}
		AssertUnmarshal(signatureHeader.GetCreator(), actual)

		AssertProtoEqual(expected, actual)
	})

	It("Installed chaincodes returned on successful proposal response", func(specCtx SpecContext) {
		expected := &lifecycle.QueryInstalledChaincodesResult{
			InstalledChaincodes: []*lifecycle.QueryInstalledChaincodesResult_InstalledChaincode{
				{
					PackageId: "PACKAGE_ID",
					Label:     "LABEL",
				},
			},
		}

		response := NewSuccessfulProposalResponse(AssertMarshal(expected))

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				proto.Merge(out, response)
			})

		mockSigner := NewMockSigner(controller, "", nil, nil)
		peer := chaincode.NewPeer(mockConnection, mockSigner)

		actual, err := peer.QueryInstalled(specCtx)
		Expect(err).NotTo(HaveOccurred())

		AssertProtoEqual(expected, actual)
	})
})
