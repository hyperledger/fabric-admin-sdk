package chaincode

import (
	"context"
	"fabric-admin-sdk/internal/pkg/identity"
	"fabric-admin-sdk/pkg/internal/proposal"
	"fmt"
	"io"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const (
	lifecycleChaincodeName = "_lifecycle"
	installTransactionName = "InstallChaincode"
)

// Install a chaincode package to specific peer.
func Install(ctx context.Context, connection grpc.ClientConnInterface, signer identity.SignerSerializer, packageReader io.Reader) error {
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

	proposalProto, err := proposal.NewProposal(signer, lifecycleChaincodeName, installTransactionName, proposal.WithArguments(installArgsBytes))
	if err != nil {
		return err
	}

	signedProposal, err := proposal.NewSignedProposal(proposalProto, signer)
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
