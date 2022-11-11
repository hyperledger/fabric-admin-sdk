package chaincode_test

import (
	"bytes"
	"context"
	"fabric-admin-sdk/internal/pkg/identity/mocks"
	"fabric-admin-sdk/pkg/chaincode"
	"io"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/runtime/protoiface"
)

//go:generate mockgen -destination ./endorser_mock_test.go -package ${GOPACKAGE} github.com/hyperledger/fabric-protos-go/peer EndorserClient

func NewProposalResponse(status common.Status, message string) *peer.ProposalResponse {
	return &peer.ProposalResponse{
		Response: &peer.Response{
			Status:  int32(status),
			Message: message,
		},
	}
}

// AssertUnmarshal ensures that a protobuf is umarshaled without error
func AssertUnmarshal(b []byte, m protoiface.MessageV1) {
	err := proto.Unmarshal(b, m)
	Expect(err).NotTo(HaveOccurred())
}

func AssertUnmarshalProposal(signedProposal *peer.SignedProposal) *peer.Proposal {
	proposal := &peer.Proposal{}
	AssertUnmarshal(signedProposal.ProposalBytes, proposal)

	return proposal
}

func AssertUnmarshalSignatureHeader(signedProposal *peer.SignedProposal) *common.SignatureHeader {
	proposal := AssertUnmarshalProposal(signedProposal)

	header := &common.Header{}
	AssertUnmarshal(proposal.Header, header)

	signatureHeader := &common.SignatureHeader{}
	AssertUnmarshal(header.SignatureHeader, signatureHeader)

	return signatureHeader
}

// AssertUnmarshalProposalPayload ensures that a ChaincodeProposalPayload protobuf is umarshalled without error
func AssertUnmarshalProposalPayload(signedProposal *peer.SignedProposal) *peer.ChaincodeProposalPayload {
	proposal := AssertUnmarshalProposal(signedProposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshal(proposal.Payload, payload)

	return payload
}

// AssertUnmarshalInvocationSpec ensures that a ChaincodeInvocationSpec protobuf is umarshalled without error
func AssertUnmarshalInvocationSpec(signedProposal *peer.SignedProposal) *peer.ChaincodeInvocationSpec {
	proposal := &peer.Proposal{}
	AssertUnmarshal(signedProposal.ProposalBytes, proposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshal(proposal.Payload, payload)

	input := &peer.ChaincodeInvocationSpec{}
	AssertUnmarshal(payload.Input, input)

	return input
}

var _ = Describe("Install", func() {
	var mockSigner *mocks.SignerSerializer
	var packageReader io.Reader

	BeforeEach(func() {
		mockSigner = &mocks.SignerSerializer{}
		packageReader = strings.NewReader("CHAINCODE_PACKAGE")
	})

	It("Endorser client called with supplied context", func(specCtx SpecContext) {
		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			Return(NewProposalResponse(common.Status_SUCCESS, ""), nil)

		err := chaincode.Install(ctx, mockEndorser, mockSigner, packageReader)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Endorser client errors returned", func(specCtx SpecContext) {
		expectedErr := errors.New("EXPECTED_ERROR")

		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, expectedErr)

		err := chaincode.Install(ctx, mockEndorser, mockSigner, packageReader)

		Expect(err).To(MatchError(expectedErr))
	})

	It("Unsuccessful proposal response gives error", func(specCtx SpecContext) {
		expectedStatus := common.Status_BAD_REQUEST
		expectedMessage := "EXPECTED_ERROR"

		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
		defer controller.Finish()

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(NewProposalResponse(expectedStatus, expectedMessage), nil)

		err := chaincode.Install(ctx, mockEndorser, mockSigner, packageReader)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(And(
			ContainSubstring("%d", expectedStatus),
			ContainSubstring(expectedStatus.String()),
			ContainSubstring(expectedMessage),
		))
	})

	It("Uses signer", func(specCtx SpecContext) {
		expected := []byte("SIGNATURE")

		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
		defer controller.Finish()

		var signedProposal *peer.SignedProposal
		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *peer.SignedProposal, _ ...grpc.CallOption) {
				signedProposal = in
			}).
			Return(NewProposalResponse(common.Status_SUCCESS, ""), nil).
			Times(1)

		mockSigner.SignReturns(expected, nil)

		err := chaincode.Install(ctx, mockEndorser, mockSigner, packageReader)
		Expect(err).NotTo(HaveOccurred())

		actual := signedProposal.GetSignature()
		Expect(actual).To(BeEquivalentTo(expected))
	})

	It("Proposal includes supplied chaincode package", func(specCtx SpecContext) {
		expected := []byte("MY_CHAINCODE_PACKAGE")

		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
		defer controller.Finish()

		var signedProposal *peer.SignedProposal
		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *peer.SignedProposal, _ ...grpc.CallOption) {
				signedProposal = in
			}).
			Return(NewProposalResponse(common.Status_SUCCESS, ""), nil).
			Times(1)

		packageReader = bytes.NewReader(expected)

		err := chaincode.Install(ctx, mockEndorser, mockSigner, packageReader)
		Expect(err).NotTo(HaveOccurred())

		invocationSpec := AssertUnmarshalInvocationSpec(signedProposal)
		args := invocationSpec.GetChaincodeSpec().GetInput().GetArgs()
		Expect(args).To(HaveLen(2), "number of arguments")

		chaincodeArgs := &lifecycle.InstallChaincodeArgs{}
		AssertUnmarshal(args[1], chaincodeArgs)

		actual := chaincodeArgs.GetChaincodeInstallPackage()
		Expect(actual).To(BeEquivalentTo(expected), "chaincode pakcage")
	})

	It("Proposal includes creator", func(specCtx SpecContext) {
		expected := []byte("MY_SIGNATURE")

		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
		defer controller.Finish()

		var signedProposal *peer.SignedProposal
		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *peer.SignedProposal, _ ...grpc.CallOption) {
				signedProposal = in
			}).
			Return(NewProposalResponse(common.Status_SUCCESS, ""), nil).
			Times(1)

		mockSigner.SerializeReturns(expected, nil)

		err := chaincode.Install(ctx, mockEndorser, mockSigner, packageReader)
		Expect(err).NotTo(HaveOccurred())

		signatureHeader := AssertUnmarshalSignatureHeader(signedProposal)

		actual := signatureHeader.GetCreator()
		Expect(actual).To(BeEquivalentTo(expected), "signature")
	})
})
