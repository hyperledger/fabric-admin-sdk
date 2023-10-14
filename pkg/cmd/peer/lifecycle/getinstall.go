package lifecycle

/*
	// Get installed chaincode package from each peer
	runParallel(peerConnections, func(target *ConnectionDetails) {
		ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
		defer cancel()
		result, err := chaincode.GetInstalled(ctx, target.connection, target.id, packageID)
		printGrpcError(err)
		Expect(err).NotTo(HaveOccurred(), "get installed chaincode package")
		Expect(result).NotTo(BeEmpty())
	})
*/
