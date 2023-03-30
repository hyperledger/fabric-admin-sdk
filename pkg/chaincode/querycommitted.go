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

// QueryCommitted returns the definitions of all committed chaincode for a given channel. The connection may be to any
// Gateway peer that is a member of the channel.
func QueryCommitted(ctx context.Context, network client.Network) (*lifecycle.QueryChaincodeDefinitionsResult, error) {
	queryArgs := &lifecycle.QueryChaincodeDefinitionsArgs{}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return nil, err
	}

	resultBytes, err := network.GetContract(lifecycleChaincodeName).
		EvaluateWithContext(
			ctx,
			queryCommittedTransactionName,
			client.WithBytesArguments(queryArgsBytes),
		)
	if err != nil {
		return nil, fmt.Errorf("failed to query committed chaincodes: %w", err)
	}

	result := &lifecycle.QueryChaincodeDefinitionsResult{}
	if err = proto.Unmarshal(resultBytes, result); err != nil {
		return nil, fmt.Errorf("failed to deserialize query committed chaincode result: %w", err)
	}

	return result, nil
}
