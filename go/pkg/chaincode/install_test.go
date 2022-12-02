package chaincode_test

import (
	"bytes"
	"context"
	"errors"
	"fabric-admin-sdk/internal/pkg/identity/mocks"
	"fabric-admin-sdk/pkg/chaincode"
	"io"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

//go:generate mockgen -destination ./clientconnection_mock_test.go -package ${GOPACKAGE} google.golang.org/grpc ClientConnInterface

func NewProposalResponse(status common.Status, message string) *peer.ProposalResponse {
	return &peer.ProposalResponse{
		Response: &peer.Response{
			Status:  int32(status),
			Message: message,
		},
	}
}

// AssertUnmarshal ensures that a protobuf is umarshaled without error
func AssertUnmarshal(b []byte, m proto.Message) {
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

func CopyProto(from proto.Message, to proto.Message) {
	protoBytes, err := proto.Marshal(from)
	Expect(err).NotTo(HaveOccurred())
	err = proto.Unmarshal(protoBytes, to)
	Expect(err).NotTo(HaveOccurred())
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

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Eq(ctx), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				CopyProto(NewProposalResponse(common.Status_SUCCESS, ""), out)
			})

		err := chaincode.Install(ctx, mockConnection, mockSigner, packageReader)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Endorser client errors returned", func(specCtx SpecContext) {
		expectedErr := errors.New("EXPECTED_ERROR")

		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr)

		err := chaincode.Install(ctx, mockConnection, mockSigner, packageReader)

		Expect(err).To(MatchError(expectedErr))
	})

	It("Unsuccessful proposal response gives error", func(specCtx SpecContext) {
		expectedStatus := common.Status_BAD_REQUEST
		expectedMessage := "EXPECTED_ERROR"

		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				CopyProto(NewProposalResponse(expectedStatus, expectedMessage), out)
			})

		err := chaincode.Install(ctx, mockConnection, mockSigner, packageReader)

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
		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				signedProposal = in
				CopyProto(NewProposalResponse(common.Status_SUCCESS, ""), out)
			}).
			Times(1)

		mockSigner.SignReturns(expected, nil)

		err := chaincode.Install(ctx, mockConnection, mockSigner, packageReader)
		Expect(err).NotTo(HaveOccurred())

		actual := signedProposal.GetSignature()
		Expect(actual).To(BeEquivalentTo(expected))
	})

	It("Proposal includes supplied chaincode package", func(specCtx SpecContext) {
		expected := []byte("MY_CHAINCODE_PACKAGE")

		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
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

		packageReader = bytes.NewReader(expected)

		err := chaincode.Install(ctx, mockConnection, mockSigner, packageReader)
		Expect(err).NotTo(HaveOccurred())

		invocationSpec := AssertUnmarshalInvocationSpec(signedProposal)
		args := invocationSpec.GetChaincodeSpec().GetInput().GetArgs()
		Expect(args).To(HaveLen(2), "number of arguments")

		chaincodeArgs := &lifecycle.InstallChaincodeArgs{}
		AssertUnmarshal(args[1], chaincodeArgs)

		actual := chaincodeArgs.GetChaincodeInstallPackage()
		Expect(actual).To(BeEquivalentTo(expected), "chaincode package")
	})

	It("Proposal includes creator", func(specCtx SpecContext) {
		expected := []byte("MY_SIGNATURE")

		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
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

		mockSigner.SerializeReturns(expected, nil)

		err := chaincode.Install(ctx, mockConnection, mockSigner, packageReader)
		Expect(err).NotTo(HaveOccurred())

		signatureHeader := AssertUnmarshalSignatureHeader(signedProposal)

		actual := signatureHeader.GetCreator()
		Expect(actual).To(BeEquivalentTo(expected), "signature")
	})
})
