/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

// collectionConfigJSON describes a single private data collection. The format
// matches the collection configuration file accepted by the Fabric peer CLI,
// so existing collections configuration files can be reused as-is.
type collectionConfigJSON struct {
	Name              string                 `json:"name"`
	Policy            string                 `json:"policy"`
	RequiredPeerCount *int32                 `json:"requiredPeerCount"`
	MaxPeerCount      *int32                 `json:"maxPeerCount"`
	BlockToLive       uint64                 `json:"blockToLive"`
	MemberOnlyRead    bool                   `json:"memberOnlyRead"`
	MemberOnlyWrite   bool                   `json:"memberOnlyWrite"`
	EndorsementPolicy *endorsementPolicyJSON `json:"endorsementPolicy,omitempty"`
}

// endorsementPolicyJSON describes the optional endorsement policy of a
// collection. Either a signature policy string or a channel config policy
// reference may be given, but not both.
type endorsementPolicyJSON struct {
	SignaturePolicy     string `json:"signaturePolicy"`
	ChannelConfigPolicy string `json:"channelConfigPolicy"`
}

// ParseCollectionConfig takes a collection configuration string, using the
// same JSON format as the Fabric peer CLI collections configuration file,
// and converts it into a CollectionConfigPackage that can be used directly
// as ChaincodeDefinition.Collections.
//
// For example:
//
//	collections, err := chaincode.ParseCollectionConfig(`[
//		{
//			"name": "collectionMarbles",
//			"policy": "OR('Org1MSP.member', 'Org2MSP.member')",
//			"requiredPeerCount": 0,
//			"maxPeerCount": 3,
//			"blockToLive": 1000000,
//			"memberOnlyRead": true
//		}
//	]`)
func ParseCollectionConfig(configString string) (*peer.CollectionConfigPackage, error) {
	var configs []collectionConfigJSON
	if err := json.Unmarshal([]byte(configString), &configs); err != nil {
		return nil, fmt.Errorf("error parsing collection configuration: %w", err)
	}

	collectionConfigs := make([]*peer.CollectionConfig, 0, len(configs))
	for _, config := range configs {
		memberOrgsPolicy, err := signaturePolicyEnvelopeFromString(config.Policy)
		if err != nil {
			return nil, fmt.Errorf("error parsing member orgs policy for collection %s: %w", config.Name, err)
		}

		var endorsementPolicy *peer.ApplicationPolicy
		if config.EndorsementPolicy != nil {
			endorsementPolicy, err = NewApplicationPolicy(
				config.EndorsementPolicy.SignaturePolicy,
				config.EndorsementPolicy.ChannelConfigPolicy,
			)
			if err != nil {
				return nil, fmt.Errorf("error parsing endorsement policy for collection %s: %w", config.Name, err)
			}
		}

		// The required and maximum peer counts default to 0 and 1 when they
		// are not specified, matching the Fabric peer CLI behavior.
		var requiredPeerCount int32
		maxPeerCount := int32(1)
		if config.RequiredPeerCount != nil {
			requiredPeerCount = *config.RequiredPeerCount
		}
		if config.MaxPeerCount != nil {
			maxPeerCount = *config.MaxPeerCount
		}

		collectionConfigs = append(collectionConfigs, &peer.CollectionConfig{
			Payload: &peer.CollectionConfig_StaticCollectionConfig{
				StaticCollectionConfig: &peer.StaticCollectionConfig{
					Name: config.Name,
					MemberOrgsPolicy: &peer.CollectionPolicyConfig{
						Payload: &peer.CollectionPolicyConfig_SignaturePolicy{
							SignaturePolicy: memberOrgsPolicy,
						},
					},
					RequiredPeerCount: requiredPeerCount,
					MaximumPeerCount:  maxPeerCount,
					BlockToLive:       config.BlockToLive,
					MemberOnlyRead:    config.MemberOnlyRead,
					MemberOnlyWrite:   config.MemberOnlyWrite,
					EndorsementPolicy: endorsementPolicy,
				},
			},
		})
	}

	return &peer.CollectionConfigPackage{Config: collectionConfigs}, nil
}
