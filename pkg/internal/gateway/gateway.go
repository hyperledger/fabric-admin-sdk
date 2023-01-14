/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
)

func self[T any](v T) T {
	return v
}

func New(connection grpc.ClientConnInterface, id identity.SigningIdentity) (*client.Gateway, error) {
	return client.Connect(id, client.WithClientConnection(connection), client.WithHash(self[[]byte]), client.WithSign(id.Sign))
}

func Submit(ctx context.Context, proposal *client.Proposal) ([]byte, error) {
	transaction, err := proposal.EndorseWithContext(ctx)
	if err != nil {
		return nil, err
	}

	commit, err := transaction.SubmitWithContext(ctx)
	if err != nil {
		return nil, err
	}

	status, err := commit.StatusWithContext(ctx)
	if err != nil {
		return nil, err
	}

	if !status.Successful {
		return nil, fmt.Errorf("transaction %s rejected with validation code %d (%s)", status.TransactionID, int32(status.Code), peer.TxValidationCode_name[int32(status.Code)])
	}

	return transaction.Result(), nil
}
