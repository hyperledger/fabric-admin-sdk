export const SystemChaincodeFunctions = {
	qscc: {
		GetBlockByNumber: 'GetBlockByNumber',
		GetChainInfo: 'GetChainInfo',
		GetBlockByHash: 'GetBlockByHash',
		GetTransactionByID: 'GetTransactionByID',
	},
	cscc: {
		JoinChain: 'JoinChain',
		GetChannels: 'GetChannels',
	},
	_lifecycle: {
		// InstallChaincodeFuncName is the chaincode function name used to install a chaincode
		InstallChaincode: 'InstallChaincode',

		// QueryInstalledChaincodeFuncName is the chaincode function name used to query SINGLE installed chaincode
		QueryInstalledChaincode: 'QueryInstalledChaincode',

		// QueryInstalledChaincodesFuncName is the chaincode function name used to query all installed chaincodes
		QueryInstalledChaincodes: 'QueryInstalledChaincodes',

		// used to approve a chaincode definition for execution by the user's own org
		ApproveChaincodeDefinitionForMyOrg: 'ApproveChaincodeDefinitionForMyOrg',

		// used to query a approved chaincode definition for the user's own org
		QueryApprovedChaincodeDefinition: 'QueryApprovedChaincodeDefinition', // TODO args and result proto message definition not found

		// used to check a specified chaincode definition is ready to be committed.
		// It returns the approval status for a given definition over a given set of orgs
		CheckCommitReadiness: 'CheckCommitReadiness',

		// used to 'commit' (previously 'instantiate') a chaincode in a channel.
		CommitChaincodeDefinition: 'CommitChaincodeDefinition',

		// used to query a committed chaincode definition in a channel.
		QueryChaincodeDefinition: 'QueryChaincodeDefinition',

		// used to query the committed chaincode definitions in a channel.
		QueryChaincodeDefinitions: 'QueryChaincodeDefinitions',

	}
};
