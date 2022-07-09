import {Timestamp} from 'google-protobuf/google/protobuf/timestamp_pb';
import {common, msp} from '@hyperledger/fabric-protos'
import SignatureHeader = common.SignatureHeader
import SerializedIdentity = msp.SerializedIdentity

export function currentTimestamp(): Timestamp {
    return Timestamp.fromDate(new Date())
}

export function buildSignatureHeader(creator: Uint8Array, nonce: Uint8Array) {
    const header = new SignatureHeader()
    header.setCreator(creator)
    header.setNonce(nonce)
    return header
}

export function buildSerializedIdentity(mspid: string, id_bytes: Uint8Array): SerializedIdentity {
    const serializedIdentity = new SerializedIdentity()
    serializedIdentity.setMspid(mspid)
    serializedIdentity.setIdBytes(id_bytes)
    return serializedIdentity
}

