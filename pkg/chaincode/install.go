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

const installTransactionName = "InstallChaincode"

// Install a chaincode package to specific peer.
func Install(ctx context.Context, connection grpc.ClientConnInterface, signingID identity.SigningIdentity, packageReader io.Reader) error {
	packageBytes, err := io.ReadAll(packageReader)
	if err != nil {
		return fmt.Errorf("failed to read chaincode package: %w", err)
	}

	installArgs := &lifecycle.InstallChaincodeArgs{
		ChaincodeInstallPackage: packageBytes,
	}
	installArgsBytes, err := proto.Marshal(installArgs)
	if err != nil {
		return err
	}

	proposalProto, err := proposal.NewProposal(signingID, lifecycleChaincodeName, installTransactionName, proposal.WithArguments(installArgsBytes))
	if err != nil {
		return err
	}

	signedProposal, err := proposal.NewSignedProposal(proposalProto, signingID)
	if err != nil {
		return err
	}

	endorser := peer.NewEndorserClient(connection)

	proposalResponse, err := endorser.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return fmt.Errorf("failed to install chaincode: %w", err)
	}

	return proposal.CheckSuccessfulResponse(proposalResponse)
}
