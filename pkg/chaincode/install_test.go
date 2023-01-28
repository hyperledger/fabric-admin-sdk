/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

//go:generate mockgen -destination ./clientconnection_mock_test.go -package ${GOPACKAGE} google.golang.org/grpc ClientConnInterface
//go:generate mockgen -destination ./signingidentity_mock_test.go -package ${GOPACKAGE} github.com/hyperledger/fabric-admin-sdk/pkg/identity SigningIdentity

func NewMockSigner(controller *gomock.Controller, mspID string, credentials []byte, signature []byte) *MockSigningIdentity {
	id := NewMockSigningIdentity(controller)
	id.EXPECT().MspID().Return(mspID).AnyTimes()
	id.EXPECT().Credentials().Return(credentials).AnyTimes()
	id.EXPECT().Sign(gomock.Any()).Return(signature, nil).AnyTimes()

	return id
}

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
	var packageReader io.Reader

	BeforeEach(func() {
		packageReader = strings.NewReader("CHAINCODE_PACKAGE")
	})

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

		err := chaincode.Install(ctx, mockConnection, mockSigner, packageReader)

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

		err := chaincode.Install(specCtx, mockConnection, mockSigner, packageReader)

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

		err := chaincode.Install(specCtx, mockConnection, mockSigner, packageReader)

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

		err := chaincode.Install(specCtx, mockConnection, mockSigner, packageReader)
		Expect(err).NotTo(HaveOccurred())

		actual := signedProposal.GetSignature()
		Expect(actual).To(BeEquivalentTo(expected))
	})

	It("Proposal includes supplied chaincode package", func(specCtx SpecContext) {
		expected := []byte("MY_CHAINCODE_PACKAGE")

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

		mockSigner := NewMockSigner(controller, "", nil, nil)
		packageReader = bytes.NewReader(expected)

		err := chaincode.Install(specCtx, mockConnection, mockSigner, packageReader)
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

		err := chaincode.Install(specCtx, mockConnection, mockSigner, packageReader)
		Expect(err).NotTo(HaveOccurred())

		signatureHeader := AssertUnmarshalSignatureHeader(signedProposal)

		actual := &msp.SerializedIdentity{}
		AssertUnmarshal(signatureHeader.GetCreator(), actual)

		AssertProtoEqual(expected, actual)
	})
})
