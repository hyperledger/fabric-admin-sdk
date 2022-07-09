import {peer, orderer} from '@hyperledger/fabric-protos'
import DeliverResponse = peer.DeliverResponse
import SeekPosition = orderer.SeekPosition

export type DeliverResponseType = DeliverResponse.TypeCase

// TODO customized, delete if not used
export enum EventListenerType {
    BLOCK = 'block', // for block type event listeners
    TX = 'tx', // for transaction type event listeners
    CHAINCODE = 'chaincode' // for chaincode event type event listeners
}
// TODO customized, delete if not used
export enum TxEventFilterType {
    ALL = 'all' // Special transaction id to indicate that the transaction listener will be notified of all transactions
}

export const BlockNumberFilterType = SeekPosition.TypeCase
export const ErrorSymptom = {
    ByClose: /^EventService has been shutdown by "close\(\)" call$/,
    OnEnd: /^fabric peer service has closed due to an "end" event$/,
    EndBlockSeen: /^Shutdown due to end block number has been seen: \d+$/,
    NewestBlockSeen: /^Newest block received:\d+ status:\w+$/,
    EndBlockOverFlow: /^End block of \d+not received. Last block received \d+$/,
    UnknownStatus: /^Event stream has received an unexpected status message. status:\w+$/,
    UNKNOWNType: /^Event stream has received an unknown response type \w+$/
};
