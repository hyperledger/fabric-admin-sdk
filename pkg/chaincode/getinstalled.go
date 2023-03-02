/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/internal/proposal"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// GetInstalled chaincode package from a specific peer.
func GetInstalled(ctx context.Context, connection grpc.ClientConnInterface, signingID identity.SigningIdentity, packageID string) ([]byte, error) {
	getInstalledArgs := &lifecycle.GetInstalledChaincodePackageArgs{
		PackageId: packageID,
	}
	getInstalledArgsBytes, err := proto.Marshal(getInstalledArgs)
	if err != nil {
		return nil, err
	}

	proposalProto, err := proposal.NewProposal(signingID, lifecycleChaincodeName, queryInstalledTransactionName, proposal.WithArguments(getInstalledArgsBytes))
	if err != nil {
		return nil, err
	}

	signedProposal, err := proposal.NewSignedProposal(proposalProto, signingID)
	if err != nil {
		return nil, err
	}

	endorser := peer.NewEndorserClient(connection)

	proposalResponse, err := endorser.ProcessProposal(ctx, signedProposal)
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
