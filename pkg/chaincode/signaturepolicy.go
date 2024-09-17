/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	mb "github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
)

// Gate values
const (
	gateand   = "And"
	gateor    = "Or"
	gateoutof = "OutOf"
)

// Role values for principals
const (
	roleadmin   = "admin"
	rolemember  = "member"
	roleclient  = "client"
	rolepeer    = "peer"
	roleorderer = "orderer"
)

var regex = regexp.MustCompile(
	fmt.Sprintf("^([[:alnum:].-]+)([.])(%s|%s|%s|%s|%s)$",
		roleadmin, rolemember, roleclient, rolepeer, roleorderer),
)

var regexErr = regexp.MustCompile("^No parameter '([^']+)' found[.]$")

// a stub function - it returns the same string as it's passed.
// This will be evaluated by second/third passes to convert to a proto policy
func outof(args ...interface{}) (interface{}, error) {
	toret := "outof("

	if len(args) < 2 {
		return nil, fmt.Errorf("expected at least two arguments to NOutOf. Given %d", len(args))
	}

	arg0 := args[0]
	// govaluate treats all numbers as float64 only. But and/or may pass int/string. Allowing int/string for flexibility of caller
	if n, ok := arg0.(float64); ok {
		toret += strconv.Itoa(int(n))
	} else if n, ok := arg0.(int); ok {
		toret += strconv.Itoa(n)
	} else if n, ok := arg0.(string); ok {
		toret += n
	} else {
		return nil, fmt.Errorf("unexpected type %s", reflect.TypeOf(arg0))
	}

	for _, arg := range args[1:] {
		toret += ", "

		switch t := arg.(type) {
		case string:
			if regex.MatchString(t) {
				toret += "'" + t + "'"
			} else {
				toret += t
			}
		default:
			return nil, fmt.Errorf("unexpected type %s", reflect.TypeOf(arg))
		}
	}

	return toret + ")", nil
}

func and(args ...interface{}) (interface{}, error) {
	args = append([]interface{}{len(args)}, args...)
	return outof(args...)
}

func or(args ...interface{}) (interface{}, error) {
	args = append([]interface{}{1}, args...)
	return outof(args...)
}

func firstPass(args ...interface{}) (interface{}, error) {
	toret := "outof(ID"
	for _, arg := range args {
		toret += ", "

		switch t := arg.(type) {
		case string:
			if regex.MatchString(t) {
				toret += "'" + t + "'"
			} else {
				toret += t
			}
		case float32:
		case float64:
			toret += strconv.Itoa(int(t))
		default:
			return nil, fmt.Errorf("unexpected type %s", reflect.TypeOf(arg))
		}
	}

	return toret + ")", nil
}

// signedBy creates a SignaturePolicy requiring a given signer's signature
func signedBy(index int32) *cb.SignaturePolicy {
	return &cb.SignaturePolicy{
		Type: &cb.SignaturePolicy_SignedBy{
			SignedBy: index,
		},
	}
}

// nOutOf creates a policy which requires N out of the slice of policies to evaluate to true
func nOutOf(n int32, policies []*cb.SignaturePolicy) *cb.SignaturePolicy {
	return &cb.SignaturePolicy{
		Type: &cb.SignaturePolicy_NOutOf_{
			NOutOf: &cb.SignaturePolicy_NOutOf{
				N:     n,
				Rules: policies,
			},
		},
	}
}

