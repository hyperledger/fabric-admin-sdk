import EventService from 'fabric-common/lib/EventService.js';

import {
	BlockEventFilterType,
	TxEventFilterType,
	BlockNumberFilterType
} from '@hyperledger-twgc/fabric-formatter/eventHub.js';

const {FULL_BLOCK} = BlockEventFilterType;
const {ALL} = TxEventFilterType;
const {NEWEST, OLDEST} = BlockNumberFilterType;
import fabproto6 from '@hyperledger/fabric-protos';
import Long from 'long';

const {BLOCK_UNTIL_READY, FAIL_IF_NOT_READY} = fabproto6.orderer.SeekInfo.SeekBehavior;

export default class EventHub {
	/**
	 * Does not support multiple stream managed in single EventService
	 * Only single `eventService._current_eventer` take effect
	 * `_current_eventer` is set from either one workable target
	 *
	 * @constructor
	 * @param {Client.Channel} [channel]
	 * @param {Eventer} [eventer]
	 * @param {EventService} [eventService] existing eventService object
	 * @param [options]
	 */
	constructor(channel, eventer, eventService, options = {}) {
		if (!eventService) {
			eventService = new EventService('-', channel);

			eventService.setTargets([eventer]);
		}

		this.eventService = eventService;
		this.eventOptions = options;
	}

	/**
	 *
	 * @param {IdentityContext} identityContext
	 * @param {BlockNumberFilterType|number} [startBlock]
	 * @param {BlockNumberFilterType|number} [endBlock]
	 * @param {boolean} [behavior] true to adopt FAIL_IF_NOT_READY, false to adopt BLOCK_UNTIL_READY
	 * 		BLOCK_UNTIL_READY will mean hold the stream open and keep sending as the blocks come in
	 * 		FAIL_IF_NOT_READY will mean if the block is not there throw an error
	 * @param {BlockEventFilterType} [blockType]
	 */
	build(identityContext, {startBlock, endBlock, behavior, blockType = FULL_BLOCK} = {}) {
		const {eventService, eventService: {channel}} = this;


		// build start proto
		const seekStart = fabproto6.orderer.SeekPosition.create();
		switch (startBlock) {
			case OLDEST:
				seekStart.oldest = fabproto6.orderer.SeekOldest.create();
				break;
			case NEWEST:
				seekStart.newest = fabproto6.orderer.SeekNewest.create();
				break;
			default:
				if (Number.isInteger(startBlock)) {
					seekStart.specified = fabproto6.orderer.SeekSpecified.create({
						number: startBlock
					});
				} else {
					seekStart.newest = fabproto6.orderer.SeekNewest.create();
				}
		}

		// build stop proto
		const seekStop = fabproto6.orderer.SeekPosition.create();
		switch (endBlock) {
			case OLDEST:
				seekStop.oldest = fabproto6.orderer.SeekOldest.create();
				break;
			case NEWEST:
				seekStop.newest = fabproto6.orderer.SeekNewest.create();
				break;
			default:
				seekStop.specified = fabproto6.orderer.SeekSpecified.create({
					number: Number.isInteger(endBlock) ? endBlock : Long.MAX_VALUE
				});
		}


		// seek info with all parts
		const seekInfo = fabproto6.orderer.SeekInfo.create({
			start: seekStart,
			stop: seekStop,
			behavior: behavior ? FAIL_IF_NOT_READY : BLOCK_UNTIL_READY
		});
		const seekInfoBuf = fabproto6.orderer.SeekInfo.encode(seekInfo).finish();

		// build the header for use with the seekInfo payload
		const channelHeaderBuf = channel.buildChannelHeader(
			fabproto6.common.HeaderType.DELIVER_SEEK_INFO,
			'',
			identityContext.transactionId
		);

		const seekPayload = fabproto6.common.Payload.create({
			header: eventService.buildHeader(identityContext, channelHeaderBuf),
			data: seekInfoBuf
		});
		eventService.blockType = blockType;
		eventService._payload = fabproto6.common.Payload.encode(seekPayload).finish();

		eventService.sign(identityContext);
	}

	async connect() {
		const {eventService} = this;
		await eventService.send(this.eventOptions);
	}

	disconnect() {
		this.eventService.close();
	}

	isConnected() {
		return this.eventService.isStarted();
	}

	/**
	 * @param {EventListener} listener
	 */
	unregisterEvent(listener) {
		const notThrow = true;
		this.eventService.unregisterEventListener(listener, notThrow);
	}

	/**
	 * @callback EventCallback
	 * @param {Error} error
	 * @param {EventInfo} event
	 */

	/**
	 *
	 * @param {string} chaincodeId
	 * @param eventName
	 * @param {EventCallback} callback
	 * @param {EventRegistrationOptions} [options]
	 * @return {EventListener}
	 */
	chaincodeEvent(chaincodeId, eventName, callback, options) {
		const {eventService} = this;
		if (!options) {
			options = {unregister: false, startBlock: undefined, endBlock: undefined};
		}


		return eventService.registerChaincodeListener(chaincodeId, eventName, callback, options);
	}


	/**
	 *
	 * @param {EventCallback} callback
	 * @param {EventRegistrationOptions} [options]
	 * @return {EventListener}
	 */
	blockEvent(callback, options) {
		const {eventService} = this;
		if (!options) {
			options = {unregister: false, startBlock: undefined, endBlock: undefined};
		}

		return eventService.registerBlockListener(callback, options);
	}

	/**
	 *
	 * @param {string|TxEventFilterType} [transactionID]
	 * @param {EventCallback} callback
	 * @param {EventRegistrationOptions} options
	 * @return {EventListener}
	 */
	txEvent(transactionID, callback, options) {
		const {eventService} = this;
		if (!transactionID) {
			transactionID = ALL;
		}
		return eventService.registerTransactionListener(transactionID, callback, options);
	}
}
