import {common, peer} from '@hyperledger/fabric-protos'
import ChaincodeID = peer.ChaincodeID
import HeaderType = common.HeaderType
import {
    ChaincodeInput,
    ChaincodeInvocationSpec,
    ChaincodeProposalPayload,
    ChaincodeSpec, Proposal,
} from "@hyperledger/fabric-protos/lib/peer";
import {buildChannelHeader, buildHeader} from "./channel-builder";
import {ChannelName, TxId} from "../types";
import {SerializedIdentity} from "@hyperledger/fabric-protos/lib/msp/identities_pb";
import {TransientMap} from "../chaincode";

export function buildChaincodeID(params: ChaincodeID.AsObject): ChaincodeID {
    const _ = new ChaincodeID()
    const {name, version = '', path = ''} = params
    _.setName(name)
    _.setVersion(version)
    _.setPath(path)
    return _
}

export function buildChaincodeInput(params: ChaincodeInput.AsObject): ChaincodeInput {
    const {isInit, decorationsMap, argsList} = params
    const _ = new ChaincodeInput()
    _.setIsInit(isInit)
    _.setArgsList(argsList)
    for (const [key, value] of decorationsMap) {
        _.getDecorationsMap().set(key, value)
    }
    return _
}

export function buildChaincodeSpec(params: ChaincodeSpec.AsObject): ChaincodeSpec {
    const {type, chaincodeId, input, timeout} = params
    const _ = new ChaincodeSpec()
    _.setType(type)
    _.setChaincodeId(buildChaincodeID(chaincodeId))
    _.setInput(buildChaincodeInput(input))
    !!timeout && _.setTimeout(timeout)
    return _
}

export function buildChaincodeInvocationSpec(chaincodeSpec: ChaincodeSpec): ChaincodeInvocationSpec {
    const _ = new ChaincodeInvocationSpec()
    _.setChaincodeSpec(chaincodeSpec)
    return _
}

export function buildChaincodeProposalPayload(input: ChaincodeInvocationSpec, transientMap: TransientMap): ChaincodeProposalPayload {
    const _ = new ChaincodeProposalPayload()
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
        identity: SerializedIdentity.AsObject,
        nonce: Uint8Array,
        chaincodeProposalPayload: ChaincodeProposalPayload): Proposal {
        const type = HeaderType.ENDORSER_TRANSACTION
        const channelHeader = buildChannelHeader(type, channelId, txId)
        const header = buildHeader(identity.mspid, identity.idBytes, nonce, channelHeader)
        const _ = new Proposal()
        _.setHeader(header.serializeBinary())
        _.setPayload(chaincodeProposalPayload.serializeBinary())
        return _
    }
}
