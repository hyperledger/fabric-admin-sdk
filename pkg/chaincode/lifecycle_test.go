package chaincode_test

import (
	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lifecycle", func() {
	Context("Definitions", func() {
		It("Validate ChannelName", func() {
			chaincodeDefinition := &chaincode.Definition{}
			err := chaincodeDefinition.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("Validate Name", func() {
			chaincodeDefinition := &chaincode.Definition{
				ChannelName: "mycc",
			}
			err := chaincodeDefinition.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("Validate Version", func() {
			chaincodeDefinition := &chaincode.Definition{
				Name:        "CHAINCODE",
				ChannelName: "mycc",
			}
			err := chaincodeDefinition.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("Validate missing / zero Sequence", func() {
			chaincodeDefinition := &chaincode.Definition{
				Name:        "CHAINCODE",
				Version:     "1.0",
				ChannelName: "mycc",
			}
			err := chaincodeDefinition.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("Validate negative Sequence", func() {
			chaincodeDefinition := &chaincode.Definition{
				Name:        "CHAINCODE",
				Version:     "1.0",
				Sequence:    -1,
				ChannelName: "mycc",
			}
			err := chaincodeDefinition.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("Pass validation", func() {
			chaincodeDefinition := &chaincode.Definition{
				Name:        "CHAINCODE",
				Version:     "1.0",
				Sequence:    1,
				ChannelName: "my_cc",
			}
			err := chaincodeDefinition.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
