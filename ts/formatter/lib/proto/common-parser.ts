import {msp} from '@hyperledger/fabric-protos';
import {Message} from "google-protobuf";

const {SerializedIdentity} = msp

/**
 * @param binary
 * @param asType The class of protobuf.
 */
export function decode<T extends Record<string, any>>(binary: Uint8Array|string, asType: typeof Message):T {
    const message = asType.deserializeBinary(<Uint8Array>binary);
    return <T>message.toObject()
}

export function identity(bytes: Uint8Array) {
    const _ = SerializedIdentity.deserializeBinary(bytes)
    return {
        mspid: _.getMspid(),
        idBytes: Buffer.from(_.getIdBytes()).toString()
    }
}