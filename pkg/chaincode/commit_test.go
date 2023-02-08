/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
)

var _ = Describe("Commit", func() {
	var channelName string
	var chaincodeDefinition *Definition

	BeforeEach(func() {
		channelName = "CHANNEL"
		chaincodeDefinition = &Definition{
			Name:        "CHAINCODE",
			Version:     "1.0",
			Sequence:    1,
			ChannelName: channelName,
		}
	})

	It("gRPC calls made with supplied context", func(specCtx SpecContext) {
		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		var endorseCtxErr error
		var submitCtxErr error

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEndorseMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.EndorseRequest, out *gateway.EndorseResponse, opts ...grpc.CallOption) {
				endorseCtxErr = ctx.Err()
				CopyProto(NewEndorseResponse(channelName, ""), out)
			})
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewaySubmitMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.SubmitRequest, out *gateway.SubmitResponse, opts ...grpc.CallOption) {
				submitCtxErr = ctx.Err()
				CopyProto(NewSubmitResponse(), out)
			})
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayCommitStatusMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, method string, in *gateway.SignedCommitStatusRequest, out *gateway.CommitStatusResponse, opts ...grpc.CallOption) error {
				CopyProto(NewCommitStatusResponse(peer.TxValidationCode_VALID, 0), out)
				return ctx.Err()
			})

		mockSigner := NewMockSigner(controller, "", nil, nil)

		ctx, cancel := context.WithCancel(specCtx)
		cancel()

		err := Commit(ctx, mockConnection, mockSigner, chaincodeDefinition)

		Expect(endorseCtxErr).To(BeIdenticalTo(context.Canceled), "endorse context error")
		Expect(submitCtxErr).To(BeIdenticalTo(context.Canceled), "submit context error")
		Expect(err).To(MatchError(context.Canceled), "endorse context error")
	})

	It("Endorse errors returned", func(specCtx SpecContext) {
		expectedErr := errors.New("EXPECTED_ERROR")

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEndorseMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr)

		mockSigner := NewMockSigner(controller, "", nil, nil)

		err := Commit(specCtx, mockConnection, mockSigner, chaincodeDefinition)

		Expect(err).To(MatchError(expectedErr))
	})

	It("Submit errors returned", func(specCtx SpecContext) {
		expectedErr := errors.New("EXPECTED_ERROR")

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEndorseMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.EndorseRequest, out *gateway.EndorseResponse, opts ...grpc.CallOption) {
				CopyProto(NewEndorseResponse(channelName, ""), out)
			})
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewaySubmitMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.SubmitRequest, out *gateway.SubmitResponse, opts ...grpc.CallOption) {
				CopyProto(NewSubmitResponse(), out)
			})
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayCommitStatusMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr)

		mockSigner := NewMockSigner(controller, "", nil, nil)

		err := Commit(specCtx, mockConnection, mockSigner, chaincodeDefinition)

		Expect(err).To(MatchError(expectedErr))
	})

	It("CommitStatus errors returned", func(specCtx SpecContext) {
		expectedErr := errors.New("EXPECTED_ERROR")

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEndorseMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.EndorseRequest, out *gateway.EndorseResponse, opts ...grpc.CallOption) {
				CopyProto(NewEndorseResponse(channelName, ""), out)
			})
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewaySubmitMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr)

		mockSigner := NewMockSigner(controller, "", nil, nil)

		err := Commit(specCtx, mockConnection, mockSigner, chaincodeDefinition)

		Expect(err).To(MatchError(expectedErr))
	})

	It("Proposal content", func(specCtx SpecContext) {
		expected := &lifecycle.CommitChaincodeDefinitionArgs{
			Name:     chaincodeDefinition.Name,
			Version:  chaincodeDefinition.Version,
			Sequence: chaincodeDefinition.Sequence,
		}

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		var endorseRequest *gateway.EndorseRequest
		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEndorseMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.EndorseRequest, out *gateway.EndorseResponse, opts ...grpc.CallOption) {
				endorseRequest = in
				CopyProto(NewEndorseResponse(channelName, ""), out)
			}).
			Times(1)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewaySubmitMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.SubmitRequest, out *gateway.SubmitResponse, opts ...grpc.CallOption) {
				CopyProto(NewSubmitResponse(), out)
			})
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayCommitStatusMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.SignedCommitStatusRequest, out *gateway.CommitStatusResponse, opts ...grpc.CallOption) {
				CopyProto(NewCommitStatusResponse(peer.TxValidationCode_VALID, 0), out)
			})

		mockSigner := NewMockSigner(controller, "", nil, nil)

		err := Commit(specCtx, mockConnection, mockSigner, chaincodeDefinition)
		Expect(err).NotTo(HaveOccurred())

		invocationSpec := AssertUnmarshalInvocationSpec(endorseRequest.GetProposedTransaction())
		args := invocationSpec.GetChaincodeSpec().GetInput().GetArgs()
		Expect(args).To(HaveLen(2), "number of arguments")

		actual := &lifecycle.CommitChaincodeDefinitionArgs{}
		AssertUnmarshal(args[1], actual)

		AssertProtoEqual(expected, actual)
	})
})
