package policydsl

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	mb "github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"google.golang.org/protobuf/proto"
)

// AcceptAllPolicy always evaluates to true
var AcceptAllPolicy *cb.SignaturePolicyEnvelope

// Gate values
const (
	GateAnd   = "And"
	GateOr    = "Or"
	GateOutOf = "OutOf"

	RoleAdmin   = "admin"
	RoleMember  = "member"
	RoleClient  = "client"
	RolePeer    = "peer"
	RoleOrderer = "orderer"
)

var regex = regexp.MustCompile(
	fmt.Sprintf("^([[:alnum:].-]+)([.])(%s|%s|%s|%s|%s)$",
		RoleAdmin, RoleMember, RoleClient, RolePeer, RoleOrderer),
)

var regexErr = regexp.MustCompile("^No parameter '([^']+)' found[.]$")

// a stub function - it returns the same string as it's passed.
// This will be evaluated by second/third passes to convert to a proto policy
func outof(args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("expected at least two arguments to NOutOf. Given %d", len(args))
	}

	var result strings.Builder

	result.WriteString("outof(")

	arg0 := args[0]
	switch n := arg0.(type) {
	case int:
		result.WriteString(strconv.Itoa(n))
	case string:
		result.WriteString(n)
	default:
		return nil, fmt.Errorf("unexpected type %T", arg0)
	}

	for _, arg := range args[1:] {
		result.WriteString(", ")

		switch t := arg.(type) {
		case string:
			if regex.MatchString(t) {
				result.WriteString("'" + t + "'")
			} else {
				result.WriteString(t)
			}
		default:
			return nil, fmt.Errorf("unexpected type %T", arg)
		}
	}

	result.WriteString(")")

	return result.String(), nil
}

func and(args ...any) (any, error) {
	args = append([]any{len(args)}, args...)
	return outof(args...)
}

func or(args ...any) (any, error) {
	args = append([]any{1}, args...)
	return outof(args...)
}

//nolint:cyclop,gocognit
func FromString(policy string) (*cb.SignaturePolicyEnvelope, error) {
	// first we translate the and/or business into outof gates
	env := map[string]any{
		GateAnd:                    and,
		strings.ToLower(GateAnd):   and,
		strings.ToUpper(GateAnd):   and,
		GateOr:                     or,
		strings.ToLower(GateOr):    or,
		strings.ToUpper(GateOr):    or,
		GateOutOf:                  outof,
		strings.ToLower(GateOutOf): outof,
		strings.ToUpper(GateOutOf): outof,
	}
	intermediate, err := expr.Compile(policy, expr.Env(env))
	if err != nil {
		return nil, err
	}

	intermediateRes, err := expr.Run(intermediate, env)
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
	// required because the parser has no means of giving context
	// to user-implemented functions other than via arguments.
	// We need this argument because we need a global place where
	// we put the identities that the policy requires
	env = map[string]any{
		"outof": firstPass,
	}
	exp, err := expr.Compile(resStr, expr.Env(env))
	if err != nil {
		return nil, err
	}

	res, err := expr.Run(exp, env)
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
	env = map[string]any{
		"outof": secondPass,
		"ID":    ctx,
	}
	exp, err = expr.Compile(resStr, expr.Env(env))
	if err != nil {
		return nil, err
	}

	res, err = expr.Run(exp, env)
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

func firstPass(args ...any) (any, error) {
	var result strings.Builder

	result.WriteString("outof(ID")

	for _, arg := range args {
		result.WriteString(", ")

		switch t := arg.(type) {
		case string:
			if regex.MatchString(t) {
				result.WriteString("'" + t + "'")
			} else {
				result.WriteString(t)
			}
		case int:
			result.WriteString(strconv.Itoa(t))
		default:
			return nil, fmt.Errorf("unexpected type %T", arg)
		}
	}

	result.WriteString(")")

	return result.String(), nil
}

type context struct {
	IDNum      int32
	principals []*mb.MSPPrincipal
}

func newContext() *context {
	return &context{IDNum: 0, principals: make([]*mb.MSPPrincipal, 0)}
}

//nolint:cyclop,gocognit
func secondPass(args ...any) (any, error) {
	/* general sanity check, we expect at least 3 args */
	if len(args) < 3 {
		return nil, fmt.Errorf("at least 3 arguments expected, got %d", len(args))
	}

	/* get the first argument, we expect it to be the context */
	var ctx *context
	switch v := args[0].(type) {
	case *context:
		ctx = v
	default:
		return nil, fmt.Errorf("unrecognized type, expected the context, got %T", args[0])
	}

	/* get the second argument, we expect an integer telling us
	   how many of the remaining we expect to have*/
	var t int
	switch arg := args[1].(type) {
	case int:
		t = arg
	default:
		return nil, fmt.Errorf("unrecognized type, expected a number, got %T", args[1])
	}

	/* get the n in the t out of n */
	n := len(args) - 2

	/* sanity check - t should be positive, permit equal to n+1, but disallow over n+1 */
	if t < 0 || t > n+1 {
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
			case RoleMember:
				r = mb.MSPRole_MEMBER
			case RoleAdmin:
				r = mb.MSPRole_ADMIN
			case RoleClient:
				r = mb.MSPRole_CLIENT
			case RolePeer:
				r = mb.MSPRole_PEER
			case RoleOrderer:
				r = mb.MSPRole_ORDERER
			default:
				return nil, fmt.Errorf("error parsing role %s", t)
			}

			/* build the principal we've been told */
			mspRole, err := proto.Marshal(&mb.MSPRole{MspIdentifier: subm[0][1], Role: r})
			if err != nil {
				return nil, fmt.Errorf("error marshalling msp role: %w", err)
			}

			p := &mb.MSPPrincipal{
				PrincipalClassification: mb.MSPPrincipal_ROLE,
				Principal:               mspRole,
			}
			ctx.principals = append(ctx.principals, p)

			/* create a SignaturePolicy that requires a signature from
			   the principal we've just built*/
			dapolicy := SignedBy(ctx.IDNum)
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
			return nil, fmt.Errorf("unrecognized type, expected a principal or a policy, got %T", principal)
		}
	}

	if t > math.MaxInt32 || t < math.MinInt32 {
		return nil, fmt.Errorf("t value of %d for t-out-of-n predicate will overflow when converted to an int32", t)
	}

	return NOutOf(int32(t), policies), nil
}

// SignedBy creates a SignaturePolicy requiring a given signer's signature
func SignedBy(index int32) *cb.SignaturePolicy {
	return &cb.SignaturePolicy{
		Type: &cb.SignaturePolicy_SignedBy{
			SignedBy: index,
		},
	}
}

// NOutOf creates a policy which requires N out of the slice of policies to evaluate to true
func NOutOf(n int32, policies []*cb.SignaturePolicy) *cb.SignaturePolicy {
	return &cb.SignaturePolicy{
		Type: &cb.SignaturePolicy_NOutOf_{
			NOutOf: &cb.SignaturePolicy_NOutOf{
				N:     n,
				Rules: policies,
			},
		},
	}
}
