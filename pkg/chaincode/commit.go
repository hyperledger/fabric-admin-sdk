package chaincode

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/internal/gateway"
	"github.com/hyperledger/fabric-gateway/pkg/client"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// Commit a chaincode package to specific peer.
func Commit(ctx context.Context, connection grpc.ClientConnInterface, id identity.SigningIdentity, chaincodeDef *Definition) error {
	err := chaincodeDef.Validate()
	if err != nil {
		return err
	}
	commitArgs := &lifecycle.CommitChaincodeDefinitionArgs{
		Name:                chaincodeDef.Name,
		Version:             chaincodeDef.Version,
		Sequence:            chaincodeDef.Sequence,
		EndorsementPlugin:   chaincodeDef.EndorsementPlugin,
		ValidationPlugin:    chaincodeDef.ValidationPlugin,
		ValidationParameter: chaincodeDef.ValidationParameter,
		Collections:         chaincodeDef.Collections,
		InitRequired:        chaincodeDef.InitRequired,
	}
	commitArgsBytes, err := proto.Marshal(commitArgs)
	if err != nil {
		return err
	}

	gw, err := gateway.New(connection, id)
	if err != nil {
		return err
	}
	defer gw.Close()

	_, err = gw.GetNetwork(chaincodeDef.ChannelName).
		GetContract(lifecycleChaincodeName).
		SubmitWithContext(
			ctx,
			commitTransactionName,
			client.WithBytesArguments(commitArgsBytes),
		)
	if err != nil {
		return fmt.Errorf("failed to commit chaincode: %w", err)
	}

	return nil
}
