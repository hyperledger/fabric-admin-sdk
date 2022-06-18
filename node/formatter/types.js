/**
 * @typedef {Object} CommitResponse
 * @property {CommonResponseStatus} status
 * @property {string} info
 */
/**
 * @typedef {string} OrgName organization name (MSPName), mapping to MSP ID
 */
/**
 * @typedef {string} MspId msp_identifier, member service provider identifier
 */
/**
 * @typedef {Object|Record<string, string | Uint8Array>} TransientMap jsObject of key<string> --> value<string>
 */

/**
 * @typedef {Object} FabricCAResponse
 * @property {FabricCAResponseResult} result
 * @property {[]} errors
 * @property {[]} messages
 * @property {boolean} success
 */

/**
 * @typedef {Object} FabricCAResponseResult
 * @property {string} Credential
 * @property Attrs
 * @property {string} CRI
 * @property {string} Nonce Base64 format
 * @property {CAInfo} CAInfo
 */

/**
 * @typedef {Object} CAInfo
 * @property {string} CAName
 * @property {string} CAChain Base64 format
 * @property {string} IssuerPublicKey Base64 format
 * @property {string} IssuerRevocationPublicKey Base64 format
 * @property {string} Version
 */
/**
 * @typedef {Object} Metadata
 * @property {Object} _internal_repr
 * @property {integer} flags
 */

/**
 * An object of a fully decoded protobuf message "Block".
 * A Block may contain the configuration of the channel or endorsing transactions on the channel.
 * @typedef {Object} Block
 * @property {BlockHeader} header
 * @property {{data:BlockData[]}} data
 * @property {{metadata:BlockMetadata}} metadata
 */

/**
 * @typedef {Object} BlockHeader
 * @property {intString|Long} number
 * @property {Buffer|hexString} previous_hash
 * @property {Buffer|hexString} data_hash
 */

/**
 * @typedef {Object} BlockData
 * @property {Buffer} signature
 * @property {BlockDataPayload} payload
 */

/**
 * @typedef {Object} BlockDataPayload
 * @property {TxHeader} header
 * @property {ConfigEnvelope|EndorseTransaction} data
 */
/**
 * @typedef {integer[]} TransactionsFilter see TxValidationCode in proto/peer/transaction.proto
 */

/**
 * @typedef {Object} LastConfig
 * @property {{index:intString}} value
 * @property {MetadataSignature[]} signatures
 */

/**
 * @typedef {[{value,signatures:MetadataSignature[]},LastConfig,TransactionsFilter]} BlockMetadata Object[3]
 */

/**
 * A signature over the metadata of a block, to ensure the authenticity of
 * the metadata that describes a Block.
 * @typedef {Object} MetadataSignature
 * @property {SignatureHeader} signature_header
 * @property {Buffer} signature
 */


/**
 * Transaction Header describe basic information about a transaction record, such as its type (configuration update, or endorser transaction, etc.),
 * the id of the channel it belongs to, the transaction id and so on.
 *
 * @typedef {Object} TxHeader
 * @property {ChannelHeader} channel_header
 * @property {SignatureHeader} signature_header
 */

/**
 * @typedef {Object} ChannelHeader
 * @property {integer} type
 * @property {integer} version
 * @property {string} timestamp
 * @property {string} channel_id
 * @property {string} tx_id
 * @property {intString} epoch
 * @property {Buffer} extension
 * @property {TransactionType} typeString
 */


/**
 * @typedef {Object} TransactionCreator
 * @property {MspId} Mspid
 * @property {CertificatePem} IdBytes
 */

/**
 * An object that is part of all signatures in Hyperledger Fabric.
 * @typedef {Object} SignatureHeader
 * @property {TransactionCreator} creator
 * @property {Buffer} nonce - a unique value to guard against replay attacks.
 */


/**
 * A ConfigEnvelope contains the channel configurations data and is the main content of a configuration block.
 * Every block, including the configuration blocks themselves, has a pointer to the latest configuration block, making it easy to query for the
 * latest channel configuration settings.
 * @typedef {Object} ConfigEnvelope
 * @property {{sequence:intString,channel_group:ConfigGroup}} config
 * @property {{signature:Buffer,payload:{header:TxHeader,data:ConfigUpdateEnvelope}}} last_update
 */


/**
 * @typedef {Object} ConfigUpdateEnvelope
 * @property {{channel_id:string,read_set:ConfigGroup,write_set:ConfigGroup}} config_update
 * @property {MetadataSignature[]} signatures
 */

