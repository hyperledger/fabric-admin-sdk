/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package lifecycle_test

import (
	"context"
	"fabric-admin-sdk/pkg/chaincode"
	"os"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chaincode install", func() {
	It("Single peer", func(specCtx SpecContext) {
		packageFilePath := os.Getenv(chaincodePackageEnv)

		packageFile, err := os.Open(packageFilePath)
		Expect(err).NotTo(HaveOccurred(), "open chaincode package file")

		connection, err := newGrpcConnection()
		Expect(err).NotTo(HaveOccurred(), "gRPC connection")

		endorser := peer.NewEndorserClient(connection)

		signer, err := newSignerSerializer()
		Expect(err).NotTo(HaveOccurred(), "signer")

		installCtx, installCancel := context.WithTimeout(specCtx, 2*time.Minute)
		defer installCancel()

		err = chaincode.Install(installCtx, endorser, signer, packageFile)
		Expect(err).NotTo(HaveOccurred(), "chaincode install")

		queryCtx, queryCancel := context.WithTimeout(specCtx, 30*time.Second)
		defer queryCancel()

		result, err := chaincode.QueryInstalled(queryCtx, endorser, signer)
		Expect(err).NotTo(HaveOccurred(), "query installed chaincode")

		installedChaincodes := result.GetInstalledChaincodes()
		Expect(installedChaincodes).To(HaveLen(1), "number of installed chaincodes")
		// Expect(installedChaincodes[0].GetPackageId()).To(Equal(packageID), "installed chaincode package ID")
		Expect(installedChaincodes[0].GetLabel()).To(Equal("basic_1.0"), "installed chaincode label")
	})
})
