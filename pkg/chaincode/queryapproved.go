/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/internal/gateway"
	"github.com/hyperledger/fabric-gateway/pkg/client"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// QueryApproved chaincode definition for the user's own organization.
func QueryApproved(ctx context.Context, connection grpc.ClientConnInterface, id identity.SigningIdentity, channelID string, chaincodeName string, sequence int64) (*lifecycle.QueryApprovedChaincodeDefinitionResult, error) {
	queryArgs := &lifecycle.QueryApprovedChaincodeDefinitionArgs{
		Name:     chaincodeName,
		Sequence: sequence,
	}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return nil, err
	}

	gw, err := gateway.New(connection, id)
	if err != nil {
		return nil, err
	}
	defer gw.Close()

	resultBytes, err := gw.GetNetwork(channelID).
		GetContract(lifecycleChaincodeName).
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
