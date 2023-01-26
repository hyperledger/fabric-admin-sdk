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

const approveTransactionName = "ApproveChaincodeDefinitionForMyOrg"

type Definition struct {
	Name                 string
	Version              string
	Sequence             int64
	EndorsementPolicy    string
	ValidationPlugin     string
	ValidationParameters []byte
	PackageID            string
}

// Install a chaincode package to specific peer.
func Approve(ctx context.Context, connection grpc.ClientConnInterface, id identity.SigningIdentity, channelName string, chaincodeDef *Definition) error {
	approveArgs := &lifecycle.ApproveChaincodeDefinitionForMyOrgArgs{
		Name:                chaincodeDef.Name,
		Version:             chaincodeDef.Version,
		Sequence:            chaincodeDef.Sequence,
		EndorsementPlugin:   chaincodeDef.EndorsementPolicy,
		ValidationPlugin:    chaincodeDef.ValidationPlugin,
		ValidationParameter: chaincodeDef.ValidationParameters,
		Source:              newChaincodeSource(chaincodeDef.PackageID),
	}
	approveArgsBytes, err := proto.Marshal(approveArgs)
	if err != nil {
		return err
	}

	gw, err := gateway.New(connection, id)
	if err != nil {
		return err
	}
	defer gw.Close()

	approveProposal, err := gw.GetNetwork(channelName).
		GetContract(lifecycleChaincodeName).
		NewProposal(approveTransactionName, client.WithBytesArguments(approveArgsBytes), client.WithEndorsingOrganizations(id.MspID()))
	if err != nil {
		return err
	}

	_, err = gateway.Submit(ctx, approveProposal)
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
