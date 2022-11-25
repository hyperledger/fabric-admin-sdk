/*
Copyright IBM Corp. 2017 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package update

import (
	"bytes"
	"fmt"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"google.golang.org/protobuf/proto"
)

func computePoliciesMapUpdate(original, updated map[string]*common.ConfigPolicy) (readSet, writeSet, sameSet map[string]*common.ConfigPolicy, updatedMembers bool) {
	readSet = make(map[string]*common.ConfigPolicy)
	writeSet = make(map[string]*common.ConfigPolicy)

	// All modified config goes into the read/write sets, but in case the map membership changes, we retain the
	// config which was the same to add to the read/write sets
	sameSet = make(map[string]*common.ConfigPolicy)

	for policyName, originalPolicy := range original {
		updatedPolicy, ok := updated[policyName]
		if !ok {
			updatedMembers = true
			continue
		}

		if originalPolicy.ModPolicy == updatedPolicy.ModPolicy && proto.Equal(originalPolicy.Policy, updatedPolicy.Policy) {
			sameSet[policyName] = &common.ConfigPolicy{
				Version: originalPolicy.Version,
			}
			continue
		}

		writeSet[policyName] = &common.ConfigPolicy{
			Version:   originalPolicy.Version + 1,
			ModPolicy: updatedPolicy.ModPolicy,
			Policy:    updatedPolicy.Policy,
		}
	}

	for policyName, updatedPolicy := range updated {
		if _, ok := original[policyName]; ok {
			// If the updatedPolicy is in the original set of policies, it was already handled
			continue
		}
		updatedMembers = true
		writeSet[policyName] = &common.ConfigPolicy{
			Version:   0,
			ModPolicy: updatedPolicy.ModPolicy,
			Policy:    updatedPolicy.Policy,
		}
	}

	return
}

func computeValuesMapUpdate(original, updated map[string]*common.ConfigValue) (readSet, writeSet, sameSet map[string]*common.ConfigValue, updatedMembers bool) {
	readSet = make(map[string]*common.ConfigValue)
	writeSet = make(map[string]*common.ConfigValue)

	// All modified config goes into the read/write sets, but in case the map membership changes, we retain the
	// config which was the same to add to the read/write sets
	sameSet = make(map[string]*common.ConfigValue)

	for valueName, originalValue := range original {
		updatedValue, ok := updated[valueName]
		if !ok {
			updatedMembers = true
			continue
		}

		if originalValue.ModPolicy == updatedValue.ModPolicy && bytes.Equal(originalValue.Value, updatedValue.Value) {
			sameSet[valueName] = &common.ConfigValue{
				Version: originalValue.Version,
			}
			continue
		}

		writeSet[valueName] = &common.ConfigValue{
			Version:   originalValue.Version + 1,
			ModPolicy: updatedValue.ModPolicy,
			Value:     updatedValue.Value,
		}
	}

	for valueName, updatedValue := range updated {
		if _, ok := original[valueName]; ok {
			// If the updatedValue is in the original set of values, it was already handled
			continue
		}
		updatedMembers = true
		writeSet[valueName] = &common.ConfigValue{
			Version:   0,
			ModPolicy: updatedValue.ModPolicy,
			Value:     updatedValue.Value,
		}
	}

	return
}

func computeGroupsMapUpdate(original, updated map[string]*common.ConfigGroup) (readSet, writeSet, sameSet map[string]*common.ConfigGroup, updatedMembers bool) {
	readSet = make(map[string]*common.ConfigGroup)
	writeSet = make(map[string]*common.ConfigGroup)

	// All modified config goes into the read/write sets, but in case the map membership changes, we retain the
	// config which was the same to add to the read/write sets
	sameSet = make(map[string]*common.ConfigGroup)

	for groupName, originalGroup := range original {
		updatedGroup, ok := updated[groupName]
		if !ok {
			updatedMembers = true
			continue
		}

		groupReadSet, groupWriteSet, groupUpdated := computeGroupUpdate(originalGroup, updatedGroup)
		if !groupUpdated {
			sameSet[groupName] = groupReadSet
			continue
		}

		readSet[groupName] = groupReadSet
		writeSet[groupName] = groupWriteSet

	}

	for groupName, updatedGroup := range updated {
		if _, ok := original[groupName]; ok {
			// If the updatedGroup is in the original set of groups, it was already handled
			continue
		}
		updatedMembers = true
		configGroup := &common.ConfigGroup{
			Groups:   make(map[string]*common.ConfigGroup),
			Values:   make(map[string]*common.ConfigValue),
			Policies: make(map[string]*common.ConfigPolicy),
		}
		_, groupWriteSet, _ := computeGroupUpdate(configGroup, updatedGroup)
		writeSet[groupName] = &common.ConfigGroup{
			Version:   0,
			ModPolicy: updatedGroup.ModPolicy,
			Policies:  groupWriteSet.Policies,
			Values:    groupWriteSet.Values,
			Groups:    groupWriteSet.Groups,
		}
	}

	return
}

func computeGroupUpdate(original, updated *common.ConfigGroup) (readSet, writeSet *common.ConfigGroup, updatedGroup bool) {
	readSetPolicies, writeSetPolicies, sameSetPolicies, policiesMembersUpdated := computePoliciesMapUpdate(original.Policies, updated.Policies)
	readSetValues, writeSetValues, sameSetValues, valuesMembersUpdated := computeValuesMapUpdate(original.Values, updated.Values)
	readSetGroups, writeSetGroups, sameSetGroups, groupsMembersUpdated := computeGroupsMapUpdate(original.Groups, updated.Groups)

	// If the updated group is 'Equal' to the updated group (none of the members nor the mod policy changed)
	if !(policiesMembersUpdated || valuesMembersUpdated || groupsMembersUpdated || original.ModPolicy != updated.ModPolicy) {

		// If there were no modified entries in any of the policies/values/groups maps
		if len(readSetPolicies) == 0 &&
			len(writeSetPolicies) == 0 &&
			len(readSetValues) == 0 &&
			len(writeSetValues) == 0 &&
			len(readSetGroups) == 0 &&
			len(writeSetGroups) == 0 {
			return &common.ConfigGroup{
					Version: original.Version,
				}, &common.ConfigGroup{
					Version: original.Version,
				}, false
		}

		return &common.ConfigGroup{
				Version:  original.Version,
				Policies: readSetPolicies,
				Values:   readSetValues,
				Groups:   readSetGroups,
			}, &common.ConfigGroup{
				Version:  original.Version,
				Policies: writeSetPolicies,
				Values:   writeSetValues,
				Groups:   writeSetGroups,
			}, true
	}

	for k, samePolicy := range sameSetPolicies {
		readSetPolicies[k] = samePolicy
		writeSetPolicies[k] = samePolicy
	}

	for k, sameValue := range sameSetValues {
		readSetValues[k] = sameValue
		writeSetValues[k] = sameValue
	}

	for k, sameGroup := range sameSetGroups {
		readSetGroups[k] = sameGroup
		writeSetGroups[k] = sameGroup
	}

	return &common.ConfigGroup{
			Version:  original.Version,
			Policies: readSetPolicies,
			Values:   readSetValues,
			Groups:   readSetGroups,
		}, &common.ConfigGroup{
			Version:   original.Version + 1,
			Policies:  writeSetPolicies,
			Values:    writeSetValues,
			Groups:    writeSetGroups,
			ModPolicy: updated.ModPolicy,
		}, true
}

func Compute(original, updated *common.Config) (*common.ConfigUpdate, error) {
	if original.ChannelGroup == nil {
		return nil, fmt.Errorf("no channel group included for original config")
	}

	if updated.ChannelGroup == nil {
		return nil, fmt.Errorf("no channel group included for updated config")
	}

	readSet, writeSet, groupUpdated := computeGroupUpdate(original.ChannelGroup, updated.ChannelGroup)
	if !groupUpdated {
		return nil, fmt.Errorf("no differences detected between original and updated config")
	}
	return &common.ConfigUpdate{
		ReadSet:  readSet,
		WriteSet: writeSet,
	}, nil
}
