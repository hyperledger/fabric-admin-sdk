package chaincode

import (
	"context"
	"fabric-admin-sdk/internal/pkg/identity"
	"fabric-admin-sdk/pkg/internal/proposal"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
)

const (
	lifecycleChaincodeName = "_lifecycle"
	installTransactionName = "InstallChaincode"
)

func Install(ctx context.Context, endorser peer.EndorserClient, signer identity.SignerSerializer, packageReader io.Reader) error {
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

	proposalResponse, err := endorser.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return fmt.Errorf("failed to install chaincode: %w", err)
	}

	return proposal.CheckSuccessfulResponse(proposalResponse)
}
