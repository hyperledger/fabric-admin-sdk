/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

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

// Gateway peer belonging to a specific organization to which you want to target requests.
type Gateway struct {
	connection grpc.ClientConnInterface
	id         identity.SigningIdentity
}

// NewGateway creates a new Gateway instance.
func NewGateway(connection grpc.ClientConnInterface, id identity.SigningIdentity) *Gateway {
	return &Gateway{
		connection: connection,
		id:         id,
	}
}

// ClientIdentity used to interact with the gateway peer.
func (g *Gateway) ClientIdentity() identity.SigningIdentity {
	return g.id
}

func (g *Gateway) lifecycleInvoke(
	channelName string,
	invocation func(contract *client.Contract) ([]byte, error),
) ([]byte, error) {
	fabricGateway, err := gateway.New(g.connection, g.id)
	if err != nil {
		return nil, err
	}
	defer fabricGateway.Close()

	contract := fabricGateway.GetNetwork(channelName).GetContract(lifecycleChaincodeName)
	return invocation(contract)
}

// Approve a chaincode package for the user's own organization.
func (g *Gateway) Approve(ctx context.Context, chaincodeDef *Definition) error {
	err := chaincodeDef.validate()
	if err != nil {
		return err
	}
	validationParameter, err := chaincodeDef.getApplicationPolicyBytes()
	if err != nil {
		return err
	}
	approveArgs := &lifecycle.ApproveChaincodeDefinitionForMyOrgArgs{
		Name:                chaincodeDef.Name,
		Version:             chaincodeDef.Version,
		Sequence:            chaincodeDef.Sequence,
		EndorsementPlugin:   chaincodeDef.EndorsementPlugin,
		ValidationPlugin:    chaincodeDef.ValidationPlugin,
		ValidationParameter: validationParameter,
		Collections:         chaincodeDef.Collections,
		InitRequired:        chaincodeDef.InitRequired,
		Source:              newChaincodeSource(chaincodeDef.PackageID),
	}
	approveArgsBytes, err := proto.Marshal(approveArgs)
	if err != nil {
		return err
	}

	_, err = g.lifecycleInvoke(chaincodeDef.ChannelName, func(contract *client.Contract) ([]byte, error) {
		return contract.SubmitWithContext(
			ctx,
			approveTransactionName,
			client.WithBytesArguments(approveArgsBytes),
			client.WithEndorsingOrganizations(g.id.MspID()),
		)
	})
	if err != nil {
		return fmt.Errorf("failed to approve chaincode: %w", err)
	}

	return nil
}

func newChaincodeSource(packageID string) *lifecycle.ChaincodeSource {
	switch packageID {
	case "":
		return &lifecycle.ChaincodeSource{
			Type: &lifecycle.ChaincodeSource_Unavailable_{
				Unavailable: &lifecycle.ChaincodeSource_Unavailable{},
			},
		}
	default:
		return &lifecycle.ChaincodeSource{
			Type: &lifecycle.ChaincodeSource_LocalPackage{
				LocalPackage: &lifecycle.ChaincodeSource_Local{
					PackageId: packageID,
				},
			},
		}
	}
}

