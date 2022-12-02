package chaincode_test

import (
	"context"
	"errors"
	"fabric-admin-sdk/internal/pkg/identity/mocks"
	"fabric-admin-sdk/pkg/chaincode"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	var mockSigner *mocks.SignerSerializer

	BeforeEach(func() {
		mockSigner = &mocks.SignerSerializer{}
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

		_, err := chaincode.QueryInstalled(ctx, mockConnection, mockSigner)
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

		_, err := chaincode.QueryInstalled(ctx, mockConnection, mockSigner)

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

		_, err := chaincode.QueryInstalled(ctx, mockConnection, mockSigner)

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

		_, err := chaincode.QueryInstalled(ctx, mockConnection, mockSigner)
		Expect(err).NotTo(HaveOccurred())

		actual := signedProposal.GetSignature()
		Expect(actual).To(BeEquivalentTo(expected))
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

		_, err := chaincode.QueryInstalled(ctx, mockConnection, mockSigner)
		Expect(err).NotTo(HaveOccurred())

		signatureHeader := AssertUnmarshalSignatureHeader(signedProposal)

		actual := signatureHeader.GetCreator()
		Expect(actual).To(BeEquivalentTo(expected), "signature")
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

		response := NewProposalResponse(common.Status_SUCCESS, "")
		response.Response.Payload = AssertMarshal(expected)

		controller, ctx := gomock.WithContext(specCtx, GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(processProposalMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *peer.SignedProposal, out *peer.ProposalResponse, opts ...grpc.CallOption) {
				CopyProto(response, out)
			})

		actual, err := chaincode.QueryInstalled(ctx, mockConnection, mockSigner)
		Expect(err).NotTo(HaveOccurred())

		AssertProtoEqual(expected, actual)
	})
})
