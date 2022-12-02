/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"
	"fabric-admin-sdk/internal/pkg/identity"
	"fabric-admin-sdk/pkg/internal/proposal"
	"fmt"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const queryInstalledTransactionName = "QueryInstalledChaincodes"

// QueryInstalled chaincode on a specific peer.
func QueryInstalled(ctx context.Context, connection grpc.ClientConnInterface, signer identity.SignerSerializer) (*lifecycle.QueryInstalledChaincodesResult, error) {
	queryArgs := &lifecycle.QueryInstalledChaincodesArgs{}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return nil, err
	}

	proposalProto, err := proposal.NewProposal(signer, lifecycleChaincodeName, queryInstalledTransactionName, proposal.WithArguments(queryArgsBytes))
	if err != nil {
		return nil, err
	}

	signedProposal, err := proposal.NewSignedProposal(proposalProto, signer)
	if err != nil {
		return nil, err
	}

	endorser := peer.NewEndorserClient(connection)

	proposalResponse, err := endorser.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return nil, fmt.Errorf("failed to query installed chaincode: %w", err)
	}

	if err = proposal.CheckSuccessfulResponse(proposalResponse); err != nil {
		return nil, err
	}

	result := &lifecycle.QueryInstalledChaincodesResult{}
	if err = proto.Unmarshal(proposalResponse.GetResponse().GetPayload(), result); err != nil {
		return nil, fmt.Errorf("failed to deserialize query installed chaincode result: %w", err)
	}

	return result, nil
}
