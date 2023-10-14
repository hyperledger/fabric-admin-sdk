package lifecycle

/*
	// Check chaincode commit readiness
	readinessCtx, readinessCancel := context.WithTimeout(specCtx, 30*time.Second)
	defer readinessCancel()
	readinessResult, err := chaincode.CheckCommitReadiness(readinessCtx, peer1Connection, org1MSP, chaincodeDef)
	printGrpcError(err)
	Expect(err).NotTo(HaveOccurred(), "check commit readiness")
	Expect(readinessResult.GetApprovals()).To(Equal((map[string]bool{org1MspID: true, org2MspID: true})))

*/
