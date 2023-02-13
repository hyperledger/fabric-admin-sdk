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

// QueryCommittedWithName gets the chaincode definition for the given chaincode name in the given channel
// with a specific signing identity
func QueryCommittedWithName(ctx context.Context, connection grpc.ClientConnInterface, signingID identity.SigningIdentity, channelID, chaincodeName string) (*lifecycle.QueryChaincodeDefinitionResult, error) {
	queryArgs := &lifecycle.QueryChaincodeDefinitionArgs{
		Name: chaincodeName,
	}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return nil, err
	}

	gw, err := gateway.New(connection, signingID)
	if err != nil {
		return nil, err
	}
	defer gw.Close()

	r, err := gw.GetNetwork(channelID).
		GetContract(lifecycleChaincodeName).
		EvaluateWithContext(
			ctx,
			queryCommittedWithNameTransactionName,
			client.WithBytesArguments(queryArgsBytes),
		)
	if err != nil {
		return nil, fmt.Errorf("failed to query committed chaincode: %w", err)
	}

	result := &lifecycle.QueryChaincodeDefinitionResult{}
	if err = proto.Unmarshal(r, result); err != nil {
		return nil, fmt.Errorf("failed to deserialize query committed chaincode result: %w", err)
	}

	return result, nil
}
