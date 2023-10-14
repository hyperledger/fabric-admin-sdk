package lifecycle

/*
	// Query named committed chaincode
	committedWithNameCtx, committedWithNameCancel := context.WithTimeout(specCtx, 30*time.Second)
	defer committedWithNameCancel()
	committedWithNameResult, err := chaincode.QueryCommittedWithName(committedWithNameCtx, peer1Connection, org1MSP, channelName, chaincodeDef.Name)
	printGrpcError(err)
	Expect(err).NotTo(HaveOccurred(), "query committed chaincode with name")
	Expect(committedWithNameResult.GetApprovals()).To(Equal(map[string]bool{org1MspID: true, org2MspID: true}), "committed chaincode approvals")
	Expect(committedWithNameResult.GetSequence()).To(Equal(chaincodeDef.Sequence), "committed chaincode sequence")
*/
