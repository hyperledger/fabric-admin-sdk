package channelconfig

import (
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	mspprotos "github.com/hyperledger/fabric-protos-go-apiv2/msp"
	ab "github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer/etcdraft"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer/smartbft"
	pb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
)

const (
	AdminsPolicyKey                       = "Admins"
	ReadersPolicyKey                      = "Readers"
	WritersPolicyKey                      = "Writers"
	OrdererGroupKey                       = "Orderer"
	ApplicationGroupKey                   = "Application"
	ConsortiumsGroupKey                   = "Consortiums"
	ConsortiumKey                         = "Consortium"
	HashingAlgorithmKey                   = "HashingAlgorithm"
	defaultHashingAlgorithm               = "SHA256"
	BlockDataHashingStructureKey          = "BlockDataHashingStructure"
	defaultBlockDataHashingStructureWidth = math.MaxUint32
	OrdererAddressesKey                   = "OrdererAddresses"
	CapabilitiesKey                       = "Capabilities"
	BatchSizeKey                          = "BatchSize"
	BatchTimeoutKey                       = "BatchTimeout"
	ChannelRestrictionsKey                = "ChannelRestrictions"
	KafkaBrokersKey                       = "KafkaBrokers"
	ConsensusTypeKey                      = "ConsensusType"
	MSPKey                                = "MSP"
	EndpointsKey                          = "Endpoints"
	ACLsKey                               = "ACLs"
	AnchorPeersKey                        = "AnchorPeers"
	ChannelCreationPolicyKey              = "ChannelCreationPolicy"
	OrderersKey                           = "Orderers"
)

// ConfigValue defines a common representation for different *cb.ConfigValue values.
type ConfigValue interface {
	// Key is the key this value should be stored in the *cb.ConfigGroup.Values map.
	Key() string

	// Value is the message which should be marshaled to opaque bytes for the *cb.ConfigValue.value.
	Value() proto.Message
}

// StandardConfigValue implements the ConfigValue interface.
type StandardConfigValue struct {
	key   string
	value proto.Message
}

// Key is the key this value should be stored in the *cb.ConfigGroup.Values map.
func (scv *StandardConfigValue) Key() string {
	return scv.key
}

// Value is the message which should be marshaled to opaque bytes for the *cb.ConfigValue.value.
func (scv *StandardConfigValue) Value() proto.Message {
	return scv.value
}

// HashingAlgorithmValue returns the only currently valid hashing algorithm.
// It is a value for the /Channel group.
func HashingAlgorithmValue() *StandardConfigValue {
	return &StandardConfigValue{
		key: HashingAlgorithmKey,
		value: &common.HashingAlgorithm{
			Name: defaultHashingAlgorithm,
		},
	}
}

// BlockDataHashingStructureValue returns the only currently valid block data hashing structure.
// It is a value for the /Channel group.
func BlockDataHashingStructureValue() *StandardConfigValue {
	return &StandardConfigValue{
		key: BlockDataHashingStructureKey,
		value: &common.BlockDataHashingStructure{
			Width: defaultBlockDataHashingStructureWidth,
		},
	}
}

// OrdererAddressesValue returns the a config definition for the orderer addresses.
// It is a value for the /Channel group.
func OrdererAddressesValue(addresses []string) *StandardConfigValue {
	return &StandardConfigValue{
		key: OrdererAddressesKey,
		value: &common.OrdererAddresses{
			Addresses: addresses,
		},
	}
}

// ConsortiumValue returns the config definition for the consortium name.
// It is a value for the channel group.
func ConsortiumValue(name string) *StandardConfigValue {
	return &StandardConfigValue{
		key: ConsortiumKey,
		value: &common.Consortium{
			Name: name,
		},
	}
}

// CapabilitiesValue returns the config definition for a a set of capabilities.
// It is a value for the /Channel/Orderer, Channel/Application/, and /Channel groups.
func CapabilitiesValue(capabilities map[string]bool) *StandardConfigValue {
	c := &common.Capabilities{
		Capabilities: make(map[string]*common.Capability),
	}

	for capability, required := range capabilities {
		if !required {
			continue
		}
		c.Capabilities[capability] = &common.Capability{}
	}

	return &StandardConfigValue{
		key:   CapabilitiesKey,
		value: c,
	}
}

// BatchSizeValue returns the config definition for the orderer batch size.
// It is a value for the /Channel/Orderer group.
func BatchSizeValue(maxMessages, absoluteMaxBytes, preferredMaxBytes uint32) *StandardConfigValue {
	return &StandardConfigValue{
		key: BatchSizeKey,
		value: &ab.BatchSize{
			MaxMessageCount:   maxMessages,
			AbsoluteMaxBytes:  absoluteMaxBytes,
			PreferredMaxBytes: preferredMaxBytes,
		},
	}
}

