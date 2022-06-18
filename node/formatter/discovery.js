import {DiscoveryResultType} from './constants.js';
import fabprotos from 'fabric-protos';
import assert from 'assert';
export const ParsePeerResult = ({identity, membership_info, state_info}) => {
	const peer = {};
	// IDENTITY
	{
		const {mspid, id_bytes} = fabprotos.msp.SerializedIdentity.decode(identity);
		peer.identity = {
			mspid,
			id_bytes: id_bytes.toString()
		};
	}

	// MEMBERSHIP - Peer.membership_info
	// gossip.Envelope.payload
	{
		const {payload, signature, secret_envelope} = membership_info;
		assert.strictEqual(secret_envelope, null);
		const {tag, alive_msg} = fabprotos.gossip.GossipMessage.decode(payload);
		assert.strictEqual(tag, 1);
		const {membership: {endpoint, pki_id}, timestamp: {inc_num, seq_num}} = alive_msg;
		// TODO WIP what is the content of pki_id and readable format
		// TODO What is the gossip inc_num, seq_num
		// TODO What is this inc_num: Long { low: -664492743, high: 377579470, unsigned: true },
		peer.membership_info = {endpoint};
	}

	// STATE
	if (state_info) {
		const {payload, signature, secret_envelope} = state_info;
		assert.strictEqual(secret_envelope, null);
		const {tag, state_info: {timestamp, pki_id, channel_MAC, properties}} = fabprotos.gossip.GossipMessage.decode(payload);
		assert.strictEqual(tag, 5);
		const {chaincodes, ledger_height} = properties;
		peer.ledger_height = ledger_height.toInt();
		peer.chaincodes = chaincodes.map(({name, version}) => ({name, version}));
	}
	return peer;
};
export const ParseResult = ({results}) => {
	const returned = {};

	for (const {result, error, config_result, cc_query_res, members} of results) {
		switch (result) {
			case DiscoveryResultType.error:
				returned.error = error;
				break;
			case DiscoveryResultType.cc_query_res:
				returned.cc_query_res = cc_query_res.content;
				break;
			case DiscoveryResultType.config_result: {

				const {msps, orderers} = config_result;

				for (const [mspid, q_msp] of Object.entries(msps)) {
					msps[mspid] = {
						organizational_unit_identifiers: q_msp.organizational_unit_identifiers,
						root_certs: q_msp.root_certs.toString().trim(),
						intermediate_certs: q_msp.intermediate_certs.toString().trim(),
						admins: q_msp.admins.toString().trim(),
						tls_root_certs: q_msp.tls_root_certs.toString().trim(),
						tls_intermediate_certs: q_msp.tls_intermediate_certs.toString().trim()
					};
				}

				for (const [mspid, {endpoint}] of Object.entries(orderers)) {
					orderers[mspid] = endpoint;
				}
				returned.config_result = config_result;
			}
				break;
			case DiscoveryResultType.members: {
				const {peers_by_org} = members;


				for (const [mspid, {peers}] of Object.entries(peers_by_org)) {

					peers_by_org[mspid] = peers;

				}
				returned.members = peers_by_org;
			}
				break;
		}
	}
	return returned;
};
