/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode_test

import (
	"context"

	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
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

var _ = Describe("CheckCommitReadiness", func() {
	var channelName string
	var chaincodeDefinition *chaincode.Definition

	BeforeEach(func() {
		channelName = "mockchannel"
		chaincodeDefinition = &chaincode.Definition{
			Name:        "CHAINCODE",
			Version:     "1.0",
			Sequence:    1,
			ChannelName: channelName,
		}
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
		gateway := chaincode.NewGateway(mockConnection, mockSigner)

		ctx, cancel := context.WithCancel(specCtx)
		cancel()

		_, _ = gateway.CheckCommitReadiness(ctx, chaincodeDefinition)

		Expect(evaluateCtxErr).To(BeIdenticalTo(context.Canceled))
	})

	It("Evaluate errors returned", func(specCtx SpecContext) {
		expectedErr := status.Error(codes.Unavailable, "EXPECTED_ERROR")

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEvaluateMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr)

		mockSigner := NewMockSigner(controller, "", nil, nil)
		gateway := chaincode.NewGateway(mockConnection, mockSigner)

		_, err := gateway.CheckCommitReadiness(specCtx, chaincodeDefinition)

		Expect(err).To(MatchError(expectedErr))
		AssertEqualStatus(expectedErr, err)
	})

	It("Proposal content", func(specCtx SpecContext) {
		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		expected := &lifecycle.CheckCommitReadinessArgs{
			Name:     chaincodeDefinition.Name,
			Version:  chaincodeDefinition.Version,
			Sequence: chaincodeDefinition.Sequence,
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
		gateway := chaincode.NewGateway(mockConnection, mockSigner)

		_, err := gateway.CheckCommitReadiness(specCtx, chaincodeDefinition)
		Expect(err).NotTo(HaveOccurred())

		invocationSpec := AssertUnmarshalInvocationSpec(evaluateRequest.GetProposedTransaction())
		args := invocationSpec.GetChaincodeSpec().GetInput().GetArgs()
		Expect(args).To(HaveLen(2), "number of arguments")

		actual := &lifecycle.CheckCommitReadinessArgs{}
		AssertUnmarshal(args[1], actual)

		AssertProtoEqual(expected, actual)
	})
})
