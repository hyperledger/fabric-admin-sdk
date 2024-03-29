import timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb.js';
import {common, msp} from '@hyperledger/fabric-protos'
import SignatureHeader = common.SignatureHeader
import SerializedIdentity = msp.SerializedIdentity

export function currentTimestamp(): timestamp_pb.Timestamp {
    return timestamp_pb.Timestamp.fromDate(new Date())
}

export function buildSignatureHeader(params: SignatureHeader.AsObject) {
    const {creator, nonce} = params
    const header = new SignatureHeader()
    header.setCreator(creator)
    header.setNonce(nonce)
    return header
}

export function buildSerializedIdentity(params:SerializedIdentity.AsObject): SerializedIdentity {
    const serializedIdentity = new SerializedIdentity()
    serializedIdentity.setMspid(params.mspid)
    serializedIdentity.setIdBytes(params.idBytes)
    return serializedIdentity
}

