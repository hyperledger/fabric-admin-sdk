package chaincode

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lifecycle", func() {
	Context("Definitions", func() {
		It("Validate ChannelName", func() {
			chaincodeDefinition := &Definition{}
			err := chaincodeDefinition.validate()
			Expect(err).To(HaveOccurred())
		})

		It("Validate Name", func() {
			chaincodeDefinition := &Definition{
				ChannelName: "mycc",
			}
			err := chaincodeDefinition.validate()
			Expect(err).To(HaveOccurred())
		})

		It("Validate Version", func() {
			chaincodeDefinition := &Definition{
				Name:        "CHAINCODE",
				ChannelName: "mycc",
			}
			err := chaincodeDefinition.validate()
			Expect(err).To(HaveOccurred())
		})

		It("Validate missing / zero Sequence", func() {
			chaincodeDefinition := &Definition{
				Name:        "CHAINCODE",
				Version:     "1.0",
				ChannelName: "mycc",
			}
			err := chaincodeDefinition.validate()
			Expect(err).To(HaveOccurred())
		})

		It("Validate negative Sequence", func() {
			chaincodeDefinition := &Definition{
				Name:        "CHAINCODE",
				Version:     "1.0",
				Sequence:    -1,
				ChannelName: "mycc",
			}
			err := chaincodeDefinition.validate()
			Expect(err).To(HaveOccurred())
		})

		It("Pass validation", func() {
			chaincodeDefinition := &Definition{
				Name:        "CHAINCODE",
				Version:     "1.0",
				Sequence:    1,
				ChannelName: "my_cc",
			}
			err := chaincodeDefinition.validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
