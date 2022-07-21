import {
    CommitResultHandler,
    ProposalResultHandler,
    TransientMap,
} from "@hyperledger-twgc/fabric-formatter/lib/chaincode";
import {
    buildChaincodeInput,
    buildChaincodeInvocationSpec,
    buildChaincodeProposalPayload, buildChaincodeSpec,
    BuildProposal
} from "@hyperledger-twgc/fabric-formatter/lib/proto/proposal-builder"
import {ChannelName, IndexDigit, TxId} from "@hyperledger-twgc/fabric-formatter/lib/index";
import {IdentityContext} from '@hyperledger-twgc/fabric-formatter/lib/user'
import {ChaincodeLabel} from "@hyperledger-twgc/fabric-formatter/lib/names";
import {calculateTransactionId} from '@hyperledger-twgc/fabric-formatter/lib/helper'

export type BuildProposalRequest = {
    /**
     * The function name used in chaincode
     */
    fcn?: string,
    /**
     * The arguments needed by the chaincode execution.
     */
    args?: string[] | Uint8Array[],
    transientMap?: TransientMap,
    /**
     * @deprecated Indicate whether this proposal is a chaincode initialization request
     */
    init?: boolean,
    txid?: TxId,
    /**
     * if specified, use this nonce instead of to generate randomly
     */
    nonce?: Uint8Array,
}


export class Proposal {

    identityContext: IdentityContext
    endorsers
    label: ChaincodeLabel
    channelName: ChannelName
    #assertProposalResult: ProposalResultHandler;
    #assertCommitResult: CommitResultHandler;

    constructor(identityContext, endorsers, chaincodeLabel, channel) {
        this.identityContext = identityContext
        this.endorsers = endorsers
        this.channelName = channel
        this.label = chaincodeLabel
    }

    set resultAssert(assertFunction: ProposalResultHandler) {
        this.#assertProposalResult = assertFunction;
    }

    set commitResultAssert(assertFunction: CommitResultHandler) {
        this.#assertCommitResult = assertFunction;
    }

    build(buildProposalRequest: BuildProposalRequest) {
        const {txid, nonce, transientMap} = buildProposalRequest

        const input = buildChaincodeInput()
        const invoke = buildChaincodeSpec({
            type, chaincodeId, input
        })

        const chaincodeProposalPayload = buildChaincodeProposalPayload(buildChaincodeInvocationSpec(invoke), transientMap)
        // build the proposal payload
        const payload = BuildProposal.asEndorserTransaction(
            this.channelName,
            txid,
            this.identityContext,
            nonce,
            chaincodeProposalPayload,
        ).serializeBinary()
        return payload
    }

    async send(buildProposalRequest: BuildProposalRequest, requestTimeout?: IndexDigit) {

        const {identityContext} = this;
        const {nonce} = buildProposalRequest;
        let {txid} = buildProposalRequest;

        if (nonce && !txid) {
            const {mspid, idBytes} = identityContext
            txid = calculateTransactionId(mspid, Buffer.from(idBytes), nonce)

        }
        this.build(identityContext, buildProposalRequest);
        this.sign(identityContext); // TODO take care of offline signing
        /**
         * @type {SendProposalRequest}
         */
        const sendProposalRequest = {
            targets: this.endorsers,
            requestTimeout,
        };
        const results = await super.send(sendProposalRequest);
        if (this.#assertProposalResult) {
            return this.#assertProposalResult(results)
        }

        return results;
    }
}