/**
 * A channel configuration record will have the following object structure.
 * @typedef {Object} ConfigGroup
 * @property {integer} version
 * @property {{Orderer:ConfigGroup,Application?:ConfigGroup,OrgName?:ConfigGroup}} groups
 * @property {Record<string,ConfigValue>} values usual keys:
 * - for global: Consortium|HashingAlgorithm|BlockDataHashingStructure,
 * - for Orderer: ConsensusType|BatchSize|BatchTimeout|ChannelRestrictions
 * - for organization: MSP
 * - for orderer organization: Endpoints
 * - for peer organization: AnchorPeers
 * - for all: Capabilities
 * @property {{Admins:ConfigPolicy,Readers:ConfigPolicy,Writers:ConfigPolicy,BlockValidation?:ConfigPolicy}} policies
 * @property {string} mod_policy
 */

/**
 * @typedef {Object} ConfigValue
 * @property {integer} version
 * @property {string} mod_policy
 * @property {ConfigValueBody} value
 */

/**
 * @typedef {Object} ConfigValueBody
 * @property {string} [name] used in HashingAlgorithm|Consortium
 * @property {string[]} [anchor_peers] used in AnchorPeers
 * @property {number} [width] BlockDataHashingStructure
 * @property {integer} [type] used in MSP
 * @property {MSPConfigValue} config used in MSP
 */

/**
 * @typedef {Object} MSPConfigValue
 * @property {string} name
 * @property {CertificatePem[]} root_certs
 * @property {CertificatePem[]} intermediate_certs
 * @property {CertificatePem[]} admins
 * @property {CertificatePem[]} revocation_list
 * @property signing_identity
 * @property {CertificatePem[]} organizational_unit_identifiers
 * @property {CertificatePem[]} tls_root_certs
 * @property {CertificatePem[]} tls_intermediate_certs
 *
 */


/**
 * ImplicitMetaPolicy is a policy type which depends on the hierarchical nature of the configuration
 * It is implicit because the rule is generate implicitly based on the number of sub policies with a threshold as in "ANY", "MAJORITY" or "ALL"
 * It is meta because it depends only on the result of other policies
 * <br><br>
 * When evaluated, this policy iterates over all immediate child sub-groups, retrieves the policy
 * of name sub_policy, evaluates the collection and applies the rule.
 * <br><br>
 * For example, with 4 sub-groups, and a policy name of "Readers", ImplicitMetaPolicy retrieves
 * each sub-group, retrieves policy "Readers" for each subgroup, evaluates it, and, in the case of ANY
 * 1 satisfied is sufficient, ALL would require 4 signatures, and MAJORITY would require 3 signatures.
 * @typedef {Object} ImplicitMetaPolicy
 * @property {PolicyType} type 'IMPLICIT_META'
 * @property {{sub_policy:PolicyName|string,rule:ImplicitMetaPolicyRule}} value
 */

/**
 * SignaturePolicy is a recursive message structure which defines a featherweight DSL for describing
 * policies which are more complicated than 'exactly this signature'.
 * @typedef {Object} SignaturePolicy
 * @property {PolicyType} type 'SIGNATURE'
 * @property {{version:integer, rule:SignaturePolicyRule,identities:SignaturePolicyIdentity[]}} value
 */

/**
 * @typedef {Object} SignaturePolicyIdentity
 * @property {integer} principal_classification
 * @property {MspId} msp_identifier
 * @property {MSPRoleType} Role
 */

/**
 * The NOutOf operator is sufficient to express AND as well as OR, as well as of course N out of the following M policies.
 * @typedef {Object} SignaturePolicyRule
 * @property {PolicyRuleType} Type 'n_out_of'
 * @property {{N:integer,rules:SignaturePolicyRuleSignedBy[]}} n_out_of
 */
/**
 * SignedBy implies that the signature is from a valid certificate which is signed by the trusted authority
 * @typedef {Object} SignaturePolicyRuleSignedBy
 * @property {PolicyRuleType} Type 'signed_by'
 * @property {integer} signed_by
 */

/**
 * @typedef {Object} ConfigPolicy
 * @property {integer} version
 * @property {string} mod_policy
 * @property {ImplicitMetaPolicy|SignaturePolicy} policy
 */

/**
 * An endorsement proposal, which includes the name of the chaincode to be invoked and the arguments to be passed to the chaincode.
 *
 * @typedef {Object} ChaincodeInvocationSpec
 * @property {integer} type
 * @property {ChaincodeType} typeString
 * @property {{args:Buffer[],decorations}} input
 * @property {{path:string,name:string,version:string}} chaincode_id
 * @property {integer} timeout
 */

