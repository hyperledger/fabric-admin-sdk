/**
 * @enum {string}
 */
export const ChaincodeProposalCommand = {
	deploy: 'deploy',
	upgrade: 'upgrade'
};
/**
 * @enum {string}
 */
export const ChannelGroupType = {
	system: 'Orderer',
	application: 'Application'
};
/**
 * @enum {string}
 */
export const SystemChaincodeID = {
	LSCC: 'lscc',
	QSCC: 'qscc',
	CSCC: 'cscc',
	VSCC: 'vscc',
	ESCC: 'escc',
	LifeCycle: '_lifecycle',
};
/**
 * @enum {string}
 */
export const PolicyName = {
	Readers: 'Readers',
	Writers: 'Writers',
	Admins: 'Admins',
	BlockValidation: 'BlockValidation'
};
/**
 * MSPRoleType defines which of the available, pre-defined MSP-roles
 * @enum {string}
 */
export const MSPRoleType = [
	'MEMBER', // Represents an MSP Member
	'ADMIN', // Represents an MSP Admin
	'CLIENT', // Represents an MSP Client
	'PEER', // Represents an MSP Peer
	'ORDERER', // Represents an MSP Orderer
];
/**
 * TODO
 * @enum {string}
 */
export const MspType = {
	idemix: 'idemix'
};

/**
 * @enum
 */
export const DiscoveryResultType = {
	config_result: 'config_result',
	error: 'error',
	cc_query_res: 'cc_query_res',
	members: 'members'
};


/**
 * @enum {string}
 */
export const OrdererType = {
	etcdraft: 'etcdraft'
};
/**
 *
 * @enum {string}
 */
export const MetricsProvider = {
	statsd: 'statsd',
	prometheus: 'prometheus',
	undefined: 'disabled',
	null: 'disabled' // value in json file could not be undefined
};
/**
 * @enum {string}
 */
export const ImplicitMetaPolicyRule = {
	ANY: 'ANY',
	ALL: 'ALL',
	MAJORITY: 'MAJORITY'
};
/**
 * @enum {string}
 */
export const TransactionType = {
	ENDORSER_TRANSACTION: 'ENDORSER_TRANSACTION',
	CONFIG: 'CONFIG'
};
/**
 * @enum {string}
 */
export const PolicyType = {
	IMPLICIT_META: 'IMPLICIT_META',
	SIGNATURE: 'SIGNATURE'
};
/**
 * @enum {string}
 */
export const PolicyRuleType = {
	n_out_of: 'n_out_of',
	signed_by: 'signed_by'
};
/**
 * @enum {string}
 */
export const IdentityType = {
	Role: 'role',
	OrganizationUnit: 'organization-unit',
	Identity: 'identity'
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
 * selected HTTtatus codes
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
/**
 *
 * @enum {string}
 */
export const CommonResponseStatus = {
	SUCCESS: 'SUCCESS',
	SERVICE_UNAVAILABLE: 'SERVICE_UNAVAILABLE',
	NOT_FOUND: 'NOT_FOUND'
};
/**
 *
 * @enum {string}
 */
export const BroadcastResponseStatus = {
	SUCCESS: CommonResponseStatus.SUCCESS,
	BAD_REQUEST: 'BAD_REQUEST'
};

/**
 * @enum
 */
export const BlockMetadataIndex = {
	SIGNATURES: 0,
	/**
	 * @deprecated Do not use
	 */
	LAST_CONFIG: 1,
	TRANSACTIONS_FILTER: 2,
	/**
	 * @deprecated Do not use
	 */
	ORDERER: 3,
	COMMIT_HASH: 4
};
