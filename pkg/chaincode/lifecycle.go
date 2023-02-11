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

// Chaincode Define
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
		return fmt.Errorf("For channel approve/commit channel name is required")
	}
	if d.Name == "" {
		return fmt.Errorf("For channel approve/commit chaincode name is required")
	}
	if d.Version == "" {
		return fmt.Errorf("For channel approve/commit chaincode version is required")
	}
	if d.Sequence == 0 {
		return fmt.Errorf("For channel approve/commit chaincode Sequence is required as bigger than 0")
	}
	return nil
}
