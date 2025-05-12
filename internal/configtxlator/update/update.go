/*
Copyright IBM Corp. 2017 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package update

import (
	"bytes"
	"errors"

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

		if originalPolicy.GetModPolicy() == updatedPolicy.GetModPolicy() && proto.Equal(originalPolicy.GetPolicy(), updatedPolicy.GetPolicy()) {
			sameSet[policyName] = &common.ConfigPolicy{
				Version: originalPolicy.GetVersion(),
			}
			continue
		}

		writeSet[policyName] = &common.ConfigPolicy{
			Version:   originalPolicy.GetVersion() + 1,
			ModPolicy: updatedPolicy.GetModPolicy(),
			Policy:    updatedPolicy.GetPolicy(),
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
			ModPolicy: updatedPolicy.GetModPolicy(),
			Policy:    updatedPolicy.GetPolicy(),
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

		if originalValue.GetModPolicy() == updatedValue.GetModPolicy() && bytes.Equal(originalValue.GetValue(), updatedValue.GetValue()) {
			sameSet[valueName] = &common.ConfigValue{
				Version: originalValue.GetVersion(),
			}
			continue
		}

		writeSet[valueName] = &common.ConfigValue{
			Version:   originalValue.GetVersion() + 1,
			ModPolicy: updatedValue.GetModPolicy(),
			Value:     updatedValue.GetValue(),
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
			ModPolicy: updatedValue.GetModPolicy(),
			Value:     updatedValue.GetValue(),
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
			ModPolicy: updatedGroup.GetModPolicy(),
			Policies:  groupWriteSet.GetPolicies(),
			Values:    groupWriteSet.GetValues(),
			Groups:    groupWriteSet.GetGroups(),
		}
	}

	return
}

func computeGroupUpdate(original, updated *common.ConfigGroup) (readSet, writeSet *common.ConfigGroup, updatedGroup bool) {
	readSetPolicies, writeSetPolicies, sameSetPolicies, policiesMembersUpdated := computePoliciesMapUpdate(original.GetPolicies(), updated.GetPolicies())
	readSetValues, writeSetValues, sameSetValues, valuesMembersUpdated := computeValuesMapUpdate(original.GetValues(), updated.GetValues())
	readSetGroups, writeSetGroups, sameSetGroups, groupsMembersUpdated := computeGroupsMapUpdate(original.GetGroups(), updated.GetGroups())

	// If the updated group is 'Equal' to the updated group (none of the members nor the mod policy changed)
	if !policiesMembersUpdated && !valuesMembersUpdated && !groupsMembersUpdated && original.GetModPolicy() == updated.GetModPolicy() {

		// If there were no modified entries in any of the policies/values/groups maps
		if len(readSetPolicies)+len(writeSetPolicies)+len(readSetValues)+len(writeSetValues)+len(readSetGroups)+len(writeSetGroups) == 0 {
			return &common.ConfigGroup{
					Version: original.GetVersion(),
				}, &common.ConfigGroup{
					Version: original.GetVersion(),
				}, false
		}

		return &common.ConfigGroup{
				Version:  original.GetVersion(),
				Policies: readSetPolicies,
				Values:   readSetValues,
				Groups:   readSetGroups,
			}, &common.ConfigGroup{
				Version:  original.GetVersion(),
				Policies: writeSetPolicies,
				Values:   writeSetValues,
				Groups:   writeSetGroups,
			}, true
	}

	copyMap(sameSetPolicies, readSetPolicies, writeSetPolicies)
	copyMap(sameSetValues, readSetValues, writeSetValues)
	copyMap(sameSetGroups, readSetGroups, writeSetGroups)

	return &common.ConfigGroup{
			Version:  original.GetVersion(),
			Policies: readSetPolicies,
			Values:   readSetValues,
			Groups:   readSetGroups,
		}, &common.ConfigGroup{
			Version:   original.GetVersion() + 1,
			Policies:  writeSetPolicies,
			Values:    writeSetValues,
			Groups:    writeSetGroups,
			ModPolicy: updated.GetModPolicy(),
		}, true
}

func copyMap[K comparable, V any](source map[K]V, targets ...map[K]V) {
	for key, value := range source {
		for _, target := range targets {
			target[key] = value
		}
	}
}

func Compute(original, updated *common.Config) (*common.ConfigUpdate, error) {
	if original.GetChannelGroup() == nil {
		return nil, errors.New("no channel group included for original config")
	}

	if updated.GetChannelGroup() == nil {
		return nil, errors.New("no channel group included for updated config")
	}

	readSet, writeSet, groupUpdated := computeGroupUpdate(original.GetChannelGroup(), updated.GetChannelGroup())
	if !groupUpdated {
		return nil, errors.New("no differences detected between original and updated config")
	}
	return &common.ConfigUpdate{
		ReadSet:  readSet,
		WriteSet: writeSet,
	}, nil
}
