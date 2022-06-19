import fabricProtos from '@hyperledger/fabric-protos';
import {BufferFrom} from './protobuf.js';

const {common: commonProto} = fabricProtos;

/**
 * @typedef {Object} RoleIdentity
 * @property {{type: integer, mspid: MspId}} [role]
 */

/**
 * @typedef {Object} Policy
 * @property {RoleIdentity[]} identities
 * @property {Object} policy Recursive policy object
 * @example
 * {
	    identities: [
	      { role: { type: 0, mspid: "org1msp" }},
	      { role: { type: 0, mspid: "org2msp" }}
	    ],
	    policy: {
	      "1-of": [{ signedBy: 0 }, { signedBy: 1 }]
	    }
	  }
 */
/**
 * @typedef {Object} PolicyElement
 * @property {PolicyElement[]} [n-of]
 * @property {integer|MSPRoleType} [signedBy]
 */

export default class Policy {

	/**
	 *
	 * @param {Policy} policy
	 */
	static buildSignaturePolicyEnvelope(policy) {
		const envelope = new commonProto.SignaturePolicyEnvelope();
		const principals = policy.identities.map((identity) => this.buildPrincipal(identity));
		const thePolicy = Policy.parsePolicy(policy.policy);

		envelope.version = 0;
		envelope.rule = thePolicy;
		envelope.identities = principals;

		return envelope;
	}

	static buildPrincipal(identity) {
		const {type, mspid} = identity.role;

		const newPrincipal = new commonProto.MSPPrincipal();
		newPrincipal.principal_classification = commonProto.MSPPrincipal.Classification.ROLE;
		newPrincipal.principal = BufferFrom({role: type, msp_identifier: mspid}, commonProto.MSPRole);

		return newPrincipal;
	}

	static parsePolicy(spec) {

		const signedByConfig = spec.signedBy;
		const signaturePolicy = new commonProto.SignaturePolicy();
		if (signedByConfig || signedByConfig === 0) {
			signaturePolicy.Type = 'signed_by';
			signaturePolicy.signed_by = signedByConfig;
		} else {
			const key = Object.keys(spec)[0];
			const array = spec[key];
			const n = key.match(/^(\d+)-of$/)[1];

			const nOutOf = new commonProto.SignaturePolicy.NOutOf();

			nOutOf.n = parseInt(n);
			nOutOf.rules = array.map((sub) => this.parsePolicy(sub));

			signaturePolicy.n_out_of = nOutOf;
			signaturePolicy.Type = 'n_out_of';
		}
		return signaturePolicy;
	}
}
