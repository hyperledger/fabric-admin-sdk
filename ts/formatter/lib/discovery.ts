import {msp, gossip} from '@hyperledger/fabric-protos'
import assert from 'assert';
import {calculatePKI_ID} from "./helper";

interface PeerResult {
    identity: msp.SerializedIdentity.AsObject,
    membership_info: {
        endpoint: string
        pki_id: string
    },
    timestamp: {
        unix_nano: number
        logical_time: number
    }
    ledger_height?: number,
    chaincodes?: { name: string, version: string }[]

}

export const ParsePeerResult = ({identity, membership_info, state_info}) => {


    // MEMBERSHIP - Peer.membership_info
    // gossip.Envelope.payload

    const {payload, signature, secret_envelope} = membership_info;
    assert.strictEqual(secret_envelope, null);
    const {tag, aliveMsg} = gossip.GossipMessage.deserializeBinary(payload).toObject();
    assert.strictEqual(tag, 1);
    const {membership: {endpoint, pkiId}, timestamp: {incNum, seqNum}} = aliveMsg;

    const identityObject = msp.SerializedIdentity.deserializeBinary(identity).toObject()

    // @ts-ignore TODO
    const pki_id = pkiId.toString('hex')
    assert.strictEqual(pki_id, calculatePKI_ID(identity))
    const peer: PeerResult = {
        identity: identityObject,
        membership_info: {
            endpoint,
            pki_id,
        },
        timestamp: {
            // Date.now() in nano second. 19 digits length. UnixSecond is 13 digits length
            unix_nano: incNum,
            // auto-increment as long as gossip alive in the blockchain network. (starting from 0)
            logical_time: seqNum
        },
    };

    // STATE
    if (state_info) {
        const {payload, signature, secret_envelope} = state_info;
        assert.strictEqual(secret_envelope, null);
        const {
            tag,
            stateInfo: {timestamp, pkiId, channelMac, properties}
        } = gossip.GossipMessage.deserializeBinary(payload).toObject();

        assert.strictEqual(tag, 5);
        const {chaincodesList, ledgerHeight} = properties;
        peer.ledger_height = ledgerHeight;
        peer.chaincodes = chaincodesList.map(({name, version}) => ({name, version}));
    }
    return peer;
};

interface DiscoveryResult {
    error?,
    cc_query_res?,
    config_result?,
    members?
}

// TODO WIP
// export const ParseResult = ({results}) => {
//     const returned:DiscoveryResult= {};
//
//     for (const {result, error, config_result, cc_query_res, members} of results) {
//         switch (result) {
//             case DiscoveryResultType.error:
//                 returned.error = error;
//                 break;
//             case DiscoveryResultType.cc_query_res:
//                 returned.cc_query_res = cc_query_res.content;
//                 break;
//             case DiscoveryResultType.config_result: {
//
//                 const {msps, orderers} = config_result;
//
//                 for (const [mspid, q_msp] of Object.entries(msps)) {
//                     msps[mspid] = {
//                         organizational_unit_identifiers: q_msp.organizational_unit_identifiers,
//                         root_certs: q_msp.root_certs.toString().trim(),
//                         intermediate_certs: q_msp.intermediate_certs.toString().trim(),
//                         admins: q_msp.admins.toString().trim(),
//                         tls_root_certs: q_msp.tls_root_certs.toString().trim(),
//                         tls_intermediate_certs: q_msp.tls_intermediate_certs.toString().trim()
//                     };
//                 }
//
//                 for (const [mspid, {endpoint}] of Object.entries(orderers)) {
//                     orderers[mspid] = endpoint;
//                 }
//                 returned.config_result = config_result;
//             }
//                 break;
//             case DiscoveryResultType.members: {
//                 const {peers_by_org} = members;
//
//
//                 for (const [mspid, {peers}] of Object.entries(peers_by_org)) {
//
//                     peers_by_org[mspid] = peers;
//
//                 }
//                 returned.members = peers_by_org;
//             }
//                 break;
//         }
//     }
//     return returned;
// };
