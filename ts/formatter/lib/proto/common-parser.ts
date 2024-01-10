import {msp} from '@hyperledger/fabric-protos';
import {Message} from "google-protobuf";

const {SerializedIdentity} = msp

/**
 * @param binary
 * @param asType The class of protobuf.
 */
export function decode(binary: Uint8Array|string, asType: typeof Message){
    const message = asType.deserializeBinary(binary);
    return message.toObject()
}

export function identity(bytes) {
    const _ = SerializedIdentity.deserializeBinary(bytes)
    return {
        mspid: _.getMspid(),
        idBytes: Buffer.from(_.getIdBytes()).toString()
    }
}