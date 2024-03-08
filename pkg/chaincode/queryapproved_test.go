/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"

	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("QueryApproved", func() {
	var channelName string
	var chaincodeName string
	var sequence int64

	BeforeEach(func() {
		channelName = "mockchannel"
		chaincodeName = "CHAINCODE_NAME"
		sequence = 1
	})

	It("gRPC calls made with supplied context", func(specCtx SpecContext) {
		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		var evaluateCtxErr error

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEvaluateMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.EvaluateRequest, out *gateway.EvaluateResponse, opts ...grpc.CallOption) {
				evaluateCtxErr = ctx.Err()
				proto.Merge(out, NewEvaluateResponse(""))
			})

		mockSigner := NewMockSigner(controller, "", nil, nil)

		ctx, cancel := context.WithCancel(specCtx)
		cancel()

		_, _ = QueryApproved(ctx, mockConnection, mockSigner, channelName, chaincodeName, sequence)

		Expect(evaluateCtxErr).To(BeIdenticalTo(context.Canceled), "endorse context error")
	})

	It("Endorse errors returned", func(specCtx SpecContext) {
		expectedErr := status.Error(codes.Unavailable, "EXPECTED_ERROR")

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEvaluateMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr)

		mockSigner := NewMockSigner(controller, "", nil, nil)

		_, err := QueryApproved(specCtx, mockConnection, mockSigner, channelName, chaincodeName, sequence)

		Expect(err).To(MatchError(expectedErr))
		AssertEqualStatus(expectedErr, err)
	})

	It("Proposal content", func(specCtx SpecContext) {
		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		expected := &lifecycle.QueryApprovedChaincodeDefinitionArgs{
			Name:     chaincodeName,
			Sequence: sequence,
		}

		var evaluateRequest *gateway.EvaluateRequest
		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEvaluateMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.EvaluateRequest, out *gateway.EvaluateResponse, opts ...grpc.CallOption) {
				evaluateRequest = in
				proto.Merge(out, NewEvaluateResponse(""))
			}).
			Times(1)
		mockSigner := NewMockSigner(controller, "", nil, nil)

		_, err := QueryApproved(specCtx, mockConnection, mockSigner, channelName, chaincodeName, sequence)
		Expect(err).NotTo(HaveOccurred())

		invocationSpec := AssertUnmarshalInvocationSpec(evaluateRequest.GetProposedTransaction())
		args := invocationSpec.GetChaincodeSpec().GetInput().GetArgs()
		Expect(args).To(HaveLen(2), "number of arguments")

		actual := &lifecycle.QueryApprovedChaincodeDefinitionArgs{}
		AssertUnmarshal(args[1], actual)

		AssertProtoEqual(expected, actual)
	})
})
