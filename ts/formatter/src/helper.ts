import {msp} from '@hyperledger/fabric-protos';

import {BinaryToTextEncoding, createHash, randomBytes} from 'crypto';

export const sha2_256 = (data, encoding: BinaryToTextEncoding = 'hex') => createHash('sha256').update(data).digest(encoding);

export const calculateTransactionId = (signature_header) => {
    const {creator: {mspid, id_bytes}, nonce} = signature_header;
    const _ = new msp.SerializedIdentity()
    _.setMspid(mspid);
    _.setIdBytes(id_bytes)
    const creator_bytes = _.serializeBinary();
    const trans_bytes = Buffer.concat([nonce, creator_bytes]);
    return sha2_256(trans_bytes);
};
// utility function to create a random number of the specified length.
export const getNonce = (length: number = 24) => {
    return randomBytes(length);
};

