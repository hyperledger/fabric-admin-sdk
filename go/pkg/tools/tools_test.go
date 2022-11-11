package tools_test

import (
	"fabric-admin-sdk/pkg/tools"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/orderer/etcdraft"
	"github.com/hyperledger/fabric/bccsp/sw"
	"github.com/hyperledger/fabric/common/channelconfig"
	"github.com/hyperledger/fabric/protoutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tools", func() {

	Context("Config tx gen", func() {
		It("Load Profile for config TX", func() {
			profile, err := tools.LoadProfile("TwoOrgsApplicationGenesis", "../../testdata")
			Expect(err).NotTo(HaveOccurred())
			Expect(profile).ToNot(BeNil())
			Expect(profile.Orderer.BatchSize.MaxMessageCount).To(Equal(uint32(10)))
		})

		It("Load Profile for config TX in error case", func() {
			profile, err := tools.LoadProfile("", "testdata/errorfile.yaml")
			Expect(err).To(HaveOccurred())
			Expect(profile).To(BeNil())
		})

		PIt("ConfigTxGen", func() {
			profile, err := tools.LoadProfile("TwoOrgsApplicationGenesis", "../testdata")
			Expect(err).NotTo(HaveOccurred())
			Expect(profile).ToNot(BeNil())
			block, err := tools.ConfigTxGen(profile, "mychannel")
			Expect(err).NotTo(HaveOccurred())
			Expect(block).ToNot(BeNil())
			envelopeConfig, err := protoutil.ExtractEnvelope(block, 0)
			Expect(err).NotTo(HaveOccurred())
			cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
			Expect(err).NotTo(HaveOccurred())
			bundle, err := channelconfig.NewBundleFromEnvelope(envelopeConfig, cryptoProvider)
			Expect(err).NotTo(HaveOccurred())
			oc, exists := bundle.OrdererConfig()
			Expect(exists).To(BeTrue())
			configMetadata := &etcdraft.ConfigMetadata{}
			proto.Unmarshal(oc.ConsensusMetadata(), configMetadata)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMetadata.Options).NotTo(BeNil())
		})
	})

	PIt("gate policy", func() {})
	PIt("peer discovery", func() {})
	PIt("generate connection profile for sdk", func() {})
})
