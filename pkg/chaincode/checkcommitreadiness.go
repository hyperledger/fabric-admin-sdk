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

// CheckCommitReadiness for a chaincode and return all approval records. The connection may be to any Gateway peer that
// is a member of the channel.
func CheckCommitReadiness(ctx context.Context, connection grpc.ClientConnInterface, id identity.SigningIdentity, chaincodeDef *Definition) (*lifecycle.CheckCommitReadinessResult, error) {
	args := &lifecycle.CheckCommitReadinessArgs{
		Name:                chaincodeDef.Name,
		Version:             chaincodeDef.Version,
		Sequence:            chaincodeDef.Sequence,
		EndorsementPlugin:   chaincodeDef.EndorsementPlugin,
		ValidationPlugin:    chaincodeDef.ValidationPlugin,
		ValidationParameter: chaincodeDef.validationParameter,
		Collections:         chaincodeDef.Collections,
		InitRequired:        chaincodeDef.InitRequired,
	}
	argsBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, err
	}

	gw, err := gateway.New(connection, id)
	if err != nil {
		return nil, err
	}
	defer gw.Close()

	r, err := gw.GetNetwork(chaincodeDef.ChannelName).
		GetContract(lifecycleChaincodeName).
		EvaluateWithContext(
			ctx,
			checkCommitReadinessTransactionName,
			client.WithBytesArguments(argsBytes),
		)
	if err != nil {
		return nil, fmt.Errorf("failed to check commit readiness: %w", err)
	}

	result := &lifecycle.CheckCommitReadinessResult{}
	if err = proto.Unmarshal(r, result); err != nil {
		return nil, fmt.Errorf("failed to deserialize check commit readiness result: %w", err)
	}

	return result, nil
}
