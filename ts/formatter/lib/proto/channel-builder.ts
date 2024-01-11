import {common, peer} from '@hyperledger/fabric-protos'
import HeaderType = common.HeaderType
import ChannelHeader = common.ChannelHeader
import Header = common.Header;
import {CertificatePEM, MspId, ValueOf, IndexDigit, ChannelName, ChaincodeLabel, TxId} from "../types.js";
import {currentTimestamp, buildSignatureHeader, buildSerializedIdentity} from './common-builder.js'

const {ChaincodeID, ChaincodeHeaderExtension} = peer

export function buildChannelHeader(
    type: ValueOf<typeof HeaderType> | IndexDigit,
    channelId: ChannelName,
    txId: TxId,
    version: IndexDigit = 1, chaincodeName?: ChaincodeLabel, TLSCertHash?: string,
    timestamp = currentTimestamp()): ChannelHeader {
    const channelHeader = new ChannelHeader();
    channelHeader.setType(type)
    channelHeader.setChannelId(channelId)
    channelHeader.setTxId(txId)
    channelHeader.setVersion(version)
    channelHeader.setEpoch(0); // uint64


    const headerExt = new ChaincodeHeaderExtension();
    if (chaincodeName) {
        const chaincodeID = new ChaincodeID();
        chaincodeID.setName(chaincodeName)
        headerExt.setChaincodeId(chaincodeID)
    }

    channelHeader.setExtension$(headerExt.serializeBinary());
    channelHeader.setTimestamp(timestamp)
    if (TLSCertHash) {
        channelHeader.setTlsCertHash(TLSCertHash);
    }

    return channelHeader;
}


export const buildHeader = (mspid: MspId,
                            certificate: Uint8Array | CertificatePEM,
                            nonce: Uint8Array,
                            channelHeader: ChannelHeader): Header => {

    const creator = buildSerializedIdentity({
        mspid,
        idBytes: typeof certificate === 'string' ? Buffer.from(certificate) : certificate
    }).serializeBinary()
    const signatureHeader = buildSignatureHeader({creator, nonce})
    const header = new Header();
    header.setSignatureHeader(signatureHeader.serializeBinary());
    header.setChannelHeader(channelHeader.serializeBinary());

    return header;
};