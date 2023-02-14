/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

const (
	lifecycleChaincodeName                = "_lifecycle"
	approveTransactionName                = "ApproveChaincodeDefinitionForMyOrg"
	commitTransactionName                 = "CommitChaincodeDefinition"
	queryInstalledTransactionName         = "QueryInstalledChaincodes"
	queryCommittedTransactionName         = "QueryChaincodeDefinitions"
	queryCommittedWithNameTransactionName = "QueryChaincodeDefinition"
	checkCommitReadinessTransactionName   = "CheckCommitReadiness"
	installTransactionName                = "InstallChaincode"
	// MetadataFile is the expected location of the metadata json document
	// in the top level of the chaincode package.
	MetadataFile = "metadata.json"

	// CodePackageFile is the expected location of the code package in the
	// top level of the chaincode package
	CodePackageFile = "code.tar.gz"
)

// Definition of a chaincode
type Definition struct {
	ChannelName         string
	PackageID           string
	Name                string
	Version             string
	EndorsementPlugin   string
	ValidationPlugin    string
	Sequence            int64
	ValidationParameter []byte
	InitRequired        bool
	Collections         *peer.CollectionConfigPackage
}

func (d *Definition) Validate() error {
	if d.ChannelName == "" {
		return fmt.Errorf("channel name is required for channel approve/commit")
	}
	if d.Name == "" {
		return fmt.Errorf("chaincode name is required for channel approve/commit")
	}
	if d.Version == "" {
		return fmt.Errorf("chaincode version is required for channel approve/commit")
	}
	if d.Sequence <= 0 {
		return fmt.Errorf("chaincode sequence must be greater than 0 for channel approve/commit")
	}
	return nil
}
