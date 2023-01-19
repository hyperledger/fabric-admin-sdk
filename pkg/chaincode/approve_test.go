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
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
)

const gatewayEndorseMethod = "/gateway.Gateway/Endorse"
const gatewaySubmitMethod = "/gateway.Gateway/Submit"
const gatewayCommitStatusMethod = "/gateway.Gateway/CommitStatus"

func NewEndorseResponse(channelName string, result string) *gateway.EndorseResponse {
	return &gateway.EndorseResponse{
		PreparedTransaction: &common.Envelope{
			Payload: AssertMarshal(&common.Payload{
				Header: &common.Header{
					ChannelHeader: AssertMarshal(&common.ChannelHeader{
						ChannelId: channelName,
					}),
				},
				Data: AssertMarshal(&peer.Transaction{
					Actions: []*peer.TransactionAction{
						{
							Payload: AssertMarshal(&peer.ChaincodeActionPayload{
								Action: &peer.ChaincodeEndorsedAction{
									ProposalResponsePayload: AssertMarshal(&peer.ProposalResponsePayload{
										Extension: AssertMarshal(&peer.ChaincodeAction{
											Response: &peer.Response{
												Payload: []byte(result),
											},
										}),
									}),
								},
							}),
						},
					},
				}),
			}),
		},
	}
}

func NewSubmitResponse() *gateway.SubmitResponse {
	return &gateway.SubmitResponse{}
}

func NewCommitStatusResponse(result peer.TxValidationCode, blockNumber uint64) *gateway.CommitStatusResponse {
	return &gateway.CommitStatusResponse{
		Result:      result,
		BlockNumber: blockNumber,
	}
}

var _ = Describe("Approve", func() {
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

		err := Approve(ctx, mockConnection, mockSigner, chaincodeDefinition)

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

		err := Approve(specCtx, mockConnection, mockSigner, chaincodeDefinition)

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

		err := Approve(specCtx, mockConnection, mockSigner, chaincodeDefinition)

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

		err := Approve(specCtx, mockConnection, mockSigner, chaincodeDefinition)

		Expect(err).To(MatchError(expectedErr))
	})

	DescribeTable("Proposal content",
		func(specCtx SpecContext, newInput func(*Definition) *Definition, newExpected func(*lifecycle.ApproveChaincodeDefinitionForMyOrgArgs) *lifecycle.ApproveChaincodeDefinitionForMyOrgArgs) {
			input := newInput(chaincodeDefinition)
			expected := newExpected(&lifecycle.ApproveChaincodeDefinitionForMyOrgArgs{
				Name:     chaincodeDefinition.Name,
				Version:  chaincodeDefinition.Version,
				Sequence: chaincodeDefinition.Sequence,
				Source:   &lifecycle.ChaincodeSource{},
			})

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

			err := Approve(specCtx, mockConnection, mockSigner, input)
			Expect(err).NotTo(HaveOccurred())

			invocationSpec := AssertUnmarshalInvocationSpec(endorseRequest.GetProposedTransaction())
			args := invocationSpec.GetChaincodeSpec().GetInput().GetArgs()
			Expect(args).To(HaveLen(2), "number of arguments")

			actual := &lifecycle.ApproveChaincodeDefinitionForMyOrgArgs{}
			AssertUnmarshal(args[1], actual)

			AssertProtoEqual(expected, actual)
		},
		Entry(
			"Proposal includes specified package ID",
			func(in *Definition) *Definition {
				in.PackageID = "PACKAGE_ID"
				return in
			},
			func(in *lifecycle.ApproveChaincodeDefinitionForMyOrgArgs) *lifecycle.ApproveChaincodeDefinitionForMyOrgArgs {
				in.Source.Type = &lifecycle.ChaincodeSource_LocalPackage{
					LocalPackage: &lifecycle.ChaincodeSource_Local{
						PackageId: chaincodeDefinition.PackageID,
					},
				}
				return in
			},
		),
		Entry(
			"Proposal includes unspecified chaincode source with no package ID specified",
			func(in *Definition) *Definition {
				return in
			},
			func(in *lifecycle.ApproveChaincodeDefinitionForMyOrgArgs) *lifecycle.ApproveChaincodeDefinitionForMyOrgArgs {
				in.Source.Type = &lifecycle.ChaincodeSource_Unavailable_{
					Unavailable: &lifecycle.ChaincodeSource_Unavailable{},
				}
				return in
			},
		),
	)
})
