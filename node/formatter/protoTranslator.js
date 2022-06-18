import fabprotos from 'fabric-protos';
import {BlockNumberFilterType} from './eventHub.js';
import {BufferFrom, ProtoFrom} from './protobuf.js';

const commonProto = fabprotos.common;
const ordererProto = fabprotos.orderer;
const protosProto = fabprotos.protos;
const {NEWEST, OLDEST} = BlockNumberFilterType;
export const buildCurrentTimestamp = () => {
	const now = new Date();
	const timestamp = new fabprotos.google.protobuf.Timestamp();
	timestamp.seconds = now.getTime() / 1000;
	timestamp.nanos = (now.getTime() % 1000) * 1000000;
	return timestamp;
};
/**
 *
 * @param Type
 * @param Version
 * @param ChannelId
 * @param TxId
 * @param [ChaincodeID]
 * @param [TLSCertHash]
 * @param [Timestamp]
 */
export const buildChannelHeader = ({Type, Version = 1, ChannelId, TxId, ChaincodeID, TLSCertHash, Timestamp}) => {
	const channelHeader = new commonProto.ChannelHeader();
	channelHeader.type = Type; // int32
	channelHeader.version = Version; // int32

	channelHeader.channel_id = ChannelId; // string
	channelHeader.tx_id = TxId; // string
	// 	channelHeader.setEpoch(epoch); // uint64


	const headerExt = new protosProto.ChaincodeHeaderExtension();
	if (ChaincodeID) {
		const chaincodeID = new protosProto.ChaincodeID();
		chaincodeID.name = ChaincodeID;
		headerExt.chaincode_id = chaincodeID;
	}

	channelHeader.extension = BufferFrom(headerExt);
	channelHeader.timestamp = Timestamp || buildCurrentTimestamp(); // google.protobuf.Timestamp
	if (TLSCertHash) {
		channelHeader.tls_cert_hash = TLSCertHash;
	}

	return channelHeader;
};

/**
 *
 * @param Creator from Identity.js#serialize
 * @param Nonce
 * @param ChannelHeader
 */
export const buildHeader = ({Creator, Nonce, ChannelHeader}) => {
	const signatureHeaderBytes = BufferFrom({creator: Creator, nonce: Nonce}, commonProto.SignatureHeader);

	const header = new commonProto.Header();
	header.signature_header = signatureHeaderBytes;
	header.channel_header = BufferFrom(ChannelHeader, commonProto.ChannelHeader);

	return header;
};
/**
 * TODO not need anymore
 * @param {commonProto.Header} Header
 * @param {Buffer} Data
 * @param {boolean} [asBuffer]
 * @return {commonProto.Payload}
 */
export const buildPayload = ({Header, Data}, asBuffer) => {
	const payload = ProtoFrom({header: Header, data: Data}, commonProto.Payload);

	if (asBuffer) {
		return BufferFrom(payload);
	}
	return payload;
};
/**
 *
 * @param {number|BlockNumberFilterType} heightFilter
 * @return {ordererProto.SeekPosition}
 */
export const buildSeekPosition = (heightFilter) => {
	const seekPosition = new ordererProto.SeekPosition();

	switch (typeof heightFilter) {
		case 'number': {
			const seekSpecified = new ordererProto.SeekSpecified();
			seekSpecified.number = heightFilter;
			seekPosition.specified = seekSpecified;
		}
			break;
		case 'string':
			switch (heightFilter) {
				case NEWEST: {
					seekPosition.newest = new ordererProto.SeekNewest();
				}
					break;
				case OLDEST: {
					seekPosition.oldest = new ordererProto.SeekOldest();
				}
					break;
			}
			break;
	}
	return seekPosition;
};
/**
 * @enum {string}
 */
export const SeekBehavior = {
	BLOCK_UNTIL_READY: 'BLOCK_UNTIL_READY',
	FAIL_IF_NOT_READY: 'FAIL_IF_NOT_READY',
};
/**
 *
 * @param {ordererProto.SeekPosition} startSeekPosition
 * @param {ordererProto.SeekPosition} stopSeekPosition
 * @param {SeekBehavior|string} [behavior]
 */
