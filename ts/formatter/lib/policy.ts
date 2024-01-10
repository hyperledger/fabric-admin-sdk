import {msp} from '@hyperledger/fabric-protos'
import * as MSPPrincipalNS from './msp-principal'
import {buildEnvelope, build, buildNOutOf} from './proto/signature-policy'
import {IndexDigit, MspId} from "./types";
import assert from "assert";
import MSPRoleTypeMap = msp.MSPRole.MSPRoleTypeMap;

/**
 * @example
 *  {
 *      role: { type: 0, mspid: "org1msp" }
 *  }
 */
type MSPRoleAlias = {
    role: {
        type: MSPRoleTypeMap
        mspid: MspId
    }
}

class N_OF extends String {
    constructor(str) {
        super(str);
        assert.ok(this.match(/^\d+-of$/), "invalid format: n-of should match ^\\d+-of$ ")
    }

    get n() {
        return parseInt(this.match(/^(\d+)-of$/)[1]);
    }
}

type SignedByPolicy = {
    signedBy?: IndexDigit,
}
/**
 * @example
 * {
 *      "1-of": [{ signedBy: 0 }, { signedBy: 1 }]
 * }
 */
type N_OfPolicy = {
    [n in number as `${n}-of`]: Policy[]
}
type Policy = SignedByPolicy | N_OfPolicy

export function isSignedByPolicy(policy): policy is SignedByPolicy {
    const keys = Object.keys(policy)
    assert.equal(keys.length, 1, `ambiguous policy: multiple key founds: ${keys}`)

    return keys[0] === 'signedBy';
}

export function buildRule(policy: Policy) {
    const keys = Object.keys(policy)

    if (isSignedByPolicy(policy)) {
        return build({signed_by: policy.signedBy})
    } else {
        const key = keys[0];
        const array = policy[key];
        const n_of = new N_OF(key)
        const n = n_of.n;

        const rules = array.map((sub) => buildRule(sub))
        const n_out_of = buildNOutOf({n, rules})
        return build({n_out_of})
    }
}

export function buildSignaturePolicyEnvelope({identities, policy}: { identities: MSPRoleAlias[], policy: Policy }) {
    const principals = identities.map((identity) => buildMSPPrincipal(identity));
    const rule = buildRule(policy);
    return buildEnvelope({identities: principals, rule})
}

export function buildMSPPrincipal({role}) {
    const {type, mspid} = role;
    return MSPPrincipalNS.build(type, mspid);
}