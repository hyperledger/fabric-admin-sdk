package lifecycle

/*
time.Sleep(time.Duration(20) * time.Second)
			PolicyStr := "AND ('Org1MSP.peer','Org2MSP.peer')"
			applicationPolicy, err := chaincode.NewApplicationPolicy(PolicyStr, "")
			Expect(err).NotTo(HaveOccurred())
			chaincodeDef := &chaincode.Definition{
				ChannelName:       channelName,
				PackageID:         "",
				Name:              "basic",
				Version:           "1.0",
				EndorsementPlugin: "",
				ValidationPlugin:  "",
				Sequence:          1,
				ApplicationPolicy: applicationPolicy,
				InitRequired:      false,
				Collections:       nil,
			}
			Expect(err).NotTo(HaveOccurred())
			// Approve chaincode for each org
			runParallel(peerConnections, func(target *ConnectionDetails) {
				ctx, cancel := context.WithTimeout(specCtx, 30*time.Second)
				defer cancel()
				err := chaincode.Approve(ctx, target.connection, target.id, chaincodeDef)
				printGrpcError(err)
				Expect(err).NotTo(HaveOccurred(), "approve chaincode for org %s", target.id.MspID())
			})
*/
