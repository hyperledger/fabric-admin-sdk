
export enum ChannelGroupType {
    /**
     * @deprecated used for system channel
     */
    system = 'Orderer',
    application = 'Application'
}

export enum PolicyName {
    Readers = 'Readers',
    Writers = 'Writers',
    Admins = 'Admins',
    BlockValidation = 'BlockValidation'
}

export const MSPRoleType = [
    'MEMBER', // Represents an MSP Member
    'ADMIN', // Represents an MSP Admin
    'CLIENT', // Represents an MSP Client
    'PEER', // Represents an MSP Peer
    'ORDERER', // Represents an MSP Orderer
];

export enum DiscoveryResultType {
    config_result = 'config_result',
    error = 'error',
    cc_query_res = 'cc_query_res',
    members = 'members'
}

export enum OrdererType {
    etcdraft = 'etcdraft'
}

export enum MetricsProvider {
    statsd = 'statsd',
    prometheus = 'prometheus',
    undefined = 'disabled',
    null = 'disabled' // value in json file could not be undefined
}

namespace ImplicitMetaPolicy {
    // ImplicitMetaPolicyType is the 'Type' string for implicit meta policies
    export const type = 'ImplicitMeta'
    export const Rule = ['ANY', 'ALL', 'MAJORITY']
}

export const TransactionType = {
    ENDORSER_TRANSACTION: 'ENDORSER_TRANSACTION',
    CONFIG: 'CONFIG'
};

/**
 * @enum
 */
export const TxValidationCode = {
    0: 'VALID',
    1: 'NIL_ENVELOPE',
    2: 'BAD_PAYLOAD',
    3: 'BAD_COMMON_HEADER',
    4: 'BAD_CREATOR_SIGNATURE',
    5: 'INVALID_ENDORSER_TRANSACTION',
    6: 'INVALID_CONFIG_TRANSACTION',
    7: 'UNSUPPORTED_TX_PAYLOAD',
    8: 'BAD_PROPOSAL_TXID',
    9: 'DUPLICATE_TXID',
    10: 'ENDORSEMENT_POLICY_FAILURE',
    11: 'MVCC_READ_CONFLICT',
    12: 'PHANTOM_READ_CONFLICT',
    13: 'UNKNOWN_TX_TYPE',
    14: 'TARGET_CHAIN_NOT_FOUND',
    15: 'MARSHAL_TX_ERROR',
    16: 'NIL_TXACTION',
    17: 'EXPIRED_CHAINCODE',
    18: 'CHAINCODE_VERSION_CONFLICT',
    19: 'BAD_HEADER_EXTENSION',
    20: 'BAD_CHANNEL_HEADER',
    21: 'BAD_RESPONSE_PAYLOAD',
    22: 'BAD_RWSET',
    23: 'ILLEGAL_WRITESET',
    24: 'INVALID_WRITESET',
    25: 'INVALID_CHAINCODE',
    254: 'NOT_VALIDATED',
    255: 'INVALID_OTHER_REASON'
};
/**
 * selected HTTP status codes
 * @enum
 */
export const Status = {
    UNKNOWN: 0,
    SUCCESS: 200,
    BAD_REQUEST: 400,
    FORBIDDEN: 403,
    NOT_FOUND: 404,
    REQUEST_ENTITY_TOO_LARGE: 413,
    INTERNAL_SERVER_ERROR: 500,
    NOT_IMPLEMENTED: 501,
    SERVICE_UNAVAILABLE: 503,
};

export enum BlockMetadataIndex {
    SIGNATURES,
    /**
     * @deprecated Do not use
     */
    LAST_CONFIG,
    TRANSACTIONS_FILTER,
    /**
     * @deprecated Do not use
     */
    ORDERER,
    COMMIT_HASH
}
