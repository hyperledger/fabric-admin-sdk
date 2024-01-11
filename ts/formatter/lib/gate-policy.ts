import {common, msp} from '@hyperledger/fabric-protos';
import * as SignaturePolicyNS from './proto/signature-policy.js'
import * as MSPPrincipalNS from './msp-principal.js'
import SignaturePolicy = common.SignaturePolicy
import NOutOf = SignaturePolicy.NOutOf;

const {SignaturePolicyEnvelope} = common
const {MSPRole} = msp

export const GateClausePattern = /^(AND|OR)\(([\w,.\s()']+)\)$/;
export const RoleClausePattern = /^'([0-9A-Za-z.-]+)(\.)(admin|member|client|peer|orderer)'$/;


/**
 *  Reference: `common/policydsl/policyparser.go`
 *      `func FromString(policy string) (*cb.SignaturePolicyEnvelope, error)`
 */
export function FromString(policyString: string) {
    const identitiesIndexMap = {};
    const identities = [];
    const parseRoleClause = (mspid, role) => {
        const key = `${mspid}.${role}`;
        if (!identitiesIndexMap[key] && identitiesIndexMap[key] !== 0) {

            const index = identities.length;

            identitiesIndexMap[key] = index;
            identities[index] = MSPPrincipalNS.build(MSPRole.MSPRoleType[role.toUpperCase()], mspid);

        }
        return SignaturePolicyNS.build({signed_by: identitiesIndexMap[key]});
    };

    const parseGateClause = (clause?, gate?, subClause?) => {
        if (clause) {
            const result = clause.match(GateClausePattern);
            gate = result[1];
            subClause = result[2];
        }
        const subClauseItems = subClause.split(',');
        let n_out_of: NOutOf;
        const rules = [];
        for (const subClauseItem of subClauseItems) {
            const trimmed = subClauseItem.trim();
            if (!trimmed) {
                continue;
            }
            let subResult = subClauseItem.match(GateClausePattern);
            if (subResult) {
                const subRule = parseGateClause(undefined, subResult[1], subResult[2]);
                rules.push(subRule);
            }
            subResult = subClauseItem.match(RoleClausePattern);
            if (subResult) {
                const surRule = parseRoleClause(subResult[1], subResult[3]);
                rules.push(surRule);
            }

        }
        if (gate === 'OR') {
            n_out_of = SignaturePolicyNS.buildNOutOf({n: 1, rules});
        } else if (gate === 'AND') {
            n_out_of = SignaturePolicyNS.buildNOutOf({n: rules.length, rules});
        }

        return SignaturePolicyNS.build({n_out_of});
    };

    const rule = parseGateClause(policyString);

    const signaturePolicyEnvelope = new SignaturePolicyEnvelope();
    signaturePolicyEnvelope.setRule(rule)
    signaturePolicyEnvelope.setIdentitiesList(identities)

    return signaturePolicyEnvelope;
}


