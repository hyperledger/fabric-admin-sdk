import {BufferFrom} from './protobuf.js';
import fabproto6 from 'fabric-protos';
import assert from 'assert';
import crypto from 'crypto';

export const sha2_256 = (data, encoding = 'hex') => crypto.createHash('sha256').update(data).digest(encoding);

export const calculateTransactionId = (signature_header) => {
	const {creator: {mspid, id_bytes}, nonce} = signature_header;
	const creator_bytes = BufferFrom({mspid, id_bytes}, fabproto6.msp.SerializedIdentity);
	const trans_bytes = Buffer.concat([nonce, creator_bytes]);
	return sha2_256(trans_bytes);
};
// utility function to create a random number of
// the specified length.
export const getNonce = (length = 24) => {
	assert.ok(Number.isInteger(length), 'Parameter must be an integer');
	return crypto.randomBytes(length);
};

