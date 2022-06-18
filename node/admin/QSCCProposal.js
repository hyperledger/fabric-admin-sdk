import ProposalManager from './proposal.js';
import {SystemChaincodeFunctions} from 'khala-fabric-formatter/systemChaincode.js';
import {SystemChaincodeID} from 'khala-fabric-formatter/constants.js';
import {EndorseALL} from './resultInterceptors.js';

const {qscc: {GetBlockByNumber, GetChainInfo, GetBlockByHash, GetTransactionByID}} = SystemChaincodeFunctions;
const {QSCC} = SystemChaincodeID;

export default class QSCCProposal extends ProposalManager {
	/**
	 *
	 * @param {IdentityContext} identityContext
	 * @param {Endorser[]} endorsers
	 * @param {Channel} channel
	 */
	constructor(identityContext, endorsers, channel) {
		super(identityContext, endorsers, QSCC, channel,);
		this.asQuery();
		this.resultHandler = EndorseALL
	}

	/**
	 * Block inside response.payload
	 * @param blockNumber
	 * @return {Promise<*>}
	 */
	async queryBlock(blockNumber) {

		/**
		 * @type {BuildProposalRequest}
		 */
		const buildProposalRequest = {
			fcn: GetBlockByNumber,
			args: [this.channel.name, blockNumber.toString()],
		};

		return this.send(buildProposalRequest);
	}

	async queryInfo() {
		/**
		 * @type {BuildProposalRequest}
		 */
		const buildProposalRequest = {
			fcn: GetChainInfo,
			args: [this.channel.name],
		};

		return this.send(buildProposalRequest);
	}

	/**
	 *
	 * @param {Buffer} blockHash
	 */
	async queryBlockByHash(blockHash) {
		/**
		 * @type {BuildProposalRequest}
		 */
		const buildProposalRequest = {
			fcn: GetBlockByHash,
			args: [this.channel.name, blockHash],
		};
		return this.send(buildProposalRequest);
	}

	async queryTransaction(tx_id) {
		/**
		 * @type {BuildProposalRequest}
		 */
		const buildProposalRequest = {
			fcn: GetTransactionByID,
			args: [this.channel.name, tx_id],
		};
		return this.send(buildProposalRequest);
	}
}
