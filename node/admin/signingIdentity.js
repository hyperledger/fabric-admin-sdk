import fabprotos from '@hyperledger/fabric-protos';
import {buildChannelHeader, buildHeader, buildPayload, buildSeekPayload} from '@hyperledger-twgc/fabric-formatter/protoTranslator.js';
import {DeliverResponseStatus, DeliverResponseType} from '@hyperledger-twgc/fabric-formatter/eventHub.js';
import {BufferFrom, ProtoFrom} from '@hyperledger-twgc/fabric-formatter/protobuf.js';

const {SERVICE_UNAVAILABLE} = DeliverResponseStatus;
const {STATUS} = DeliverResponseType;
const commonProto = fabprotos.common;
const sleep = (ms) => {
	return new Promise(resolve => setTimeout(resolve, ms));
};

export default class SigningIdentityUtil {
	/**
	 *
	 * @param {SigningIdentity} signingIdentity
	 */
	constructor(signingIdentity) {
		this.signingIdentity = signingIdentity;
	}

	/**
	 *
	 * @param config
	 * @param {Buffer} nonce 24 bits random bytes
	 * @return {ConfigSignature}
	 */
	signChannelConfig(config, nonce) {

		const {signingIdentity} = this;

		const signature_header = BufferFrom({creator: signingIdentity.serialize(), nonce}, commonProto.SignatureHeader);

		// get all the bytes to be signed together, then sign
		const signing_bytes = Buffer.concat([signature_header, config]);
		const signature = Buffer.from(signingIdentity.sign(signing_bytes));

		// build the return object
		return ProtoFrom({signature_header, signature}, commonProto.ConfigSignature);
	}

	/**
	 * @typedef {Object} ChannelConfigurationUpdateContent
	 * @property {Buffer} [config] config_update of commonProto.ConfigUpdateEnvelope
	 * @property {Buffer[]} [signatures]
	 * @property {Buffer<commonProto.Envelope>} [envelope]
	 * @property [name] ChannelId
	 */


	/**
	 *
	 * Channel configuration updates can be sent to the orderers to be processed.
	 * The orderer ensures channel updates will be made only when enough signatures are discovered in the request.
	 * Channel creation policy can be customized when the consortium is defined.
	 * @param transactionId
	 * @param nonce
	 * @param [config]
	 * @param {ChannelConfigurationUpdateContent} channelConfigurationUpdateContent
	 * @param {Committer} committer
	 * @param [commitTimeout]
	 */
	async updateChannel({transactionId, nonce}, channelConfigurationUpdateContent, committer, commitTimeout) {
		const {signingIdentity} = this;
		const {config, signatures, envelope, name} = channelConfigurationUpdateContent;
		let signature, payload;
		if (envelope) {
			const envelopeDecoded = commonProto.Envelope.decode(envelope);
			signature = envelopeDecoded.signature;
			payload = envelopeDecoded.payload;
		} else {

			const configUpdateEnvelope = ProtoFrom({config_update: config, signatures}, commonProto.ConfigUpdateEnvelope);
			const channelHeader = buildChannelHeader({
				Type: commonProto.HeaderType.CONFIG_UPDATE,
				ChannelId: name,
				TxId: transactionId
			});

			const header = buildHeader({
				Creator: signingIdentity.serialize(),
				Nonce: nonce,
				ChannelHeader: channelHeader
			});
			payload = buildPayload({
				Header: header,
				Data: BufferFrom(configUpdateEnvelope),
			}, true);
			signature = Buffer.from(signingIdentity.sign(payload));
		}

		return await committer.sendBroadcast({
			signature,
			payload
		}, commitTimeout);

	}

	async getSpecificBlock({transactionId, nonce}, ChannelId, orderer, blockHeight, opts = {}) {
		const {signingIdentity} = this;
		const payload = buildSeekPayload({
			Creator: signingIdentity.serialize(),
			Nonce: nonce,
			ChannelId,
			TxId: transactionId,
		}, blockHeight, blockHeight, undefined, true);
		const signature = Buffer.from(signingIdentity.sign(payload));

		const {waitIfUNAVAILABLE, requestTimeout} = opts;

		const sendTry = async () => {
			try {
				const result = await orderer.sendDeliver({signature, payload}, requestTimeout);
				return result[0];
			} catch (e) {
				if (waitIfUNAVAILABLE) {
					const {status, Type, block} = e;
					if (status === SERVICE_UNAVAILABLE && Type === STATUS && block === null) {
						await sleep(waitIfUNAVAILABLE);
						return sendTry();
					}
				}
				throw e;
			}
		};
		return await sendTry();

	}
}