//nolint:gocognit,gocyclo
func secondPass(args ...interface{}) (interface{}, error) {
	/* general sanity check, we expect at least 3 args */
	if len(args) < 3 {
		return nil, fmt.Errorf("at least 3 arguments expected, got %d", len(args))
	}

	/* get the first argument, we expect it to be the context */
	var ctx *policyContext
	switch v := args[0].(type) {
	case *policyContext:
		ctx = v
	default:
		return nil, fmt.Errorf("unrecognized type, expected the context, got %s", reflect.TypeOf(args[0]))
	}

	/* get the second argument, we expect an integer telling us
	   how many of the remaining we expect to have*/
	var t int32
	switch arg := args[1].(type) {
	case float64:
		t = int32(arg)
	default:
		return nil, fmt.Errorf("unrecognized type, expected a number, got %s", reflect.TypeOf(args[1]))
	}

	/* get the n in the t out of n */
	n := len(args) - 2

	/* sanity check - t should be positive, permit equal to n+1, but disallow over n+1 */
	if t < 0 || int(t) > n+1 {
		return nil, fmt.Errorf("invalid t-out-of-n predicate, t %d, n %d", t, n)
	}

	policies := make([]*cb.SignaturePolicy, 0)

	/* handle the rest of the arguments */
	for _, principal := range args[2:] {
		switch t := principal.(type) {
		/* if it's a string, we expect it to be formed as
		   <MSP_ID> . <ROLE>, where MSP_ID is the MSP identifier
		   and ROLE is either a member, an admin, a client, a peer or an orderer*/
		case string:
			/* split the string */
			subm := regex.FindAllStringSubmatch(t, -1)
			if subm == nil || len(subm) != 1 || len(subm[0]) != 4 {
				return nil, fmt.Errorf("error parsing principal %s", t)
			}

			/* get the right role */
			var r mb.MSPRole_MSPRoleType

			switch subm[0][3] {
			case rolemember:
				r = mb.MSPRole_MEMBER
			case roleadmin:
				r = mb.MSPRole_ADMIN
			case roleclient:
				r = mb.MSPRole_CLIENT
			case rolepeer:
				r = mb.MSPRole_PEER
			case roleorderer:
				r = mb.MSPRole_ORDERER
			default:
				return nil, fmt.Errorf("error parsing role %s", t)
			}

			/* build the principal we've been told */
			mspRole, err := proto.Marshal(&mb.MSPRole{MspIdentifier: subm[0][1], Role: r})
			if err != nil {
				return nil, fmt.Errorf("error marshalling msp role: %s", err)
			}

			p := &mb.MSPPrincipal{
				PrincipalClassification: mb.MSPPrincipal_ROLE,
				Principal:               mspRole,
			}
			ctx.principals = append(ctx.principals, p)

			/* create a SignaturePolicy that requires a signature from
			   the principal we've just built*/
			dapolicy := signedBy(int32(ctx.IDNum))
			policies = append(policies, dapolicy)

			/* increment the identity counter. Note that this is
			   suboptimal as we are not reusing identities. We
			   can deduplicate them easily and make this puppy
			   smaller. For now it's fine though */
			// TODO: deduplicate principals
			ctx.IDNum++

		/* if we've already got a policy we're good, just append it */
		case *cb.SignaturePolicy:
			policies = append(policies, t)

		default:
			return nil, fmt.Errorf("unrecognized type, expected a principal or a policy, got %s", reflect.TypeOf(principal))
		}
	}

	return nOutOf(int32(t), policies), nil
}

type policyContext struct {
	IDNum      int32
	principals []*mb.MSPPrincipal
}

func newContext() *policyContext {
	return &policyContext{IDNum: 0, principals: make([]*mb.MSPPrincipal, 0)}
}

func NewApplicationPolicy(signaturePolicy, channelConfigPolicy string) (*peer.ApplicationPolicy, error) {
	signaturePolicyEnvelope, err := signaturePolicyEnvelopeFromString(signaturePolicy)
	if err != nil {
		return nil, err
	}
	applicationPolicy := &peer.ApplicationPolicy{
		Type: &peer.ApplicationPolicy_SignaturePolicy{
			SignaturePolicy: signaturePolicyEnvelope,
		},
	}
	if channelConfigPolicy != "" {
		applicationPolicy = &peer.ApplicationPolicy{
			Type: &peer.ApplicationPolicy_ChannelConfigPolicyReference{
				ChannelConfigPolicyReference: channelConfigPolicy,
			},
		}
	}
	return applicationPolicy, nil
}

