import {msp, common} from '@hyperledger/fabric-protos';
import assert from 'assert';
import {calculateTransactionId} from './helper';
import {BlockMetadataIndex} from './constants';
import LastConfig = common.LastConfig
import Metadata = common.Metadata

const {SIGNATURES, TRANSACTIONS_FILTER, LAST_CONFIG, ORDERER, COMMIT_HASH} = BlockMetadataIndex;

/**
 * TODO not completed yet
 */
export class BlockDecoder {
    block;
    logger = console;

    header() {
        const {header} = this.block;
        const {number, previous_hash, data_hash} = header;

        return {
            number: number.toInt(),
            previous_hash: previous_hash.toString('hex'),
            data_hash: data_hash.toString('hex')
        };
    }

    data() {
        const txs = [];
        const {data: {data: data}} = this.block;
        for (const entry of data) {
            const {channel_header, signature_header} = entry.payload.header;
            // assert.strictEqual(calculateTransactionId(signature_header), channel_header.tx_id);

            const {config, actions} = entry.payload.data;
            if (config) {
                this.logger.info('a config transaction');
            } else if (actions) {
                assert.strictEqual(actions.length, 1);

                const {payload, header} = actions[0];
                const {creator: {mspid, id_bytes}, nonce} = header;
                const {chaincode_proposal_payload, action} = payload;
                const {proposal_response_payload, endorsements} = action;
                for (const {endorser, signature} of endorsements) {
                    const {mspid, id_bytes} = endorser;
                }
                const {proposal_hash, extension} = proposal_response_payload;
                const {results, events, response, chaincode_id} = extension;
                const {chaincode_spec} = chaincode_proposal_payload.input;
                const {chaincode_id: {name}, type, typeString, input: {args, decorations, is_init}} = chaincode_spec;
                if (name === '_lifecycle') {
                    this.logger.info('a chaincode lifecycle transaction');
                } else {
                    this.logger.info(`a application transaction on [${name}]`);
                }
                txs.push(Object.assign({
                    tx_id: channel_header.tx_id, args, is_init, chaincode_id: name
                }, signature_header));

            } else {
                assert.fail('unknown transaction type found');
            }
        }
        return [data, txs];
    }

    metadata() {
        const {metadata: {metadata}} = this.block;
        assert.strictEqual(metadata.length, 5);
        const {value, signatures} = metadata[SIGNATURES];

        for (const {signature_header, signature} of signatures) {
            this.logger.info({signature_header, signature});
        }
        metadata[SIGNATURES] = {value: {index: LastConfig.deserializeBinary(value).getIndex()}, signatures};
        const [flag] = metadata[TRANSACTIONS_FILTER];
        const buf = metadata[COMMIT_HASH];

        const _metadata = Metadata.deserializeBinary(buf);
        const _signatures = _metadata.getSignaturesList()
        assert.ok(Array.isArray(_signatures) && _signatures.length === 0);
        _metadata[COMMIT_HASH] = {commitHash: _metadata.getValue_asB64()};

        assert.deepStrictEqual(_metadata[LAST_CONFIG], {});
        assert.deepStrictEqual(_metadata[ORDERER], {});
        assert.strictEqual(flag, 0);
        return _metadata;
    }
}
