/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"
	"fabric-admin-sdk/pkg/identity"
	"fabric-admin-sdk/pkg/internal/proposal"
	"fmt"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const queryCommittedTransactionName = "QueryChaincodeDefinitions"

// QueryCommitted chaincode on a specific peer.
func QueryCommitted(ctx context.Context, connection grpc.ClientConnInterface, signingID identity.SigningIdentity) error {
	queryArgs := &lifecycle.QueryChaincodeDefinitionArgs{}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return fmt.Errorf("failed to query committed chaincode fail to marshal data: %w", err)
	}

	proposalProto, err := proposal.NewProposal(signingID, lifecycleChaincodeName, queryCommittedTransactionName, proposal.WithArguments(queryArgsBytes))
	if err != nil {
		return fmt.Errorf("failed to query committed chaincode fail to create proposal: %w", err)
	}

	signedProposal, err := proposal.NewSignedProposal(proposalProto, signingID)
	if err != nil {
		return fmt.Errorf("failed to query committed chaincode fail to sign proposal: %w", err)
	}

	endorser := peer.NewEndorserClient(connection)

	proposalResponse, err := endorser.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return fmt.Errorf("failed to query committed chaincode error from endorser: %w", err)
	}

	if err = proposal.CheckSuccessfulResponse(proposalResponse); err != nil {
		return fmt.Errorf("failed to query committed chaincode error from check response checking: %w", err)
	}

	return printResponseAsJSON(proposalResponse, &lifecycle.QueryChaincodeDefinitionResult{})
}
