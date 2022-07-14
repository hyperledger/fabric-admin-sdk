import {common, peer} from '@hyperledger/fabric-protos'
import HeaderType = common.HeaderType
import ChannelHeader = common.ChannelHeader
import Header = common.Header;
import {CertificatePEM, MspId, ValueOf, IndexDigit} from "../d";
import {currentTimestamp, buildSignatureHeader, buildSerializedIdentity} from './common-builder'

const {ChaincodeID, ChaincodeHeaderExtension} = peer

export function buildChannelHeader(
    type: ValueOf<typeof HeaderType> | IndexDigit,
    channelId: string, txId: string, version: IndexDigit = 1, chaincodeName?: string, TLSCertHash?: string,
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
                            certificate: CertificatePEM, nonce: Uint8Array,
                            channelHeader: ChannelHeader): Header => {

    const creator = buildSerializedIdentity(mspid, Buffer.from(certificate)).serializeBinary()
    const signatureHeader = buildSignatureHeader(creator, nonce)
    const header = new Header();
    header.setSignatureHeader(signatureHeader.serializeBinary());
    header.setChannelHeader(channelHeader.serializeBinary());

    return header;
};