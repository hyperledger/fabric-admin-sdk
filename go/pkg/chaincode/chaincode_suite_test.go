package chaincode_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChaincode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaincode Suite")
}

const processProposalMethod = "/protos.Endorser/ProcessProposal"

var tmpDir string

var _ = BeforeSuite(func() {
	//tmpDir, _ = ioutil.TempDir("", "chaincode")
	tmpDir = "/tmp/chiancode"
})

var _ = AfterSuite(func() {
	//os.RemoveAll(tmpDir)
})
