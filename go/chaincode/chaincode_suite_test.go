package chaincode_test

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChaincode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaincode Suite")
}

var tmpDir string

var _ = BeforeSuite(func() {
	tmpDir, _ = ioutil.TempDir("", "chaincode")
})

var _ = AfterSuite(func() {
	os.RemoveAll(tmpDir)
})
