/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package lifecycle_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const chaincodePackageEnv = "CHAINCODE_PACKAGE"

func TestChaincode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaincode lifecycle suite")
}

var tmpDir string

var _ = BeforeSuite(func() {
	tmpDir = os.TempDir()
})

var _ = AfterSuite(func() {
	os.RemoveAll(tmpDir)
})
