# fabric-admin-sdk: Nodejs
A way to compensate missing admin level node-sdk

## Documents
Code as document at current stage. Welcome raise issue or create Github discussion topic. 


## Design
- It depends on fabric-sdk components
- Object-oriented

## Notes
- channel.getChannelConfig(peer) will not be rebuilt, please use event service to get block from peer; or get it from orderer
- `gatePolicy.ts`: translator for GateDSL <=> N out of
    - reference: `common/policydsl/policyparser.go`

## Components Mapping
- `peer.js` work as [endorser] and [eventer]
- `orderer.js` work as [committer], [eventer] and [discoverer]
		
### Channel create
Channel create is same as channel update action

### Channel join
peer joining to a channel in nature is sending a system chaincode proposal to peers
- See in `CSCCProposal.js`, you could extract the genesis block data from a file,    
- Or you could use method `getSpecificBlock` in `signingIdentity.js` to get genesis block from orderer

### Channel Update
See in `channelUpdate.js`, Align with fabric-common design, you could either on below as channel configuration content 
- `useEnvelope(envelope)` for cases if you have already a signed configUpdate envelop from CLI `peer channel signconfigtx ...` 
- `useSignatures(config, signatures)` for cases if you have in-memory config object, signingIdentity and signatures done by fabric-sdk-node. `All within nodejs` fashion   

### Chaincode Package
There are multiples way we could make archive. A sample use `chaincodePackage.js` in `npm khala-fabric-sdk-node`     

### Chaincode Install
See in `lifeCycleProposal.installChaincode`

### Chaincode Approve
See in `lifeCycleProposal.approveForMyOrg`

### Chaincode Commit
See in `lifeCycleProposal.commitChaincodeDefinition`

### Discovery service
SlimDiscoveryService in `discovery.js` 
- It returns the raw representation of discovery result. No further object rebuild inside.