export const buildSeekInfo = (startSeekPosition, stopSeekPosition, behavior, asBuffer) => {
	const seekInfo = ProtoFrom({start: startSeekPosition, stop: stopSeekPosition}, ordererProto.SeekInfo);
	if (behavior) {
		seekInfo.behavior = ordererProto.SeekInfo.SeekBehavior[behavior];
	}
	seekInfo.error_response = ordererProto.SeekInfo.SeekErrorResponse.STRICT;
	if (asBuffer) {
		return BufferFrom(seekInfo);
	}
	return seekInfo;
};

/**
 * @enum {number}
 */
const HeaderType = {
	MESSAGE: 0,                     // Used for messages which are signed but opaque
	CONFIG: 1,                      // Used for messages which express the channel config
	CONFIG_UPDATE: 2,               // Used for transactions which update the channel config
	ENDORSER_TRANSACTION: 3,        // Used by the SDK to submit endorser based transactions
	ORDERER_TRANSACTION: 4,         // Used internally by the orderer for management
	DELIVER_SEEK_INFO: 5,          // Used as the type for Envelope messages submitted to instruct the Deliver API to seek
	CHAINCODE_PACKAGE: 6,           // Used for packaging chaincode artifacts for install
	PEER_ADMIN_OPERATION: 8,        // Used for invoking an administrative operation on a peer
};

/**
 *
 * @param Creator
 * @param Nonce
 * @param ChannelId
 * @param TxId
 * @param startHeight
 * @param stopHeight
 * @param {SeekBehavior|string} [behavior]
 * @return {commonProto.Payload}
 */
export const buildSeekPayload = ({Creator, Nonce, ChannelId, TxId}, startHeight, stopHeight, behavior = SeekBehavior.FAIL_IF_NOT_READY, asBuffer) => {

	const startPosition = buildSeekPosition(startHeight);
	const stopPosition = buildSeekPosition(stopHeight);
	const seekInfoBytes = buildSeekInfo(startPosition, stopPosition, behavior, true);


	const seekInfoHeader = buildChannelHeader({
		Type: HeaderType.DELIVER_SEEK_INFO,
		ChannelId,
		TxId,
	});

	const seekHeader = buildHeader({Creator, Nonce, ChannelHeader: seekInfoHeader});

	return buildPayload({Header: seekHeader, Data: seekInfoBytes}, asBuffer);

};
export const extractLastConfigIndex = (block) => {
	const metadata = commonProto.Metadata.decode(block.metadata.metadata[commonProto.BlockMetadataIndex.LAST_CONFIG]); // TODO it shows as deprecated in hyperledger/fabric-protos
	const lastConfig = commonProto.LastConfig.decode(metadata.value);
	return parseInt(lastConfig.index);
};
/**
 * Extracts the protobuf 'ConfigUpdate' object out of the 'ConfigEnvelope' object
 * @param {Buffer} configEnvelope - channel config file content
 */
export const extractConfigUpdate = (configEnvelope) => {
	const envelope = commonProto.Envelope.decode(configEnvelope);
	const payload = commonProto.Payload.decode(envelope.payload);
	const configtx = commonProto.ConfigUpdateEnvelope.decode(payload.data);
	return configtx.config_update;
};

/**
 *
 * @param {BlockData} blockData
 */
export const extractConfigEnvelopeFromBlockData = (blockData) => {
	const envelope = commonProto.Envelope.decode(blockData);
	const payload = commonProto.Payload.decode(envelope.payload);
	return commonProto.ConfigEnvelope.decode(payload.data);
};

export const assertConfigBlock = (block) => {
	if (block.data.data.length !== 1) {
		throw new Error('Config block must only contain one transaction');
	}
	const envelope = commonProto.Envelope.decode(block.data.data[0]);
	const payload = commonProto.Payload.decode(envelope.payload);
	const channel_header = commonProto.ChannelHeader.decode(payload.header.channel_header);
	if (channel_header.type !== HeaderType.CONFIG) {
		throw new Error(`Block must be of type "CONFIG" , but got "${HeaderType[channel_header.type]}" instead`);
	}

};
