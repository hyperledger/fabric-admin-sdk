import {ValueOf, WithConnectionInfo} from "./index";
import {Status} from './constants'

/**
 * A map with the key value pairs of the transient data.
 */
export type TransientMap = Record<string, string | Uint8Array> | object

export type ProposalResultHandler = (ProposalResponse) => ProposalResponse;

export type CommitResultHandler = (result: CommitResponse) => CommitResponse

export interface ServiceError extends Error, WithConnectionInfo {
}

export type CommitResponse = {
    status: ValueOf<typeof Status>
    info: string
}

export interface EndorsementResponse extends WithConnectionInfo {
    response: {
        status: number;
        message: string;
        payload: Buffer;
    };
    payload: Buffer;
    endorsement?: {
        endorser: Buffer;
        signature: Buffer;
    };
}

export interface ProposalResponse {
    errors: ServiceError[];
    responses: EndorsementResponse[];
}

export namespace SystemChaincode {
    namespace qscc {
        export const name = 'qscc'
        namespace functionName {
            export const GetBlockByNumber = 'GetBlockByNumber';
            export const GetChainInfo = 'GetChainInfo';
            export const GetBlockByHash = 'GetBlockByHash';
            export const GetTransactionByID = 'GetTransactionByID';
        }
    }

    namespace cscc {
        export const name = 'cscc'
        namespace functionName {
            export const JoinChain = 'JoinChain'
            export const GetChannels = 'GetChannels'
        }
    }

    namespace _lifecycle {
        export const name = '_lifecycle'
        namespace functionName {
            // InstallChaincodeFuncName is the chaincode function name used to install a chaincode
            export const InstallChaincode = 'InstallChaincode'

            // QueryInstalledChaincodeFuncName is the chaincode function name used to query SINGLE installed chaincode
            export const QueryInstalledChaincode = 'QueryInstalledChaincode'

            // QueryInstalledChaincodesFuncName is the chaincode function name used to query all installed chaincodes
            export const QueryInstalledChaincodes = 'QueryInstalledChaincodes'

            // used to approve a chaincode definition for execution by the user's own org
            export const ApproveChaincodeDefinitionForMyOrg = 'ApproveChaincodeDefinitionForMyOrg'

            // used to query a approved chaincode definition for the user's own org
            export const QueryApprovedChaincodeDefinition = 'QueryApprovedChaincodeDefinition' // TODO args and result proto message definition not found

            // used to check a specified chaincode definition is ready to be committed.
            // It returns the approval status for a given definition over a given set of orgs
            export const CheckCommitReadiness = 'CheckCommitReadiness'

            // used to 'commit' (previously 'instantiate') a chaincode in a channel.
            export const CommitChaincodeDefinition = 'CommitChaincodeDefinition'

            // used to query a committed chaincode definition in a channel.
            export const QueryChaincodeDefinition = 'QueryChaincodeDefinition'

            // used to query the committed chaincode definitions in a channel.
            export const QueryChaincodeDefinitions = 'QueryChaincodeDefinitions'
        }
    }
}

export enum ChaincodeSpecType {
    UNDEFINED,
    GOLANG,
    NODE,
    CAR,
    JAVA,
}