// signaturePolicyEnvelopeFromString takes a string representation of the policy,
// parses it and returns a SignaturePolicyEnvelope that
// implements that policy. The supported language is as follows:
//
// GATE(P[, P])
//
// where:
//   - GATE is either "and" or "or"
//   - P is either a principal or another nested call to GATE
//
// A principal is defined as:
//
// # ORG.ROLE
//
// where:
//   - ORG is a string (representing the MSP identifier)
//   - ROLE takes the value of any of the RoleXXX constants representing
//     the required role
//
//nolint:gocognit,gocyclo
func signaturePolicyEnvelopeFromString(policy string) (*cb.SignaturePolicyEnvelope, error) {
	// first we translate the and/or business into outof gates
	intermediate, err := govaluate.NewEvaluableExpressionWithFunctions(
		policy, map[string]govaluate.ExpressionFunction{
			gateand:                    and,
			strings.ToLower(gateand):   and,
			strings.ToUpper(gateand):   and,
			gateor:                     or,
			strings.ToLower(gateor):    or,
			strings.ToUpper(gateor):    or,
			gateoutof:                  outof,
			strings.ToLower(gateoutof): outof,
			strings.ToUpper(gateoutof): outof,
		},
	)
	if err != nil {
		return nil, err
	}

	intermediateRes, err := intermediate.Evaluate(map[string]interface{}{})
	if err != nil {
		// attempt to produce a meaningful error
		if regexErr.MatchString(err.Error()) {
			sm := regexErr.FindStringSubmatch(err.Error())
			if len(sm) == 2 {
				return nil, fmt.Errorf("unrecognized token '%s' in policy string", sm[1])
			}
		}

		return nil, err
	}

	resStr, ok := intermediateRes.(string)
	if !ok {
		return nil, fmt.Errorf("invalid policy string '%s'", policy)
	}

	// we still need two passes. The first pass just adds an extra
	// argument ID to each of the outof calls. This is
	// required because govaluate has no means of giving context
	// to user-implemented functions other than via arguments.
	// We need this argument because we need a global place where
	// we put the identities that the policy requires
	exp, err := govaluate.NewEvaluableExpressionWithFunctions(
		resStr,
		map[string]govaluate.ExpressionFunction{"outof": firstPass},
	)
	if err != nil {
		return nil, err
	}

	res, err := exp.Evaluate(map[string]interface{}{})
	if err != nil {
		// attempt to produce a meaningful error
		if regexErr.MatchString(err.Error()) {
			sm := regexErr.FindStringSubmatch(err.Error())
			if len(sm) == 2 {
				return nil, fmt.Errorf("unrecognized token '%s' in policy string", sm[1])
			}
		}

		return nil, err
	}

	resStr, ok = res.(string)
	if !ok {
		return nil, fmt.Errorf("invalid policy string '%s'", policy)
	}

	ctx := newContext()
	parameters := make(map[string]interface{}, 1)
	parameters["ID"] = ctx

	exp, err = govaluate.NewEvaluableExpressionWithFunctions(
		resStr,
		map[string]govaluate.ExpressionFunction{"outof": secondPass},
	)
	if err != nil {
		return nil, err
	}

	res, err = exp.Evaluate(parameters)
	if err != nil {
		// attempt to produce a meaningful error
		if regexErr.MatchString(err.Error()) {
			sm := regexErr.FindStringSubmatch(err.Error())
			if len(sm) == 2 {
				return nil, fmt.Errorf("unrecognized token '%s' in policy string", sm[1])
			}
		}

		return nil, err
	}

	rule, ok := res.(*cb.SignaturePolicy)
	if !ok {
		return nil, fmt.Errorf("invalid policy string '%s'", policy)
	}

	p := &cb.SignaturePolicyEnvelope{
		Identities: ctx.principals,
		Version:    0,
		Rule:       rule,
	}

	return p, nil
}

// SignaturePolicyEnvelopeToString parse a SignaturePolicyEnvelope to human readable expression
// the returned expression is GATE(P[, P])
//
// where:
//   - GATE is either "and" or "or" or "outof"
//   - P is either a principal or another nested call to GATE
//
// A principal is defined as:
//
// # ORG.ROLE
//
// where:
//   - ORG is a string (representing the MSP identifier)
//   - ROLE takes the value of any of the RoleXXX constants representing
//     the required role
func SignaturePolicyEnvelopeToString(policy *cb.SignaturePolicyEnvelope) (string, error) {
	ids := []string{}
	for _, id := range policy.GetIdentities() {
		var mspRole mb.MSPRole
		if err := proto.Unmarshal(id.GetPrincipal(), &mspRole); err != nil {
			return "", err
		}

		mspID := mspRole.GetMspIdentifier() + "." + strings.ToLower(mb.MSPRole_MSPRoleType_name[int32(mspRole.GetRole())])
		ids = append(ids, mspID)
	}

	var buf bytes.Buffer
	policyParse(policy.GetRule(), ids, &buf)
	return buf.String(), nil
}

// recursive parse
func policyParse(rule *cb.SignaturePolicy, ids []string, buf *bytes.Buffer) {
	switch p := rule.GetType().(type) {
	case *cb.SignaturePolicy_SignedBy:
		buf.WriteString("'")
		buf.WriteString(ids[p.SignedBy])
		buf.WriteString("'")

	case *cb.SignaturePolicy_NOutOf_:
		n := p.NOutOf.GetN()
		rules := p.NOutOf.GetRules()

		switch n {
		case int32(len(rules)): //#nosec:G115
			buf.WriteString("AND(")
		case 1:
			buf.WriteString("OR(")
		default:
			buf.WriteString("OutOf(")
			buf.WriteString(strconv.Itoa(int(n)))
			buf.WriteString(",")
		}

		for i, r := range rules {
			if i > 0 {
				buf.WriteString(",")
			}
			policyParse(r, ids, buf)
		}
		buf.WriteString(")")
	}
}
