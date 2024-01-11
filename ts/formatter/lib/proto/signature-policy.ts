import {common, msp} from '@hyperledger/fabric-protos';
import SignaturePolicy = common.SignaturePolicy
import SignaturePolicyEnvelope = common.SignaturePolicyEnvelope
import NOutOf = SignaturePolicy.NOutOf;
import {IndexDigit, isIndex} from "../types.js";
import assert from 'assert'
import MSPPrincipal = msp.MSPPrincipal


/**
 * SignaturePolicyType is the 'Type' string for signature policies
 */
export const SignaturePolicyType = "Signature"
export const TypeCase = ['TYPE_NOT_SET', 'SIGNED_BY', 'N_OUT_OF']

/**
 * @param {NOutOf} [n_out_of] exclusive with signed_by
 * @param [signed_by] exclusive with n_out_of
 * @return {SignaturePolicy}
 */
export function build({n_out_of, signed_by}: { n_out_of?: NOutOf; signed_by?: IndexDigit }): SignaturePolicy {
    const signaturePolicy = new SignaturePolicy();
    if (n_out_of) {
        signaturePolicy.setNOutOf(n_out_of)
    } else if (isIndex(signed_by)) {
        signaturePolicy.setSignedBy(signed_by)
    }
    return signaturePolicy;
}

export function buildNOutOf({n, rules}: { n: IndexDigit, rules: SignaturePolicy[] }): NOutOf {
    const nOutOf = new SignaturePolicy.NOutOf()
    assert.ok(isIndex(n))
    nOutOf.setN(n)
    nOutOf.setRulesList(rules)
    return nOutOf
}

export function buildEnvelope({identities, rule}: {
    identities: MSPPrincipal[],
    rule: SignaturePolicy
}): SignaturePolicyEnvelope {
    const envelope = new SignaturePolicyEnvelope();
    envelope.setVersion(0);
    envelope.setRule(rule);
    envelope.setIdentitiesList(identities);

    return envelope;
}
