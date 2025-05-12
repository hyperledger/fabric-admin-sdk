/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package chaincode provides functions for driving chaincode lifecycle.
package chaincode

import (
	"errors"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
)

const (
	lifecycleChaincodeName                = "_lifecycle"                         // #nosec G101
	approveTransactionName                = "ApproveChaincodeDefinitionForMyOrg" // #nosec G101
	commitTransactionName                 = "CommitChaincodeDefinition"          // #nosec G101
	queryInstalledTransactionName         = "QueryInstalledChaincodes"           // #nosec G101
	queryApprovedTransactionName          = "QueryApprovedChaincodeDefinition"   // #nosec G101
	queryCommittedTransactionName         = "QueryChaincodeDefinitions"          // #nosec G101
	queryCommittedWithNameTransactionName = "QueryChaincodeDefinition"           // #nosec G101
	checkCommitReadinessTransactionName   = "CheckCommitReadiness"               // #nosec G101
	installTransactionName                = "InstallChaincode"                   // #nosec G101
	getInstalledTransactionName           = "GetInstalledChaincodePackage"       // #nosec G101

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

	// ApplicationPolicy defines the endorsement policy for the chaincode. This can be an explicit endorsement policy
	// it follows fabric-protos-go-apiv2/peer format and it will convert to validationParameter during validation phase.
	ApplicationPolicy *peer.ApplicationPolicy

	// InitRequired is true only if the chaincode defines an Init function using the low-level shim API, which must be
	// invoked before other transaction functions may be invoked; otherwise false. It is not recommended to rely on
	// functionality.
	InitRequired bool

	// Collections configuration for private data collections accessed by the chaincode.
	Collections *peer.CollectionConfigPackage
}

func (d *Definition) validate() error {
	if d.ChannelName == "" {
		return errors.New("channel name is required for channel approve/commit")
	}
	if d.Name == "" {
		return errors.New("chaincode name is required for channel approve/commit")
	}
	if d.Version == "" {
		return errors.New("chaincode version is required for channel approve/commit")
	}
	if d.Sequence <= 0 {
		return errors.New("chaincode sequence must be greater than 0 for channel approve/commit")
	}
	return nil
}

// getApplicationPolicyBytes is to convert ApplicationPolicy to Bytes for proto usage
func (d *Definition) getApplicationPolicyBytes() ([]byte, error) {
	if d.ApplicationPolicy != nil {
		AppPolicydata, err := proto.Marshal(d.ApplicationPolicy)
		if err != nil {
			return nil, err
		}
		return AppPolicydata, nil
	}
	return nil, nil
}
