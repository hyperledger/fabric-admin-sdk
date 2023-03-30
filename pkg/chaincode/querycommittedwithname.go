/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/protobuf/proto"
)

// QueryCommittedWithName returns the definition of the named chaincode for a given channel. The connection may be to
// any Gateway peer that is a member of the channel.
func QueryCommittedWithName(ctx context.Context, network client.Network, chaincodeName string) (*lifecycle.QueryChaincodeDefinitionResult, error) {
	queryArgs := &lifecycle.QueryChaincodeDefinitionArgs{
		Name: chaincodeName,
	}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return nil, err
	}

	resultBytes, err := network.GetContract(lifecycleChaincodeName).
		EvaluateWithContext(
			ctx,
			queryCommittedWithNameTransactionName,
			client.WithBytesArguments(queryArgsBytes),
		)
	if err != nil {
		return nil, fmt.Errorf("failed to query committed chaincode: %w", err)
	}

	result := &lifecycle.QueryChaincodeDefinitionResult{}
	if err = proto.Unmarshal(resultBytes, result); err != nil {
		return nil, fmt.Errorf("failed to deserialize query committed chaincode result: %w", err)
	}

	return result, nil
}
