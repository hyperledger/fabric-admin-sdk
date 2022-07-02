import {msp} from '@hyperledger/fabric-protos';

const {SerializedIdentity} = msp

export class Parser {
	identity(bytes) {
		const _ = SerializedIdentity.deserializeBinary(bytes)
		return {
			mspid: _.getMspid(),
			idBytes: Buffer.from(_.getIdBytes()).toString()
		}
	}
}
