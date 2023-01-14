package policies

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"

	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
)

// ConfigPolicy defines a common representation for different *cb.ConfigPolicy values.
type ConfigPolicy interface {
	// Key is the key this value should be stored in the *cb.ConfigGroup.Policies map.
	Key() string

	// Value is the backing policy implementation for this ConfigPolicy
	Value() *cb.Policy
}

// StandardConfigValue implements the ConfigValue interface.
type StandardConfigPolicy struct {
	key   string
	value *cb.Policy
}

// Key is the key this value should be stored in the *cb.ConfigGroup.Values map.
func (scv *StandardConfigPolicy) Key() string {
	return scv.key
}

// Value is the *cb.Policy which should be stored as the *cb.ConfigPolicy.Policy.
func (scv *StandardConfigPolicy) Value() *cb.Policy {
	return scv.value
}

func makeImplicitMetaPolicy(subPolicyName string, rule cb.ImplicitMetaPolicy_Rule) *cb.Policy {
	return &cb.Policy{
		Type: int32(cb.Policy_IMPLICIT_META),
		Value: protoutil.MarshalOrPanic(&cb.ImplicitMetaPolicy{
			Rule:      rule,
			SubPolicy: subPolicyName,
		}),
	}
}

// ImplicitMetaAllPolicy defines an implicit meta policy whose sub_policy and key is policyname with rule ALL.
func ImplicitMetaAllPolicy(policyName string) *StandardConfigPolicy {
	return &StandardConfigPolicy{
		key:   policyName,
		value: makeImplicitMetaPolicy(policyName, cb.ImplicitMetaPolicy_ALL),
	}
}

// ImplicitMetaAnyPolicy defines an implicit meta policy whose sub_policy and key is policyname with rule ANY.
func ImplicitMetaAnyPolicy(policyName string) *StandardConfigPolicy {
	return &StandardConfigPolicy{
		key:   policyName,
		value: makeImplicitMetaPolicy(policyName, cb.ImplicitMetaPolicy_ANY),
	}
}

// ImplicitMetaMajorityPolicy defines an implicit meta policy whose sub_policy and key is policyname with rule MAJORITY.
func ImplicitMetaMajorityPolicy(policyName string) *StandardConfigPolicy {
	return &StandardConfigPolicy{
		key:   policyName,
		value: makeImplicitMetaPolicy(policyName, cb.ImplicitMetaPolicy_MAJORITY),
	}
}

// SignaturePolicy defines a policy with key policyName and the given signature policy.
func SignaturePolicy(policyName string, sigPolicy *cb.SignaturePolicyEnvelope) *StandardConfigPolicy {
	return &StandardConfigPolicy{
		key: policyName,
		value: &cb.Policy{
			Type:  int32(cb.Policy_SIGNATURE),
			Value: protoutil.MarshalOrPanic(sigPolicy),
		},
	}
}

func ImplicitMetaFromString(input string) (*cb.ImplicitMetaPolicy, error) {
	args := strings.Split(input, " ")
	if len(args) != 2 {
		return nil, fmt.Errorf("expected two space separated tokens, but got %d", len(args))
	}

	res := &cb.ImplicitMetaPolicy{
		SubPolicy: args[1],
	}

	switch args[0] {
	case cb.ImplicitMetaPolicy_ANY.String():
		res.Rule = cb.ImplicitMetaPolicy_ANY
	case cb.ImplicitMetaPolicy_ALL.String():
		res.Rule = cb.ImplicitMetaPolicy_ALL
	case cb.ImplicitMetaPolicy_MAJORITY.String():
		res.Rule = cb.ImplicitMetaPolicy_MAJORITY
	default:
		return nil, fmt.Errorf("unknown rule type '%s', expected ALL, ANY, or MAJORITY", args[0])
	}

	return res, nil
}
