package tools_test

import (
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/pkg/tools"

	"errors"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("Tools", func() {

	Context("Config tx gen", func() {
		It("Load Profile for config TX", func() {
			profile, err := tools.LoadProfile("TwoOrgsApplicationGenesis", "../../test/data")
			Expect(err).NotTo(HaveOccurred())
			Expect(profile).ToNot(BeNil())
			Expect(profile.Orderer.BatchSize.MaxMessageCount).To(Equal(uint32(10)))
		})

		It("Load Profile for config TX in error case", func() {
			profile, err := tools.LoadProfile("", "test/data/errorfile.yaml")
			Expect(err).To(HaveOccurred())
			Expect(profile).To(BeNil())
		})

		It("ConfigTxGen", func() {
			profile, _ := tools.LoadProfile("TwoOrgsApplicationGenesis", "../../test/data")
			block, err := tools.ConfigTxGen(profile, "mychannel")
			Expect(err).NotTo(HaveOccurred())
			Expect(block).ToNot(BeNil())
			_, err = ExtractEnvelope(block, 0)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

// ExtractEnvelope retrieves the requested envelope from a given block and
// unmarshals it
func ExtractEnvelope(block *common.Block, index int) (*common.Envelope, error) {
	if block.Data == nil {
		return nil, errors.New("block data is nil")
	}

	envelopeCount := len(block.Data.Data)
	if index < 0 || index >= envelopeCount {
		return nil, fmt.Errorf("envelope index %d out of bounds for block containing %d envelopes", index, envelopeCount)
	}
	marshaledEnvelope := block.Data.Data[index]
	return UnmarshalEnvelope(marshaledEnvelope)
}

func UnmarshalEnvelope(data []byte) (*common.Envelope, error) {
	env := &common.Envelope{}
	if err := proto.Unmarshal(data, env); err != nil {
		return nil, fmt.Errorf("error unmarshaling Envelope: %w", err)
	}
	return env, nil
}
