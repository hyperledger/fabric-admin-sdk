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
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
)

var _ = Describe("QueryCommitted", func() {
	var channelName string

	BeforeEach(func() {
		channelName = "mockchannel"
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
				CopyProto(NewEvaluateResponse(channelName, ""), out)
			})

		mockSigner := NewMockSigner(controller, "", nil, nil)

		ctx, cancel := context.WithCancel(specCtx)
		cancel()

		_, _ = QueryCommitted(ctx, mockConnection, mockSigner, channelName)

		Expect(evaluateCtxErr).To(BeIdenticalTo(context.Canceled), "endorse context error")
	})

	It("Endorse errors returned", func(specCtx SpecContext) {
		expectedErr := errors.New("EXPECTED_ERROR")

		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEvaluateMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr)

		mockSigner := NewMockSigner(controller, "", nil, nil)

		_, err := QueryCommitted(specCtx, mockConnection, mockSigner, channelName)

		Expect(err).To(MatchError(expectedErr))
	})

	It("Proposal content", func(specCtx SpecContext) {
		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		expected := &lifecycle.QueryChaincodeDefinitionsArgs{}

		var evaluateRequest *gateway.EvaluateRequest
		mockConnection := NewMockClientConnInterface(controller)
		mockConnection.EXPECT().
			Invoke(gomock.Any(), gomock.Eq(gatewayEvaluateMethod), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, method string, in *gateway.EvaluateRequest, out *gateway.EvaluateResponse, opts ...grpc.CallOption) {
				evaluateRequest = in
				CopyProto(NewEvaluateResponse(channelName, ""), out)
			}).
			Times(1)
		mockSigner := NewMockSigner(controller, "", nil, nil)

		_, err := QueryCommitted(specCtx, mockConnection, mockSigner, channelName)
		Expect(err).NotTo(HaveOccurred())

		invocationSpec := AssertUnmarshalInvocationSpec(evaluateRequest.GetProposedTransaction())
		args := invocationSpec.GetChaincodeSpec().GetInput().GetArgs()
		Expect(args).To(HaveLen(2), "number of arguments")

		actual := &lifecycle.QueryChaincodeDefinitionsArgs{}
		AssertUnmarshal(args[1], actual)

		AssertProtoEqual(expected, actual)
	})
})