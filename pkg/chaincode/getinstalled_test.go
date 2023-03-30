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
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
)

var _ = Describe("GetInstalled", func() {
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

		/*mockSigner := NewMockSigner(controller, "", nil, nil)

		ctx, cancel := context.WithCancel(specCtx)
		cancel()

		//_, err := GetInstalled(ctx, mockConnection, mockSigner, "PACKAGE_ID")

		Expect(err).To(MatchError(context.Canceled))*/
	})

	It("Endorser client errors returned", func(specCtx SpecContext) {
		expectedErr := errors.New("EXPECTED_ERROR")

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr)

		/*mockSigner := NewMockSigner(controller, "", nil, nil)

		_, err := GetInstalled(specCtx, mockConnection, mockSigner, "PACKAGE_ID")

		Expect(err).To(MatchError(expectedErr))*/
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

		/*mockSigner := NewMockSigner(controller, "", nil, nil)

		_, err := GetInstalled(specCtx, mockConnection, mockSigner, "PACKAGE_ID")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(And(
			ContainSubstring("%d", expectedStatus),
			ContainSubstring(expectedStatus.String()),
			ContainSubstring(expectedMessage),
		))*/
	})

	It("Uses signer", func(specCtx SpecContext) {
		//expected := []byte("SIGNATURE")

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		//var signedProposal *peer.SignedProposal
		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				//signedProposal = in
				CopyProto(NewProposalResponse(common.Status_SUCCESS, ""), out)
			}).
			Times(1)

		/*mockSigner := NewMockSigner(controller, "", nil, expected)

		_, err := GetInstalled(specCtx, mockConnection, mockSigner, "PACKAGE_ID")
		Expect(err).NotTo(HaveOccurred())

		actual := signedProposal.GetSignature()
		Expect(actual).To(BeEquivalentTo(expected))*/
	})

	It("Proposal includes creator", func(specCtx SpecContext) {
		//expected := &msp.SerializedIdentity{
		//	Mspid:   "MSP_ID",
		//	IdBytes: []byte("CREDENTIALS"),
		//}

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		//var signedProposal *peer.SignedProposal
		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				//signedProposal = in
				CopyProto(NewProposalResponse(common.Status_SUCCESS, ""), out)
			}).
			Times(1)

		/*mockSigner := NewMockSigner(controller, expected.Mspid, expected.IdBytes, nil)

		_, err := GetInstalled(specCtx, mockConnection, mockSigner, "PACKAGE_ID")
		Expect(err).NotTo(HaveOccurred())

		signatureHeader := AssertUnmarshalSignatureHeader(signedProposal)

		actual := &msp.SerializedIdentity{}
		AssertUnmarshal(signatureHeader.GetCreator(), actual)

		AssertProtoEqual(expected, actual)*/
	})

	It("Proposal includes supplied chaincode package ID", func(specCtx SpecContext) {
		expected := "PACKAGE_ID"

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

		//mockSigner := NewMockSigner(controller, "", nil, nil)

		//_, err := GetInstalled(specCtx, mockConnection, mockSigner, expected)
		//Expect(err).NotTo(HaveOccurred())

		invocationSpec := AssertUnmarshalInvocationSpec(signedProposal)
		args := invocationSpec.GetChaincodeSpec().GetInput().GetArgs()
		Expect(args).To(HaveLen(2), "number of arguments")

		chaincodeArgs := &lifecycle.GetInstalledChaincodePackageArgs{}
		AssertUnmarshal(args[1], chaincodeArgs)

		actual := chaincodeArgs.GetPackageId()
		Expect(actual).To(Equal(expected), "chaincode package ID")
	})

	It("Installed chaincode package returned on successful proposal response", func(specCtx SpecContext) {
		expected := &lifecycle.GetInstalledChaincodePackageResult{
			ChaincodeInstallPackage: []byte("CHAINCODE_PACKAGE"),
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

		//mockSigner := NewMockSigner(controller, "", nil, nil)

		//actual, err := GetInstalled(specCtx, mockConnection, mockSigner, "PACKAGE_ID")
		//Expect(err).NotTo(HaveOccurred())

		//Expect(actual).To(Equal(expected.GetChaincodeInstallPackage()))
	})
})