/**
 * @typedef {Object} EndorseTransactionActionPayLoad
 * @property {{input:{chaincode_spec:ChaincodeInvocationSpec}}}} chaincode_proposal_payload
 * @property {{proposal_response_payload:EndorseTransactionProposalResponsePayload,endorsements:Endorsement[]}} action
 */

/**
 * @typedef {Object} EndorseTransactionProposalResponsePayload
 * @property {hexString} proposal_hash
 * @property {EndorseTransactionProposalResponseExtension} extension
 */

/**
 * @typedef {Object} EndorseTransactionProposalResponseExtension
 * @property {{data_model:integer,ns_rwset:ReadWriteSet[]}} results
 * @property {ChaincodeEvent} events
 * @property {Client.Response} response
 */

/**
 * @typedef {Object} ReadWriteSet
 * @property {string} namespace chaincodeName
 * @property {{reads:Read[],range_queries_info:[],writes:Write[],metadata_writes:MetadataWrite[]}} rwset
 * @property {PrivateReadWriteSet[]} collection_hashed_rwset
 */
/**
 * @typedef {Object} PrivateReadWriteSet
 * @property {string} collection_name
 * @property {HashedReadWriteSet} hashed_rwset
 * @property {Buffer} pvt_rwset_hash
 */

/**
 * @typedef {Object} HashedReadWriteSet
 * @property {HashedRead[]} hashed_reads
 * @property {HashedWrite[]} hashed_writes
 * @property {HashedMetadataWrite[]} metadata_writes
 */

/**
 * @typedef {Object} HashedRead
 * @property {Buffer} key_hash
 * @property {{block_num:intString,tx_num:intString}} version
 */

/**
 * @typedef {Object} HashedWrite
 * @property {Buffer} key_hash
 * @property {boolean} is_delete
 * @property {Buffer} value_hash
 */

/**
 * @typedef {Object} HashedMetadataWrite
 * @property key_hash //TODO
 * @property {{name:string,value}[]} entries //TODO
 */

/**
 * item of ReadSet
 * @typedef {Object} Read
 * @property {string} key
 * @property {{block_num:intString,tx_num:intString}|null} version
 */
/**
 * item of WriteSet
 * @typedef {Object} Write
 * @property {string} key
 * @property {boolean} is_delete
 * @property {string} value
 */
/**
 * item of MetadataWriteSet
 * @typedef {Object} MetadataWrite
 * @property {string} key
 * @property {{name:string,value}[]} entries //TODO
 */

/**
 * EndorseTransactionAction contains a chaincode proposal and corresponding proposal responses
 * that encapsulate the endorsing peer's decisions on whether the proposal is considered valid.
 * @typedef {Object} EndorseTransactionAction
 * @property {SignatureHeader} header
 * @property {EndorseTransactionActionPayLoad} payload
 */

/**
 * An endorsement is a signature of an endorser over a proposal response. By producing an endorsement message,
 * an endorser implicitly "approves" that proposal response and the actions contained therein. When enough endorsements have been collected,
 * a transaction can be generated out of a set of proposal responses
 *
 * @typedef {Object} Endorsement
 * @property {TransactionCreator} endorser
 * @property {Buffer} signature
 */

/**
 * An Endorser Transaction, is the result of invoking chaincodes to collect endorsements, getting globally ordered in the context of a channel,
 * and getting validated by the committer peer as part of a block before finally being formally "committed" to the ledger inside a Block.
 * <br><br>
 * Note that even if a transaction proposal(s) is considered valid by the endorsing peers, it may still be rejected by the committers during
 * transaction validation. Whether a transaction as a whole is valid or not, is not reflected in the transaction record itself,
 * but rather recorded in a separate field in the Block's metadata.

 * @typedef {Object} EndorseTransaction
 * @property {EndorseTransactionAction[]} actions These represent different steps for executing a transaction,
 * and those steps will be processed atomically, meaning if any one step failed then the whole transaction will be marked as rejected.
 */

/**
 * @typedef {Object} ExternalConnect
 * @property {string} address "your.chaincode.host.com:9999"
 * @property {string} dial_timeout “10s”, “500ms”, “1m”
 * @property {boolean} tls_required
 * @property {boolean} client_auth_required
 * @property {string} client_key used when client_auth_required
 * @property {string} client_cert used when client_auth_required
 * @property {string} root_cert used when tls_required
 */