// CheckCommitReadiness for a chaincode and return all approval records.
func (g *Gateway) CheckCommitReadiness(ctx context.Context, chaincodeDef *Definition) (*lifecycle.CheckCommitReadinessResult, error) {
	validationParameter, err := chaincodeDef.getApplicationPolicyBytes()
	if err != nil {
		return nil, err
	}
	args := &lifecycle.CheckCommitReadinessArgs{
		Name:                chaincodeDef.Name,
		Version:             chaincodeDef.Version,
		Sequence:            chaincodeDef.Sequence,
		EndorsementPlugin:   chaincodeDef.EndorsementPlugin,
		ValidationPlugin:    chaincodeDef.ValidationPlugin,
		ValidationParameter: validationParameter,
		Collections:         chaincodeDef.Collections,
		InitRequired:        chaincodeDef.InitRequired,
	}
	argsBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, err
	}

	resultBytes, err := g.lifecycleInvoke(chaincodeDef.ChannelName, func(contract *client.Contract) ([]byte, error) {
		return contract.EvaluateWithContext(
			ctx,
			checkCommitReadinessTransactionName,
			client.WithBytesArguments(argsBytes),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check commit readiness: %w", err)
	}

	result := &lifecycle.CheckCommitReadinessResult{}
	if err = proto.Unmarshal(resultBytes, result); err != nil {
		return nil, fmt.Errorf("failed to deserialize check commit readiness result: %w", err)
	}

	return result, nil
}

// Commit a chaincode definition to the channel.
func (g *Gateway) Commit(ctx context.Context, chaincodeDef *Definition) error {
	err := chaincodeDef.validate()
	if err != nil {
		return err
	}
	validationParameter, err := chaincodeDef.getApplicationPolicyBytes()
	if err != nil {
		return err
	}
	commitArgs := &lifecycle.CommitChaincodeDefinitionArgs{
		Name:                chaincodeDef.Name,
		Version:             chaincodeDef.Version,
		Sequence:            chaincodeDef.Sequence,
		EndorsementPlugin:   chaincodeDef.EndorsementPlugin,
		ValidationPlugin:    chaincodeDef.ValidationPlugin,
		ValidationParameter: validationParameter,
		Collections:         chaincodeDef.Collections,
		InitRequired:        chaincodeDef.InitRequired,
	}
	commitArgsBytes, err := proto.Marshal(commitArgs)
	if err != nil {
		return err
	}

	_, err = g.lifecycleInvoke(chaincodeDef.ChannelName, func(contract *client.Contract) ([]byte, error) {
		return contract.SubmitWithContext(
			ctx,
			commitTransactionName,
			client.WithBytesArguments(commitArgsBytes),
		)
	})
	if err != nil {
		return fmt.Errorf("failed to commit chaincode: %w", err)
	}

	return nil
}

// QueryApproved chaincode definition for the user's own organization.
//
// In Fabric v3, this function supports additional query modes:
//   - Omit the sequence number (pass 0) to query the latest approved chaincode definition.
//   - Omit the chaincode name (pass empty string) to query all approved chaincode definitions on the channel.
func (g *Gateway) QueryApproved(ctx context.Context, channelName string, chaincodeName string, sequence int64) (*lifecycle.QueryApprovedChaincodeDefinitionResult, error) {
	queryArgs := &lifecycle.QueryApprovedChaincodeDefinitionArgs{
        	if chaincodeName != "" {
		queryArgs.Name = chaincodeName
	}
	if sequence != 0 {
		queryArgs.Sequence = sequence
	}
		
	}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return nil, err
	}

	resultBytes, err := g.lifecycleInvoke(channelName, func(contract *client.Contract) ([]byte, error) {
		return contract.EvaluateWithContext(
			ctx,
			queryApprovedTransactionName,
			client.WithBytesArguments(queryArgsBytes),
			client.WithEndorsingOrganizations(g.id.MspID()),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query approved chaincode: %w", err)
	}

	result := &lifecycle.QueryApprovedChaincodeDefinitionResult{}
	if err = proto.Unmarshal(resultBytes, result); err != nil {
		return nil, fmt.Errorf("failed to deserialize query approved chaincode result: %w", err)
	}

	return result, nil
}

// QueryCommitted returns the definitions of all committed chaincode for a given channel.
func (g *Gateway) QueryCommitted(ctx context.Context, channelName string) (*lifecycle.QueryChaincodeDefinitionsResult, error) {
	queryArgs := &lifecycle.QueryChaincodeDefinitionsArgs{}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return nil, err
	}

	resultBytes, err := g.lifecycleInvoke(channelName, func(contract *client.Contract) ([]byte, error) {
		return contract.EvaluateWithContext(
			ctx,
			queryCommittedTransactionName,
			client.WithBytesArguments(queryArgsBytes),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query committed chaincodes: %w", err)
	}

	result := &lifecycle.QueryChaincodeDefinitionsResult{}
	if err = proto.Unmarshal(resultBytes, result); err != nil {
		return nil, fmt.Errorf("failed to deserialize query committed chaincode result: %w", err)
	}

	return result, nil
}

// QueryCommittedWithName returns the definition of the named chaincode for a given channel.
func (g *Gateway) QueryCommittedWithName(ctx context.Context, channelName string, chaincodeName string) (*lifecycle.QueryChaincodeDefinitionResult, error) {
	queryArgs := &lifecycle.QueryChaincodeDefinitionArgs{
		Name: chaincodeName,
	}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return nil, err
	}

	resultBytes, err := g.lifecycleInvoke(channelName, func(contract *client.Contract) ([]byte, error) {
		return contract.EvaluateWithContext(
			ctx,
			queryCommittedWithNameTransactionName,
			client.WithBytesArguments(queryArgsBytes),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query committed chaincode: %w", err)
	}

	result := &lifecycle.QueryChaincodeDefinitionResult{}
	if err = proto.Unmarshal(resultBytes, result); err != nil {
		return nil, fmt.Errorf("failed to deserialize query committed chaincode result: %w", err)
	}

	return result, nil
}
