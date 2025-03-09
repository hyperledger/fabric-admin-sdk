import {common} from '@hyperledger/fabric-protos';
import assert from 'assert';
import {calculateTransactionId} from './helper.js';
import {BlockMetadataIndex, HeaderType} from './constants.js';
import {decode} from "./proto/common-parser.js";

const {SIGNATURES, TRANSACTIONS_FILTER, COMMIT_HASH} = BlockMetadataIndex;


export class BlockDecoder {
    block: common.Block;
    logger: Console;

    constructor(block, logger = console) {
        this.block = block
        this.logger = logger
    }

    header() {
        const header = this.block.getHeader();
        const previousHash = header.getPreviousHash_asU8();
        const dataHash = header.getDataHash_asU8();

        return {
            number: header.getNumber(),
            previousHash: Buffer.from(previousHash).toString('hex'),
            dataHash: Buffer.from(dataHash).toString('hex')
        };
    }

    data() {
        const transactions = [];
        const dataListAsU8 = this.block.getData().getDataList_asU8();
        for (const envelope of dataListAsU8) {
            const {payload} = decode(envelope, common.Envelope)
            const payloadObject: common.Payload.AsObject = decode(payload, common.Payload);
            const {channelHeader, signatureHeader} = payloadObject.header
            const channelHeaderObject: common.ChannelHeader.AsObject = decode(channelHeader, common.ChannelHeader)
            assert.strictEqual(calculateTransactionId(decode(signatureHeader, common.SignatureHeader)), channelHeaderObject.txId);
            // TODO parse data according to type
            transactions.push({type: HeaderType[channelHeaderObject.type], data: payloadObject.data})
        }
        return transactions;
    }

    metadata() {
        const metadatas = this.block.getMetadata().getMetadataList();
        assert.strictEqual(metadatas.length, 5);
        const result: Record<string, any> = {}
        for (let i = 0; i < metadatas.length; i++) {
            const metadata: common.Metadata.AsObject = decode(metadatas[i], common.Metadata);
            // TODO further parser
            switch (i) {
                case SIGNATURES:
                    result.SIGNATURES = metadata
                    break;
                case TRANSACTIONS_FILTER:
                    result.TRANSACTIONS_FILTER = metadata
                    break;
                case COMMIT_HASH:
                    result.COMMIT_HASH = metadata.value
                    break;
            }
        }
        return result;
    }
}
