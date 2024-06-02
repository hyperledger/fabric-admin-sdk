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

// Peer in a Fabric network.
type Peer struct {
	endorser peer.EndorserClient
	id       identity.SigningIdentity
}

// NewPeer creates a new Peer instance.
func NewPeer(connection grpc.ClientConnInterface, id identity.SigningIdentity) *Peer {
	return &Peer{
		endorser: peer.NewEndorserClient(connection),
		id:       id,
	}
}

// ClientIdentity used to interact with the peer.
func (p *Peer) ClientIdentity() identity.SigningIdentity {
	return p.id
}

func (p *Peer) newSignedProposal(
	chaincodeName string,
	transactionName string,
	options ...proposal.Option,
) (*peer.SignedProposal, error) {
	proposalProto, err := proposal.NewProposal(p.id, chaincodeName, transactionName, options...)
	if err != nil {
		return nil, err
	}

	signedProposal, err := proposal.NewSignedProposal(proposalProto, p.id)
	if err != nil {
		return nil, err
	}

	return signedProposal, nil
}

// GetInstalled chaincode package from a specific peer.
func (p *Peer) GetInstalled(ctx context.Context, packageID string) ([]byte, error) {
	getInstalledArgs := &lifecycle.GetInstalledChaincodePackageArgs{
		PackageId: packageID,
	}
	getInstalledArgsBytes, err := proto.Marshal(getInstalledArgs)
	if err != nil {
		return nil, err
	}

	signedProposal, err := p.newSignedProposal(lifecycleChaincodeName, queryInstalledTransactionName, proposal.WithArguments(getInstalledArgsBytes))
	if err != nil {
		return nil, err
	}

	proposalResponse, err := p.endorser.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed chaincode: %w", err)
	}

	if err = proposal.CheckSuccessfulResponse(proposalResponse); err != nil {
		return nil, err
	}

	result := &lifecycle.GetInstalledChaincodePackageResult{}
	if err = proto.Unmarshal(proposalResponse.GetResponse().GetPayload(), result); err != nil {
		return nil, fmt.Errorf("failed to deserialize get installed chaincode result: %w", err)
	}

	return result.GetChaincodeInstallPackage(), nil
}

// Install a chaincode package to specific peer.
func (p *Peer) Install(ctx context.Context, packageReader io.Reader) (*lifecycle.InstallChaincodeResult, error) {
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

	signedProposal, err := p.newSignedProposal(lifecycleChaincodeName, installTransactionName, proposal.WithArguments(installArgsBytes))
	if err != nil {
		return nil, err
	}

	proposalResponse, err := p.endorser.ProcessProposal(ctx, signedProposal)
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

// QueryInstalled chaincode on a specific peer.
func (p *Peer) QueryInstalled(ctx context.Context) (*lifecycle.QueryInstalledChaincodesResult, error) {
	queryArgs := &lifecycle.QueryInstalledChaincodesArgs{}
	queryArgsBytes, err := proto.Marshal(queryArgs)
	if err != nil {
		return nil, err
	}

	signedProposal, err := p.newSignedProposal(lifecycleChaincodeName, queryInstalledTransactionName, proposal.WithArguments(queryArgsBytes))
	if err != nil {
		return nil, err
	}

	proposalResponse, err := p.endorser.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return nil, fmt.Errorf("failed to query installed chaincode: %w", err)
	}

	if err = proposal.CheckSuccessfulResponse(proposalResponse); err != nil {
		return nil, err
	}

	result := &lifecycle.QueryInstalledChaincodesResult{}
	if err = proto.Unmarshal(proposalResponse.GetResponse().GetPayload(), result); err != nil {
		return nil, fmt.Errorf("failed to deserialize query installed chaincode result: %w", err)
	}

	return result, nil
}
