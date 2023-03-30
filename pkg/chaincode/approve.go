/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/client"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/protobuf/proto"
)

// Approve a chaincode package for the user's own organization. The connection may be to any Gateway peer that is a
// member of the channel.
func Approve(ctx context.Context, network client.Network, id identity.SigningIdentity, chaincodeDef *Definition) error {
	err := chaincodeDef.validate()
	if err != nil {
		return err
	}
	approveArgs := &lifecycle.ApproveChaincodeDefinitionForMyOrgArgs{
		Name:                chaincodeDef.Name,
		Version:             chaincodeDef.Version,
		Sequence:            chaincodeDef.Sequence,
		EndorsementPlugin:   chaincodeDef.EndorsementPlugin,
		ValidationPlugin:    chaincodeDef.ValidationPlugin,
		ValidationParameter: chaincodeDef.ValidationParameter,
		Collections:         chaincodeDef.Collections,
		InitRequired:        chaincodeDef.InitRequired,
		Source:              newChaincodeSource(chaincodeDef.PackageID),
	}
	approveArgsBytes, err := proto.Marshal(approveArgs)
	if err != nil {
		return err
	}

	_, err = network.GetContract(lifecycleChaincodeName).
		SubmitWithContext(
			ctx,
			approveTransactionName,
			client.WithBytesArguments(approveArgsBytes),
			client.WithEndorsingOrganizations(id.MspID()),
		)
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
