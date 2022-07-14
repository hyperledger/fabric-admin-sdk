import {common} from '@hyperledger/fabric-protos';
import SignaturePolicy = common.SignaturePolicy
import NOutOf = SignaturePolicy.NOutOf;
import {IndexDigit, isIndexDigit} from "./d";
import assert from 'assert'

const {SignaturePolicyEnvelope} = common

// SignaturePolicyType is the 'Type' string for signature policies
export const type = "Signature"
export const TypeCase = ['TYPE_NOT_SET', 'SIGNED_BY', 'N_OUT_OF']

/**
 *
 * @param {NOutOf} [n_out_of] exclusive with signed_by
 * @param {IndexDigit} [signed_by] exclusive with n_out_of
 * @return {SignaturePolicy}
 */
export function build({n_out_of, signed_by}: { n_out_of?: NOutOf; signed_by?: IndexDigit }) {
    const signaturePolicy = new SignaturePolicy();
    if (n_out_of) {
        signaturePolicy.setNOutOf(n_out_of)
    } else if (isIndexDigit(signed_by)) {
        signaturePolicy.setSignedBy(signed_by)
    }
    return signaturePolicy;
}

/**
 *
 * @param {IndexDigit} n
 * @param {SignaturePolicy[]} rules
 * @return {NOutOf}
 */
export function buildNOutOf({n, rules}: { n: IndexDigit, rules: SignaturePolicy[] }) {
    const nOutOf = new SignaturePolicy.NOutOf()
    assert.ok(isIndexDigit(n))
    nOutOf.setN(n)
    nOutOf.setRulesList(rules)
    return nOutOf
}

export function buildEnvelope({identities, rule}) {
    const envelope = new SignaturePolicyEnvelope();
    envelope.setVersion(0);
    envelope.setRule(rule);
    envelope.setIdentitiesList(identities);

    return envelope;
}
