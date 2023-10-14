package lifecycle

/*
	runParallel(peerConnections, func(target *ConnectionDetails) {
		ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
		defer cancel()
		result, err := chaincode.QueryInstalled(ctx, target.connection, target.id)
		printGrpcError(err)
		Expect(err).NotTo(HaveOccurred(), "query installed chaincode")
		installedChaincodes := result.GetInstalledChaincodes()
		Expect(installedChaincodes).To(HaveLen(1), "number of installed chaincodes")
		Expect(installedChaincodes[0].GetPackageId()).To(Equal(packageID), "installed chaincode package ID")
		Expect(installedChaincodes[0].GetLabel()).To(Equal(dummyMeta.Label), "installed chaincode label")
	})
*/
