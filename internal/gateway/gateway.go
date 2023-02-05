/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"google.golang.org/grpc"
)

func self[T any](v T) T {
	return v
}

func New(connection grpc.ClientConnInterface, id identity.SigningIdentity) (*client.Gateway, error) {
	return client.Connect(id, client.WithClientConnection(connection), client.WithHash(self[[]byte]), client.WithSign(id.Sign))
}
