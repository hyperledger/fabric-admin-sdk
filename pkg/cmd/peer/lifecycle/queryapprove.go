package lifecycle

/*
	runParallel(peerConnections, func(target *ConnectionDetails) {
		ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
		defer cancel()
		result, err := chaincode.QueryApproved(ctx, target.connection, target.id, channelName, chaincodeDef.Name, chaincodeDef.Sequence)
		printGrpcError(err)
		Expect(err).NotTo(HaveOccurred(), "query approved chaincode for org %s", target.id.MspID())
		Expect(result.GetVersion()).To(Equal(chaincodeDef.Version))
	})
*/
