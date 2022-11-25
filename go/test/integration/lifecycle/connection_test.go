/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package lifecycle_test

import (
	"fabric-admin-sdk/internal/pkg/identity"
	"fabric-admin-sdk/pkg/tools"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	mspID         = "Org1MSP"
	clientCertEnv = "CLIENT_CERT"
	clientKeyEnv  = "CLIENT_KEY"
	caCertEnv     = "CA_CERT"
	endpointEnv   = "ENDPOINT"
)

func newGrpcConnection() (*grpc.ClientConn, error) {
	transportCredentials := insecure.NewCredentials()
	return grpc.Dial(os.Getenv(endpointEnv), grpc.WithTransportCredentials(transportCredentials))
}

func newSignerSerializer() (identity.SignerSerializer, error) {
	return tools.CreateSigner(os.Getenv(clientKeyEnv), os.Getenv(clientCertEnv), mspID)
}
