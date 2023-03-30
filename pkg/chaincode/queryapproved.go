/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/client"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/protobuf/proto"
)

// QueryApproved chaincode definition for the user's own organization. The connection may be to any Gateway peer that is
// a member of the channel.
func QueryApproved(ctx context.Context, network client.Network, id identity.SigningIdentity, chaincodeName string, sequence int64) (*lifecycle.QueryApprovedChaincodeDefinitionResult, error) {
	queryArgs := &lifecycle.QueryApprovedChaincodeDefinitionArgs{
		Name:     chaincodeName,
		Sequence: sequence,
	}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return nil, err
	}

	resultBytes, err := network.GetContract(lifecycleChaincodeName).
		EvaluateWithContext(
			ctx,
			queryApprovedTransactionName,
			client.WithBytesArguments(queryArgsBytes),
			client.WithEndorsingOrganizations(id.MspID()),
		)
	if err != nil {
		return nil, fmt.Errorf("failed to query approved chaincode: %w", err)
	}

	result := &lifecycle.QueryApprovedChaincodeDefinitionResult{}
	if err = proto.Unmarshal(resultBytes, result); err != nil {
		return nil, fmt.Errorf("failed to deserialize query approved chaincode result: %w", err)
	}

	return result, nil
}
