package chaincode_test

import (
	"context"
	"fabric-admin-sdk/internal/pkg/identity/mocks"
	"fabric-admin-sdk/pkg/chaincode"

	"errors"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/runtime/protoiface"
)

func AssertMarshal(m protoiface.MessageV1) []byte {
	result, err := proto.Marshal(m)
	Expect(err).NotTo(HaveOccurred())
	return result
}

// AssertProtoEqual ensures an expected protobuf message matches an actual message
func AssertProtoEqual(expected protoiface.MessageV1, actual protoiface.MessageV1) {
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

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			Return(NewProposalResponse(common.Status_SUCCESS, ""), nil)

		_, err := chaincode.QueryInstalled(ctx, mockEndorser, mockSigner)
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

		_, err := chaincode.QueryInstalled(ctx, mockEndorser, mockSigner)

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

		_, err := chaincode.QueryInstalled(ctx, mockEndorser, mockSigner)

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

		_, err := chaincode.QueryInstalled(ctx, mockEndorser, mockSigner)
		Expect(err).NotTo(HaveOccurred())

		actual := signedProposal.GetSignature()
		Expect(actual).To(BeEquivalentTo(expected))
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

		_, err := chaincode.QueryInstalled(ctx, mockEndorser, mockSigner)
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

		mockEndorser := NewMockEndorserClient(controller)
		mockEndorser.EXPECT().
			ProcessProposal(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(response, nil)

		actual, err := chaincode.QueryInstalled(ctx, mockEndorser, mockSigner)
		Expect(err).NotTo(HaveOccurred())

		AssertProtoEqual(expected, actual)
	})
})
