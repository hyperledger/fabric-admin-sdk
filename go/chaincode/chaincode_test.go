package chaincode_test

import (
	"archive/tar"
	"fabric-admin-sdk/chaincode"
	"io"
	"log"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chaincode", func() {

	It("Chaincode package ccaas", func() {
		dummyConnection := chaincode.Connection{
			Address:      "127.0.0.1:8080",
			Dial_timeout: "10s",
			Tls_required: false,
		}
		dummyMeta := chaincode.Metadata{
			Type:  "ccaas",
			Label: "basic-asset",
		}
		str, err := chaincode.PackageCCAAS(dummyConnection, dummyMeta)
		Expect(err).NotTo(HaveOccurred())
		// so far no plan to verify the file
		file, err := os.OpenFile(tmpDir+"/chaincode.tar.gz", os.O_CREATE|os.O_WRONLY, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
		_, err = file.WriteString(str)
		Expect(err).NotTo(HaveOccurred())
		file.Close()

		file, err = os.Open(tmpDir + "/chaincode.tar.gz")
		if err != nil {
			log.Fatalln(err)
		}
		defer file.Close()
		tr := tar.NewReader(file)
		i := 0
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			Expect(err).NotTo(HaveOccurred())
			if i == 0 {
				Expect(hdr.Name).To(Equal("code.tar.gz"))
			}
			if i == 1 {
				Expect(hdr.Name).To(Equal("metadata.json"))
			}
			i++
		}
	})

	PIt("Chaincode package java", func() {})
	PIt("Chaincode package golang", func() {})
	PIt("Chaincode package nodejs", func() {})
	PIt("Chaincode install", func() {})
	PIt("Chaincode queryinstalled", func() {})
	PIt("Chaincode getinstalledpackage", func() {})
	PIt("Chaincode calculatepackageid", func() {})
	PIt("Chaincode approveformyorg", func() {})
	PIt("Chaincode queryapproved", func() {})
	PIt("Chaincode checkcommitreadiness", func() {})
	PIt("Chaincode commit", func() {})
	PIt("Chaincode querycommitted", func() {})
})
