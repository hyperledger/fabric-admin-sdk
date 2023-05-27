/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"
	"fmt"
	"io"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/internal/proposal"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// Install a chaincode package to specific peer. The connection must be to the specific peer where the chaincode is to
// be installed.
func Install(ctx context.Context, connection grpc.ClientConnInterface, id identity.SigningIdentity, packageReader io.Reader) (*lifecycle.InstallChaincodeResult, error) {
	packageBytes, err := io.ReadAll(packageReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read chaincode package: %w", err)
	}

	installArgs := &lifecycle.InstallChaincodeArgs{
		ChaincodeInstallPackage: packageBytes,
	}
	installArgsBytes, err := proto.Marshal(installArgs)
	if err != nil {
		return nil, err
	}

	proposalProto, err := proposal.NewProposal(id, lifecycleChaincodeName, installTransactionName, proposal.WithArguments(installArgsBytes))
	if err != nil {
		return nil, err
	}

	signedProposal, err := proposal.NewSignedProposal(proposalProto, id)
	if err != nil {
		return nil, err
	}

	endorser := peer.NewEndorserClient(connection)

	proposalResponse, err := endorser.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return nil, fmt.Errorf("failed to install chaincode: %w", err)
	}

	if err := proposal.CheckSuccessfulResponse(proposalResponse); err != nil {
		return nil, err
	}

	result := &lifecycle.InstallChaincodeResult{}
	if err := proto.Unmarshal(proposalResponse.GetResponse().GetPayload(), result); err != nil {
		return nil, fmt.Errorf("failed to deserialize install chaincode result: %w", err)
	}

	return result, nil
}
