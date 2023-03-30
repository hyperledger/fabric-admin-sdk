/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode_test

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"os"
	"time"

	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const peerName = "peer.example.org"
const peerEndpoint = peerName + ":7051"
const mspID = "Org1"
const chaincodePackageFile = "basic.tar.gz"

func Example() {
	// gRPC connection a target peer.
	connection := newGrpcConnection()
	defer connection.Close()
	Endorser := peer.NewEndorserClient(connection)

	// Client identity used to carry out deployment tasks.
	id, err := identity.NewPrivateKeySigningIdentity(mspID, readCertificate(), readPrivateKey())
	panicOnError(err)

	// Context used to manage Fabric invocations.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Read existing chaincode package file.
	chaincodePackage, err := os.Open(chaincodePackageFile)
	panicOnError(err)

	// Install chaincode package. This must be performed for each peer on which the chaincode is to be installed.
	err = chaincode.Install(ctx, Endorser, id, chaincodePackage)
	panicOnError(err)

	Gateway, err := newGateway(connection, id)
	defer Gateway.Close()
	network := Gateway.GetNetwork("mychannel")

	// Definition of the chaincode as it should appear on the channel.
	chaincodeDefinition := &chaincode.Definition{
		ChannelName: "mychannel",
		Name:        "basic",
		Version:     "1.0",
		Sequence:    1,
	}

	// Approve chaincode definition. This must be performed using client identities from sufficient organizations to
	// satisfy the approval policy.
	err = chaincode.Approve(ctx, *network, id, chaincodeDefinition)
	panicOnError(err)

	// Commit approved chaincode definition. This can be carried out by any organization once enough approvals have
	// been recorded.
	err = chaincode.Commit(ctx, *network, chaincodeDefinition)
	panicOnError(err)
}

func newGrpcConnection() *grpc.ClientConn {
	caCertificate := readCertificate()
	certPool := x509.NewCertPool()
	certPool.AddCert(caCertificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "")

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	panicOnError(err)

	return connection
}

func self[T any](v T) T {
	return v
}

func newGateway(connection grpc.ClientConnInterface, id identity.SigningIdentity) (*client.Gateway, error) {
	return client.Connect(id, client.WithClientConnection(connection), client.WithHash(self[[]byte]), client.WithSign(id.Sign))
}

func readCertificate() *x509.Certificate {
	certificatePEM, err := os.ReadFile("certificate.pem")
	panicOnError(err)

	block, _ := pem.Decode([]byte(certificatePEM))
	if block == nil {
		panic("failed to parse certificate PEM")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}

	return certificate
}

func readPrivateKey() crypto.PrivateKey {
	privateKeyPEM, err := os.ReadFile("privateKey.pem")
	panicOnError(err)

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		panic("failed to parse private key PEM")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic("failed to parse PKCS8 encoded private key: " + err.Error())
	}

	return privateKey
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
