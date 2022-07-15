import {Timestamp} from 'google-protobuf/google/protobuf/timestamp_pb';
import {common, msp} from '@hyperledger/fabric-protos'
import SignatureHeader = common.SignatureHeader
import SerializedIdentity = msp.SerializedIdentity

export function currentTimestamp(): Timestamp {
    return Timestamp.fromDate(new Date())
}

export function buildSignatureHeader(params: SignatureHeader.AsObject) {
    const {creator, nonce} = params
    const header = new SignatureHeader()
    header.setCreator(creator)
    header.setNonce(nonce)
    return header
}

export function buildSerializedIdentity(params:SerializedIdentity.AsObject): SerializedIdentity {
    const {mspid, idBytes} = params
    const serializedIdentity = new SerializedIdentity()
    serializedIdentity.setMspid(mspid)
    serializedIdentity.setIdBytes(idBytes)
    return serializedIdentity
}

