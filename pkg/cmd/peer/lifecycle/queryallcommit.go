package lifecycle

/*
	// Query all committed chaincode
	committedCtx, committedCancel := context.WithTimeout(specCtx, 30*time.Second)
	defer committedCancel()
	committedResult, err := chaincode.QueryCommitted(committedCtx, peer1Connection, org1MSP, channelName)
	printGrpcError(err)
	Expect(err).NotTo(HaveOccurred(), "query all committed chaincodes")
	committedChaincodes := committedResult.GetChaincodeDefinitions()
	Expect(committedChaincodes).To(HaveLen(1), "number of committed chaincodes")
	Expect(committedChaincodes[0].GetName()).To(Equal("basic"), "committed chaincode name")
	Expect(committedChaincodes[0].GetSequence()).To(Equal(int64(1)), "committed chaincode sequence")
*/
