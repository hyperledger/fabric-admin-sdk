import {common, peer, msp} from '@hyperledger/fabric-protos'
import ChaincodeID = peer.ChaincodeID
import HeaderType = common.HeaderType

import {buildChannelHeader, buildHeader} from "./channel-builder.js";
import {ChannelName, TxId} from "../types.js";
import {TransientMap} from "../chaincode.js";



export function buildChaincodeID(params: ChaincodeID.AsObject): ChaincodeID {
    const _ = new ChaincodeID()
    const {name, version = '', path = ''} = params
    _.setName(name)
    _.setVersion(version)
    _.setPath(path)
    return _
}

export function buildChaincodeInput(params: peer.ChaincodeInput.AsObject): peer.ChaincodeInput {
    const {isInit, decorationsMap, argsList} = params
    const _ = new peer.ChaincodeInput()
    _.setIsInit(isInit)
    _.setArgsList(argsList)
    for (const [key, value] of decorationsMap) {
        _.getDecorationsMap().set(key, value)
    }
    return _
}

export function buildChaincodeSpec(params: peer.ChaincodeSpec.AsObject): peer.ChaincodeSpec {
    const {type, chaincodeId, input, timeout} = params
    const _ = new peer.ChaincodeSpec()
    _.setType(type)
    _.setChaincodeId(buildChaincodeID(chaincodeId))
    _.setInput(buildChaincodeInput(input))
    !!timeout && _.setTimeout(timeout)
    return _
}

export function buildChaincodeInvocationSpec(chaincodeSpec: peer.ChaincodeSpec): peer.ChaincodeInvocationSpec {
    const _ = new peer.ChaincodeInvocationSpec()
    _.setChaincodeSpec(chaincodeSpec)
    return _
}

export function buildChaincodeProposalPayload(input: peer.ChaincodeInvocationSpec, transientMap: TransientMap): peer.ChaincodeProposalPayload {
    const _ = new peer.ChaincodeProposalPayload()
    if (transientMap) {
        for (const [key, value] of Object.entries(transientMap)) {
            _.getTransientmapMap().set(key, value)
        }
    }
    _.setInput(input.serializeBinary())
    return _
}

export namespace BuildProposal {
    export function asEndorserTransaction(
        channelId: ChannelName,
        txId: TxId,
        identity: msp.SerializedIdentity.AsObject,
        nonce: Uint8Array,
        chaincodeProposalPayload: peer.ChaincodeProposalPayload): peer.Proposal {
        const type = HeaderType.ENDORSER_TRANSACTION
        const channelHeader = buildChannelHeader(type, channelId, txId)
        const header = buildHeader(identity.mspid, identity.idBytes, nonce, channelHeader)
        const _ = new peer.Proposal()
        _.setHeader(header.serializeBinary())
        _.setPayload(chaincodeProposalPayload.serializeBinary())
        return _
    }
}
