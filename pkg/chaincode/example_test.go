/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode_test

import (
	"context"
	"crypto/x509"
	"os"
	"time"

	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const peerEndpoint = "peer.example.org:7051"
const mspID = "Org1"
const chaincodePackageFile = "basic.tar.gz"

func Example() {
	// gRPC connection a target peer.
	connection := newGrpcConnection()
	defer connection.Close()

	certificate, err := identity.ReadCertificate("client-certificate.pem")
	panicOnError(err)
	privateKey, err := identity.ReadPrivateKey("client-private-key.pem")
	panicOnError(err)

	// Client identity used to carry out deployment tasks.
	id, err := identity.NewPrivateKeySigningIdentity(mspID, certificate, privateKey)
	panicOnError(err)

	peer := chaincode.NewPeer(connection, id)
	gateway := chaincode.NewGateway(connection, id)

	// Context used to manage Fabric invocations.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Read existing chaincode package file.
	chaincodePackage, err := os.Open(chaincodePackageFile)
	panicOnError(err)

	// Install chaincode package. This must be performed for each peer on which
	// the chaincode is to be installed.
	_, err = peer.Install(ctx, chaincodePackage)
	panicOnError(err)

	// Definition of the chaincode as it should appear on the channel.
	chaincodeDefinition := &chaincode.Definition{
		ChannelName: "mychannel",
		Name:        "basic",
		Version:     "1.0",
		Sequence:    1,
	}

	// Approve chaincode definition. This must be performed using client
	// identities from sufficient organizations to satisfy the approval policy.
	err = gateway.Approve(ctx, chaincodeDefinition)
	panicOnError(err)

	// Commit approved chaincode definition. This can be carried out by any
	// organization once enough approvals have been recorded.
	err = gateway.Commit(ctx, chaincodeDefinition)
	panicOnError(err)
}

func newGrpcConnection() *grpc.ClientConn {
	caCertificate, err := identity.ReadCertificate("ca-certificate.pem")
	panicOnError(err)

	certPool := x509.NewCertPool()
	certPool.AddCert(caCertificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "")

	connection, err := grpc.NewClient(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	panicOnError(err)

	return connection
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
