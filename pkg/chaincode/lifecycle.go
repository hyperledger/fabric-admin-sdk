/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package chaincode provides functions for driving chaincode lifecycle.
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
	queryApprovedTransactionName          = "QueryApprovedChaincodeDefinition"
	queryCommittedTransactionName         = "QueryChaincodeDefinitions"
	queryCommittedWithNameTransactionName = "QueryChaincodeDefinition"
	checkCommitReadinessTransactionName   = "CheckCommitReadiness"
	installTransactionName                = "InstallChaincode"
	getInstalledTransactionName           = "GetInstalledChaincodePackage"

	// metadataFile is the expected location of the metadata json document
	// in the top level of the chaincode package.
	metadataFile = "metadata.json"

	// codePackageFile is the expected location of the code package in the
	// top level of the chaincode package.
	codePackageFile = "code.tar.gz"
)

// Definition of a chaincode.
type Definition struct {
	// ChannelName on which the chaincode is deployed.
	ChannelName string

	// PackageID is a unique identifier for a chaincode package, combining the package label with a hash of the package.
	PackageID string

	// Name used when invoking the chaincode.
	Name string

	// Version associated with a given chaincode package.
	Version string

	// EndorsementPlugin used by the chaincode. May be omitted unless a custom plugin is required.
	EndorsementPlugin string

	// ValidationPlugin used by the chaincode. May be omitted unless a custom plugin is required.
	ValidationPlugin string

	// Sequence number indicating the number of times the chaincode has been defined on a channel, and used to keep
	// track of chaincode upgrades.
	Sequence int64

	// ValidationParameter defines the endorsement policy for the chaincode. This can be an explicit endorsement policy
	// string or reference a policy in the channel configuration.
	ValidationParameter []byte

	// InitRequired is true only if the chaincode defines an Init function using the low-level shim API, which must be
	// invoked before other transaction functions may be invoked; otherwise false. It is not recommended to rely on
	// functionality.
	InitRequired bool

	// Collections configuration for private data collections accessed by the chaincode.
	Collections *peer.CollectionConfigPackage
}

func (d *Definition) validate() error {
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
