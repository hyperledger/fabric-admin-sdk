import {BinaryToTextEncoding, createHash, randomBytes} from 'crypto';
import {IndexDigit, TxId} from "./types.js";
import {BinaryLike} from "node:crypto";
import {common, msp} from "@hyperledger/fabric-protos";

export function sha2_256(data: BinaryLike, encoding: BinaryToTextEncoding = 'hex') {
    return createHash('sha256').update(data).digest(encoding);
}

export function calculateTransactionId(signature_header: common.SignatureHeader.AsObject): TxId {
    const {creator: creator_bytes, nonce} = signature_header;

    const trans_bytes = Buffer.concat([<Uint8Array>nonce, <Uint8Array>creator_bytes]);
    return sha2_256(trans_bytes);
}

/**
 * pki_id is a digest(sha256) of [mspID, IdBytes] from a peer.
 * See in Fabric core code `GetPKIidOfCert(peerIdentity api.PeerIdentityType) common.PKIidType`
 * @param identity
 */
export function calculatePKI_ID(identity: msp.SerializedIdentity) {
    return sha2_256(Buffer.concat([Buffer.from(identity.getMspid()), identity.getIdBytes_asU8()]))
}

// utility function to create a random number of the specified length.
export function getNonce(length: IndexDigit = 24): Buffer {
    return randomBytes(length);
}

/*
 * Make sure there's a start line with '-----BEGIN CERTIFICATE-----'
 * and end line with '-----END CERTIFICATE-----', so as to be compliant
 * with x509 parsers
 */
export function normalizeX509(raw: string) {
    const regex = /(-----\s*BEGIN ?[^-]+?-----)([\s\S]*)(-----\s*END ?[^-]+?-----)/;
    const matches = raw.match(regex);
    if (!matches || matches.length !== 4) {
        throw new Error('Failed to find start line or end line of the certificate.');
    }

    // remove the first element that is the whole match
    matches.shift();
    // remove LF or CR
    const trimmedMatches = matches.map((element) => {
        return element.trim();
    });

    // make sure '-----BEGIN CERTIFICATE-----' and '-----END CERTIFICATE-----' are in their own lines
    // and that it ends in a new line
    let result = trimmedMatches.join('\n') + '\n';
    // could be this has multiple certs within that are not separated by a newline
    const regex2 = /----------/;
    result = result.replace(new RegExp(regex2, 'g'), '-----\n-----');
    return result;
}
