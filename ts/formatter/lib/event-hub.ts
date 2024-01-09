import {peer, orderer} from '@hyperledger/fabric-protos'
import DeliverResponse = peer.DeliverResponse
import SeekPosition = orderer.SeekPosition

export type DeliverResponseType = DeliverResponse.TypeCase

export type BlockNumberFilterType = SeekPosition.TypeCase
