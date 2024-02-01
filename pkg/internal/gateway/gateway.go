/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"google.golang.org/grpc"
)

func New(connection grpc.ClientConnInterface, id identity.SigningIdentity) (*client.Gateway, error) {
	return client.Connect(id, client.WithClientConnection(connection), client.WithHash(hash.NONE), client.WithSign(id.Sign))
}
