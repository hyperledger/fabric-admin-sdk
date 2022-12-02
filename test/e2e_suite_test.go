package test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "e2e Suite")
}

var tmpDir string

var _ = BeforeSuite(func() {
	tmpDir = os.TempDir()
})

var _ = AfterSuite(func() {
	os.RemoveAll(tmpDir)
})
