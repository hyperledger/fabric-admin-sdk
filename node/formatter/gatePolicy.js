import fabricProtos from '@hyperledger/fabric-protos';

const {common: commonProtos, msp:mspProtos} = fabricProtos;
const {SignaturePolicy, SignaturePolicyEnvelope} = commonProtos
const {MSPPrincipal, MSPRole} = mspProtos

export const GateClausePattern = /^(AND|OR)\(([\w,.\s()']+)\)$/;
export const RoleClausePattern = /^'([0-9A-Za-z.-]+)(\.)(admin|member|client|peer|orderer)'$/;

/**
 *  Reference: `common/policydsl/policyparser.go`
 *      `func FromString(policy string) (*cb.SignaturePolicyEnvelope, error)`
 */
export default class GatePolicy {

	/**
	 *
	 * @param {MSPRoleTypeMap[keyof MSPRoleTypeMap]} MSPRoleType
	 * @param {string} mspid
	 * @return {MSPPrincipal}
	 */
	static buildMSPPrincipal(MSPRoleType, mspid) {
		const newPrincipal = new MSPPrincipal();
		newPrincipal.setPrincipalClassification(MSPPrincipal.Classification.ROLE);
		const newRole = new MSPRole()
		newRole.setRole(MSPRoleType)
		newRole.setMspIdentifier(mspid)
		newPrincipal.setPrincipal(newRole.serializeBinary());
		return newPrincipal;
	}

	/**
	 *
	 * @param {number} n
	 * @param {Array<SignaturePolicy>} rules
	 * @return {SignaturePolicy.NOutOf}
	 */
	static buildNOutOf({n, rules}) {
		const nOutOf= new SignaturePolicy.NOutOf()
		nOutOf.setN(n)
		nOutOf.setRulesList(rules)
		return nOutOf
	}


	/**
	 *
	 * @param {SignaturePolicy.NOutOf} [n_out_of] exclusive with signed_by
	 * @param {number} [signed_by] exclusive with n_out_of
	 * @return {SignaturePolicy}
	 */
	static buildSignaturePolicy({n_out_of, signed_by}) {
		const signaturePolicy = new SignaturePolicy();
		if (n_out_of) {
			signaturePolicy.setNOutOf(n_out_of)
		} else if (signed_by || signed_by === 0) {
			signaturePolicy.setSignedBy(signed_by)
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
				identities[index] = GatePolicy.buildMSPPrincipal(MSPRole.MSPRoleType[role.toUpperCase()], mspid);

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

		const signaturePolicyEnvelope = new SignaturePolicyEnvelope();
		signaturePolicyEnvelope.setRule(rule)
		signaturePolicyEnvelope.setIdentitiesList(identities)

		return signaturePolicyEnvelope;
	}
}


