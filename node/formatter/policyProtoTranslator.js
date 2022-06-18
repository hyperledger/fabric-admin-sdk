import fabprotos from 'fabric-protos';
import {MSPRoleType} from './constants.js';
const {ROLE, ORGANIZATION_UNIT, IDENTITY} = fabprotos.common.MSPPrincipal.Classification;
export const decodeIdentity = (id_bytes) => {
	const identity = {};

	const proto_identity = fabprotos.msp.SerializedIdentity.decode(id_bytes);
	identity.Mspid = proto_identity.mspid;
	identity.IdBytes = proto_identity.id_bytes.toBuffer().toString();

	return identity;
};

export const decodeMSPPrincipal = (proto_msp_principal) => {
	let msp_principal = {};
	msp_principal.principal_classification = proto_msp_principal.principal_classification;
	let proto_principal = null;
	switch (msp_principal.principal_classification) {
		case ROLE:
			proto_principal = fabprotos.common.MSPRole.decode(proto_msp_principal.principal);
			msp_principal.msp_identifier = proto_principal.msp_identifier;
			msp_principal.Role = MSPRoleType[proto_principal.role];
			break;
		case ORGANIZATION_UNIT:
			proto_principal = fabprotos.common.OrganizationUnit.decode(proto_msp_principal.principal);
			msp_principal.msp_identifier = proto_principal.msp_identifier; // string
			msp_principal.organizational_unit_identifier = proto_principal.organizational_unit_identifier; // string
			msp_principal.certifiers_identifier = proto_principal.certifiers_identifier.toBuffer(); // bytes
			break;
		case IDENTITY:
			msp_principal = decodeIdentity(proto_msp_principal.principal);
			break;
	}

	return msp_principal;
};
export const decodeSignaturePolicy = (proto_signature_policy) => {
	const signature_policy = {};
	signature_policy.Type = proto_signature_policy.Type;
	switch (signature_policy.Type) {
		case 'n_out_of':
			signature_policy.n_out_of = {};
			signature_policy.n_out_of.N = proto_signature_policy.n_out_of.n;
			signature_policy.n_out_of.rules = [];
			for (const proto_policy of proto_signature_policy.n_out_of.rules) {
				const policy = decodeSignaturePolicy(proto_policy);
				signature_policy.n_out_of.rules.push(policy);
			}
			break;
		case 'signed_by':
			signature_policy.signed_by = proto_signature_policy.signed_by;
			break;
	}
	return signature_policy;
};
export const decodeSignaturePolicyEnvelope = (signature_policy_envelope_bytes) => {
	const signature_policy_envelope = {};
	const proto_signature_policy_envelope = fabprotos.common.SignaturePolicyEnvelope.decode(signature_policy_envelope_bytes);
	signature_policy_envelope.version = proto_signature_policy_envelope.version;
	signature_policy_envelope.rule = decodeSignaturePolicy(proto_signature_policy_envelope.rule);
	const proto_identities = proto_signature_policy_envelope.identities;
	if (proto_identities) {
		signature_policy_envelope.identities = proto_identities.map(proto_identity => {
			return decodeMSPPrincipal(proto_identity);
		});
	}


	return signature_policy_envelope;
};
