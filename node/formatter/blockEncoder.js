import fabricProtos from 'fabric-protos';
import {BufferFrom} from './protobuf.js';

const {common: commonProto} = fabricProtos;
export const fromEvent = ({block}) => {
	const blockHeader = new commonProto.BlockHeader();
	blockHeader.number = block.header.number;
	blockHeader.previous_hash = block.header.previous_hash;
	blockHeader.data_hash = block.header.data_hash;
	const blockData = new commonProto.BlockData();
	blockData.data = block.data.data;
	const blockMetadata = new commonProto.BlockMetadata();
	blockMetadata.metadata = block.metadata.metadata;

	return BufferFrom({header: blockHeader, data: blockData, metadata: blockMetadata}, commonProto.Block);
};

