import fabricProtos from 'fabric-protos';
import {BufferFrom, ProtoFrom} from './protobuf.js';

const {common: commonProtos} = fabricProtos;
const MSPRoleTypeInverse = {
	'MEMBER': 0,
	'ADMIN': 1,
	'CLIENT': 2,
	'PEER': 3,
	'ORDERER': 4
};
const GateClausePattern = /^(AND|OR)\(([\w,.\s()']+)\)$/;
const RoleClausePattern = /^'([0-9A-Za-z.-]+)(\.)(admin|member|client|peer|orderer)'$/;

/**
 *  Reference: `common/policydsl/policyparser.go`
 *      `func FromString(policy string) (*cb.SignaturePolicyEnvelope, error)`
 */
export default class GatePolicy {

	static buildMSPPrincipal(MSPRoleType, mspid) {
		const newPrincipal = new commonProtos.MSPPrincipal();
		newPrincipal.principal_classification = commonProtos.MSPPrincipal.Classification.ROLE;
		const newRole = {role: MSPRoleType, msp_identifier: mspid};
		newPrincipal.principal = BufferFrom(newRole, commonProtos.MSPRole);
		return newPrincipal;
	}

	static buildNOutOf({n, rules: SignaturePolicyArray}) {
		return ProtoFrom({n, rules: SignaturePolicyArray}, commonProtos.SignaturePolicy.NOutOf);
	}


	static buildSignaturePolicy({n_out_of, signed_by}) {
		const signaturePolicy = new commonProtos.SignaturePolicy();
		if (n_out_of) {
			signaturePolicy.Type = 'n_out_of';
			signaturePolicy.n_out_of = n_out_of;
		} else if (signed_by || signed_by === 0) {
			signaturePolicy.Type = 'signed_by';
			signaturePolicy.signed_by = signed_by;
		}
		return signaturePolicy;
	}


	static FromString(policyString) {
		const identitiesIndexMap = {};
		const identities = [];
		const parseRoleClause = (mspid, role) => {
			const key = `${mspid}.${role}`;
			if (!identitiesIndexMap[key] && identitiesIndexMap[key] !== 0) {

				const index = identities.length;

				identitiesIndexMap[key] = index;
				identities[index] = GatePolicy.buildMSPPrincipal(MSPRoleTypeInverse[role.toUpperCase()], mspid);

			}
			return GatePolicy.buildSignaturePolicy({signed_by: identitiesIndexMap[key]});
		};

		const parseGateClause = (clause, gate, subClause) => {
			if (clause) {
				const result = clause.match(GateClausePattern);
				gate = result[1];
				subClause = result[2];
			}
			const subClauseItems = subClause.split(',');
			let n_out_of;
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
				n_out_of = GatePolicy.buildNOutOf({n: 1, rules});
			} else if (gate === 'AND') {
				n_out_of = GatePolicy.buildNOutOf({n: rules.length, rules});
			}

			return GatePolicy.buildSignaturePolicy({n_out_of});
		};

		const rule = parseGateClause(policyString);

		const signaturePolicyEnvelope = new commonProtos.SignaturePolicyEnvelope();
		signaturePolicyEnvelope.rule = rule;
		signaturePolicyEnvelope.identities = identities;

		return signaturePolicyEnvelope;
	}
}

GatePolicy.MSPRoleTypeInverse = MSPRoleTypeInverse;
GatePolicy.GateClausePattern = GateClausePattern;
GatePolicy.RoleClausePattern = RoleClausePattern;