// BatchTimeoutValue returns the config definition for the orderer batch timeout.
// It is a value for the /Channel/Orderer group.
func BatchTimeoutValue(timeout string) *StandardConfigValue {
	return &StandardConfigValue{
		key: BatchTimeoutKey,
		value: &ab.BatchTimeout{
			Timeout: timeout,
		},
	}
}

// ChannelRestrictionsValue returns the config definition for the orderer channel restrictions.
// It is a value for the /Channel/Orderer group.
func ChannelRestrictionsValue(maxChannelCount uint64) *StandardConfigValue {
	return &StandardConfigValue{
		key: ChannelRestrictionsKey,
		value: &ab.ChannelRestrictions{
			MaxCount: maxChannelCount,
		},
	}
}

// ConsensusTypeValue returns the config definition for the orderer consensus type.
// It is a value for the /Channel/Orderer group.
func ConsensusTypeValue(consensusType string, consensusMetadata []byte) *StandardConfigValue {
	return &StandardConfigValue{
		key: ConsensusTypeKey,
		value: &ab.ConsensusType{
			Type:     consensusType,
			Metadata: consensusMetadata,
		},
	}
}

// MSPValue returns the config definition for an MSP.
// It is a value for the /Channel/Orderer/*, /Channel/Application/*, and /Channel/Consortiums/*/*/* groups.
func MSPValue(mspDef *mspprotos.MSPConfig) *StandardConfigValue {
	return &StandardConfigValue{
		key:   MSPKey,
		value: mspDef,
	}
}

func OrderersValue(consenters []*common.Consenter) *StandardConfigValue {
	o := &common.Orderers{
		ConsenterMapping: consenters,
	}
	return &StandardConfigValue{
		key:   OrderersKey,
		value: o,
	}
}

// MarshalBFTOptions serializes smartbft options.
func MarshalBFTOptions(op *smartbft.Options) ([]byte, error) {
	if copyMd, ok := proto.Clone(op).(*smartbft.Options); ok {
		return proto.Marshal(copyMd)
	} else {
		return nil, errors.New("consenter options type mismatch")
	}
}

// EndpointsValue returns the config definition for the orderer addresses at an org scoped level.
// It is a value for the /Channel/Orderer/<OrgName> group.
func EndpointsValue(addresses []string) *StandardConfigValue {
	return &StandardConfigValue{
		key: EndpointsKey,
		value: &common.OrdererAddresses{
			Addresses: addresses,
		},
	}
}

// ACLValues returns the config definition for an applications resources based ACL definitions.
// It is a value for the /Channel/Application/.
func ACLValues(acls map[string]string) *StandardConfigValue {
	a := &pb.ACLs{
		Acls: make(map[string]*pb.APIResource),
	}

	for apiResource, policyRef := range acls {
		a.Acls[apiResource] = &pb.APIResource{PolicyRef: policyRef}
	}

	return &StandardConfigValue{
		key:   ACLsKey,
		value: a,
	}
}

// AnchorPeersValue returns the config definition for an org's anchor peers.
// It is a value for the /Channel/Application/*.
func AnchorPeersValue(anchorPeers []*pb.AnchorPeer) *StandardConfigValue {
	return &StandardConfigValue{
		key:   AnchorPeersKey,
		value: &pb.AnchorPeers{AnchorPeers: anchorPeers},
	}
}

// ChannelCreationPolicyValue returns the config definition for a consortium's channel creation policy
// It is a value for the /Channel/Consortiums/*/*.
func ChannelCreationPolicyValue(policy *common.Policy) *StandardConfigValue {
	return &StandardConfigValue{
		key:   ChannelCreationPolicyKey,
		value: policy,
	}
}

// MarshalEtcdRaftMetadata serializes etcd RAFT metadata.
func MarshalEtcdRaftMetadata(md *etcdraft.ConfigMetadata) ([]byte, error) {
	copyMd := proto.Clone(md).(*etcdraft.ConfigMetadata)
	for _, c := range copyMd.GetConsenters() {
		// Expect the user to set the config value for client/server certs to the
		// path where they are persisted locally, then load these files to memory.
		clientCert, err := os.ReadFile(string(c.GetClientTlsCert()))
		if err != nil {
			return nil, fmt.Errorf("cannot load client cert for consenter %s:%d: %w", c.GetHost(), c.GetPort(), err)
		}
		c.ClientTlsCert = clientCert

		serverCert, err := os.ReadFile(string(c.GetServerTlsCert()))
		if err != nil {
			return nil, fmt.Errorf("cannot load server cert for consenter %s:%d: %w", c.GetHost(), c.GetPort(), err)
		}
		c.ServerTlsCert = serverCert
	}
	return proto.Marshal(copyMd)
}
