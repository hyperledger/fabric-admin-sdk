/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package encoder

import (
	"errors"
	"fmt"

	icc "github.com/hyperledger/fabric-admin-sdk/internal/channelconfig"
	"github.com/hyperledger/fabric-admin-sdk/internal/configtxgen/genesisconfig"
	"github.com/hyperledger/fabric-admin-sdk/internal/configtxlator/update"
	"github.com/hyperledger/fabric-admin-sdk/internal/genesis"
	"github.com/hyperledger/fabric-admin-sdk/internal/msp"
	"github.com/hyperledger/fabric-admin-sdk/internal/pkg/identity"
	ipc "github.com/hyperledger/fabric-admin-sdk/internal/policies"
	ipd "github.com/hyperledger/fabric-admin-sdk/internal/policydsl"
	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-admin-sdk/internal/util"

	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	pb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
)

const (
	ordererAdminsPolicyName = "/Channel/Orderer/Admins"

	msgVersion = int32(0)
	epoch      = 0
)

const (
	// ConsensusTypeSolo identifies the solo consensus implementation.
	ConsensusTypeSolo = "solo"
	// ConsensusTypeKafka identifies the Kafka-based consensus implementation.
	ConsensusTypeKafka = "kafka"
	// ConsensusTypeKafka identifies the Kafka-based consensus implementation.
	ConsensusTypeEtcdRaft = "etcdraft"

	// BlockValidationPolicyKey TODO
	BlockValidationPolicyKey = "BlockValidation"

	// OrdererAdminsPolicy is the absolute path to the orderer admins policy
	OrdererAdminsPolicy = "/Channel/Orderer/Admins"

	// SignaturePolicyType is the 'Type' string for signature policies
	SignaturePolicyType = "Signature"

	// ImplicitMetaPolicyType is the 'Type' string for implicit meta policies
	ImplicitMetaPolicyType = "ImplicitMeta"
)

func addValue(cg *cb.ConfigGroup, value icc.ConfigValue, modPolicy string) {
	cg.Values[value.Key()] = &cb.ConfigValue{
		Value:     protoutil.MarshalOrPanic(value.Value()),
		ModPolicy: modPolicy,
	}
}

func addPolicy(cg *cb.ConfigGroup, policy ipc.ConfigPolicy, modPolicy string) {
	cg.Policies[policy.Key()] = &cb.ConfigPolicy{
		Policy:    policy.Value(),
		ModPolicy: modPolicy,
	}
}

func AddOrdererPolicies(cg *cb.ConfigGroup, policyMap map[string]*genesisconfig.Policy, modPolicy string) error {
	switch {
	case policyMap == nil:
		return fmt.Errorf("no policies defined")
	case policyMap[BlockValidationPolicyKey] == nil:
		return fmt.Errorf("no BlockValidation policy defined")
	}

	return AddPolicies(cg, policyMap, modPolicy)
}

func AddPolicies(cg *cb.ConfigGroup, policyMap map[string]*genesisconfig.Policy, modPolicy string) error {
	switch {
	case policyMap == nil:
		return fmt.Errorf("no policies defined")
	case policyMap[icc.AdminsPolicyKey] == nil:
		return fmt.Errorf("no Admins policy defined")
	case policyMap[icc.ReadersPolicyKey] == nil:
		return fmt.Errorf("no Readers policy defined")
	case policyMap[icc.WritersPolicyKey] == nil:
		return fmt.Errorf("no Writers policy defined")
	}

	for policyName, policy := range policyMap {
		switch policy.Type {
		case ImplicitMetaPolicyType:
			imp, err := ipc.ImplicitMetaFromString(policy.Rule)
			if err != nil {
				return fmt.Errorf("invalid implicit meta policy rule '%s' %w", policy.Rule, err)
			}
			cg.Policies[policyName] = &cb.ConfigPolicy{
				ModPolicy: modPolicy,
				Policy: &cb.Policy{
					Type:  int32(cb.Policy_IMPLICIT_META),
					Value: protoutil.MarshalOrPanic(imp),
				},
			}
		case SignaturePolicyType:
			sp, err := ipd.FromString(policy.Rule)
			if err != nil {
				return fmt.Errorf("invalid signature policy rule '%s' %w", policy.Rule, err)
			}
			cg.Policies[policyName] = &cb.ConfigPolicy{
				ModPolicy: modPolicy,
				Policy: &cb.Policy{
					Type:  int32(cb.Policy_SIGNATURE),
					Value: protoutil.MarshalOrPanic(sp),
				},
			}
		default:
			return fmt.Errorf("unknown policy type: %s", policy.Type)
		}
	}
	return nil
}

func NewConfigGroup() *cb.ConfigGroup {
	return &cb.ConfigGroup{
		Groups:   make(map[string]*cb.ConfigGroup),
		Values:   make(map[string]*cb.ConfigValue),
		Policies: make(map[string]*cb.ConfigPolicy),
	}
}

// NewChannelGroup defines the root of the channel configuration.  It defines basic operating principles like the hashing
// algorithm used for the blocks, as well as the location of the ordering service.  It will recursively call into the
// NewOrdererGroup, NewConsortiumsGroup, and NewApplicationGroup depending on whether these sub-elements are set in the
// configuration.  All mod_policy values are set to "Admins" for this group, with the exception of the OrdererAddresses
// value which is set to "/Channel/Orderer/Admins".
func NewChannelGroup(conf *genesisconfig.Profile) (*cb.ConfigGroup, error) {
	channelGroup := NewConfigGroup()
	if err := AddPolicies(channelGroup, conf.Policies, icc.AdminsPolicyKey); err != nil {
		return nil, fmt.Errorf("error adding policies to channel group %w", err)
	}

	addValue(channelGroup, icc.HashingAlgorithmValue(), icc.AdminsPolicyKey)
	addValue(channelGroup, icc.BlockDataHashingStructureValue(), icc.AdminsPolicyKey)
	if conf.Orderer != nil && len(conf.Orderer.Addresses) > 0 {
		addValue(channelGroup, icc.OrdererAddressesValue(conf.Orderer.Addresses), ordererAdminsPolicyName)
	}

	if conf.Consortium != "" {
		addValue(channelGroup, icc.ConsortiumValue(conf.Consortium), icc.AdminsPolicyKey)
	}

	if len(conf.Capabilities) > 0 {
		addValue(channelGroup, icc.CapabilitiesValue(conf.Capabilities), icc.AdminsPolicyKey)
	}

	var err error
	if conf.Orderer != nil {
		channelGroup.Groups[icc.OrdererGroupKey], err = NewOrdererGroup(conf.Orderer)
		if err != nil {
			return nil, fmt.Errorf("could not create orderer group %w", err)
		}
	}

	if conf.Application != nil {
		channelGroup.Groups[icc.ApplicationGroupKey], err = NewApplicationGroup(conf.Application)
		if err != nil {
			return nil, fmt.Errorf("could not create application group %w", err)
		}
	}

	if conf.Consortiums != nil {
		channelGroup.Groups[icc.ConsortiumsGroupKey], err = NewConsortiumsGroup(conf.Consortiums)
		if err != nil {
			return nil, fmt.Errorf("could not create consortiums group %w", err)
		}
	}

	channelGroup.ModPolicy = icc.AdminsPolicyKey
	return channelGroup, nil
}

// NewOrdererGroup returns the orderer component of the channel configuration.  It defines parameters of the ordering service
// about how large blocks should be, how frequently they should be emitted, etc. as well as the organizations of the ordering network.
// It sets the mod_policy of all elements to "Admins".  This group is always present in any channel configuration.
func NewOrdererGroup(conf *genesisconfig.Orderer) (*cb.ConfigGroup, error) {
	ordererGroup := NewConfigGroup()
	if err := AddOrdererPolicies(ordererGroup, conf.Policies, icc.AdminsPolicyKey); err != nil {
		return nil, fmt.Errorf("error adding policies to orderer group %w", err)
	}
	addValue(ordererGroup, icc.BatchSizeValue(
		conf.BatchSize.MaxMessageCount,
		conf.BatchSize.AbsoluteMaxBytes,
		conf.BatchSize.PreferredMaxBytes,
	), icc.AdminsPolicyKey)
	addValue(ordererGroup, icc.BatchTimeoutValue(conf.BatchTimeout.String()), icc.AdminsPolicyKey)
	addValue(ordererGroup, icc.ChannelRestrictionsValue(conf.MaxChannels), icc.AdminsPolicyKey)

	if len(conf.Capabilities) > 0 {
		addValue(ordererGroup, icc.CapabilitiesValue(conf.Capabilities), icc.AdminsPolicyKey)
	}

	var consensusMetadata []byte
	var err error

	switch conf.OrdererType {
	case ConsensusTypeSolo:
	case ConsensusTypeKafka:
		addValue(ordererGroup, icc.KafkaBrokersValue(conf.Kafka.Brokers), icc.AdminsPolicyKey)
	case ConsensusTypeEtcdRaft:
		if consensusMetadata, err = icc.MarshalEtcdRaftMetadata(conf.EtcdRaft); err != nil {
			return nil, fmt.Errorf("cannot marshal metadata for orderer type %s: %s", ConsensusTypeEtcdRaft, err)
		}
	default:
		return nil, fmt.Errorf("unknown orderer type: %s", conf.OrdererType)
	}

	addValue(ordererGroup, icc.ConsensusTypeValue(conf.OrdererType, consensusMetadata), icc.AdminsPolicyKey)

	for _, org := range conf.Organizations {
		var err error
		ordererGroup.Groups[org.Name], err = NewOrdererOrgGroup(org)
		if err != nil {
			return nil, fmt.Errorf("failed to create orderer org %w", err)
		}
	}

	ordererGroup.ModPolicy = icc.AdminsPolicyKey
	return ordererGroup, nil
}

// NewConsortiumsGroup returns an org component of the channel configuration.  It defines the crypto material for the
// organization (its MSP).  It sets the mod_policy of all elements to "Admins".
func NewConsortiumOrgGroup(conf *genesisconfig.Organization) (*cb.ConfigGroup, error) {
	consortiumsOrgGroup := NewConfigGroup()
	consortiumsOrgGroup.ModPolicy = icc.AdminsPolicyKey

	if conf.SkipAsForeign {
		return consortiumsOrgGroup, nil
	}

	mspConfig, err := msp.GetVerifyingMspConfig(conf.MSPDir, conf.ID, conf.MSPType)
	if err != nil {
		return nil, fmt.Errorf("1 - Error loading MSP configuration for org: %s %w", conf.Name, err)
	}

	if err := AddPolicies(consortiumsOrgGroup, conf.Policies, icc.AdminsPolicyKey); err != nil {
		return nil, fmt.Errorf("error adding policies to consortiums org group '%s' %w", conf.Name, err)
	}

	addValue(consortiumsOrgGroup, icc.MSPValue(mspConfig), icc.AdminsPolicyKey)

	return consortiumsOrgGroup, nil
}

// NewOrdererOrgGroup returns an orderer org component of the channel configuration.  It defines the crypto material for the
// organization (its MSP).  It sets the mod_policy of all elements to "Admins".
func NewOrdererOrgGroup(conf *genesisconfig.Organization) (*cb.ConfigGroup, error) {
	ordererOrgGroup := NewConfigGroup()
	ordererOrgGroup.ModPolicy = icc.AdminsPolicyKey

	if conf.SkipAsForeign {
		return ordererOrgGroup, nil
	}

	mspConfig, err := msp.GetVerifyingMspConfig(conf.MSPDir, conf.ID, conf.MSPType)
	if err != nil {
		return nil, fmt.Errorf("1 - Error loading MSP configuration for org: %s %w", conf.Name, err)
	}

	if err := AddPolicies(ordererOrgGroup, conf.Policies, icc.AdminsPolicyKey); err != nil {
		return nil, fmt.Errorf("error adding policies to orderer org group '%s' %w", conf.Name, err)
	}

	addValue(ordererOrgGroup, icc.MSPValue(mspConfig), icc.AdminsPolicyKey)

	if len(conf.OrdererEndpoints) > 0 {
		addValue(ordererOrgGroup, icc.EndpointsValue(conf.OrdererEndpoints), icc.AdminsPolicyKey)
	}

	return ordererOrgGroup, nil
}

// NewApplicationGroup returns the application component of the channel configuration.  It defines the organizations which are involved
// in application logic like chaincodes, and how these members may interact with the orderer.  It sets the mod_policy of all elements to "Admins".
func NewApplicationGroup(conf *genesisconfig.Application) (*cb.ConfigGroup, error) {
	applicationGroup := NewConfigGroup()
	if err := AddPolicies(applicationGroup, conf.Policies, icc.AdminsPolicyKey); err != nil {
		return nil, fmt.Errorf("error adding policies to application group %w ", err)
	}

	if len(conf.ACLs) > 0 {
		addValue(applicationGroup, icc.ACLValues(conf.ACLs), icc.AdminsPolicyKey)
	}

	if len(conf.Capabilities) > 0 {
		addValue(applicationGroup, icc.CapabilitiesValue(conf.Capabilities), icc.AdminsPolicyKey)
	}

	for _, org := range conf.Organizations {
		var err error
		applicationGroup.Groups[org.Name], err = NewApplicationOrgGroup(org)
		if err != nil {
			return nil, fmt.Errorf("failed to create application org %w ", err)
		}
	}

	applicationGroup.ModPolicy = icc.AdminsPolicyKey
	return applicationGroup, nil
}

// NewApplicationOrgGroup returns an application org component of the channel configuration.  It defines the crypto material for the organization
// (its MSP) as well as its anchor peers for use by the gossip network.  It sets the mod_policy of all elements to "Admins".
func NewApplicationOrgGroup(conf *genesisconfig.Organization) (*cb.ConfigGroup, error) {
	applicationOrgGroup := NewConfigGroup()
	applicationOrgGroup.ModPolicy = icc.AdminsPolicyKey

	if conf.SkipAsForeign {
		return applicationOrgGroup, nil
	}

	mspConfig, err := msp.GetVerifyingMspConfig(conf.MSPDir, conf.ID, conf.MSPType)
	if err != nil {
		return nil, fmt.Errorf("1 - Error loading MSP configuration for org %s %w", conf.Name, err)
	}

	if err := AddPolicies(applicationOrgGroup, conf.Policies, icc.AdminsPolicyKey); err != nil {
		return nil, fmt.Errorf("error adding policies to application org group %s %w", conf.Name, err)
	}
	addValue(applicationOrgGroup, icc.MSPValue(mspConfig), icc.AdminsPolicyKey)

	var anchorProtos []*pb.AnchorPeer
	for _, anchorPeer := range conf.AnchorPeers {
		anchorProtos = append(anchorProtos, &pb.AnchorPeer{
			Host: anchorPeer.Host,
			Port: int32(anchorPeer.Port),
		})
	}

	// Avoid adding an unnecessary anchor peers element when one is not required.  This helps
	// prevent a delta from the orderer system channel when computing more complex channel
	// creation transactions
	if len(anchorProtos) > 0 {
		addValue(applicationOrgGroup, icc.AnchorPeersValue(anchorProtos), icc.AdminsPolicyKey)
	}

	return applicationOrgGroup, nil
}

// NewConsortiumsGroup returns the consortiums component of the channel configuration.  This element is only defined for the ordering system channel.
// It sets the mod_policy for all elements to "/Channel/Orderer/Admins".
func NewConsortiumsGroup(conf map[string]*genesisconfig.Consortium) (*cb.ConfigGroup, error) {
	consortiumsGroup := NewConfigGroup()
	// This policy is not referenced anywhere, it is only used as part of the implicit meta policy rule at the channel level, so this setting
	// effectively degrades control of the ordering system channel to the ordering admins
	addPolicy(consortiumsGroup, ipc.SignaturePolicy(icc.AdminsPolicyKey, ipd.AcceptAllPolicy), ordererAdminsPolicyName)

	for consortiumName, consortium := range conf {
		var err error
		consortiumsGroup.Groups[consortiumName], err = NewConsortiumGroup(consortium)
		if err != nil {
			return nil, fmt.Errorf("failed to create consortium %s %w", consortiumName, err)
		}
	}

	consortiumsGroup.ModPolicy = ordererAdminsPolicyName
	return consortiumsGroup, nil
}

// NewConsortiums returns a consortiums component of the channel configuration.  Each consortium defines the organizations which may be involved in channel
// creation, as well as the channel creation policy the orderer checks at channel creation time to authorize the action.  It sets the mod_policy of all
// elements to "/Channel/Orderer/Admins".
func NewConsortiumGroup(conf *genesisconfig.Consortium) (*cb.ConfigGroup, error) {
	consortiumGroup := NewConfigGroup()

	for _, org := range conf.Organizations {
		var err error
		consortiumGroup.Groups[org.Name], err = NewConsortiumOrgGroup(org)
		if err != nil {
			return nil, fmt.Errorf("failed to create consortium org %w", err)
		}
	}

	addValue(consortiumGroup, icc.ChannelCreationPolicyValue(ipc.ImplicitMetaAnyPolicy(icc.AdminsPolicyKey).Value()), ordererAdminsPolicyName)

	consortiumGroup.ModPolicy = ordererAdminsPolicyName
	return consortiumGroup, nil
}

// NewChannelCreateConfigUpdate generates a ConfigUpdate which can be sent to the orderer to create a new channel.  Optionally, the channel group of the
// ordering system channel may be passed in, and the resulting ConfigUpdate will extract the appropriate versions from this file.
func NewChannelCreateConfigUpdate(channelID string, conf *genesisconfig.Profile, templateConfig *cb.ConfigGroup) (*cb.ConfigUpdate, error) {
	if conf.Application == nil {
		return nil, errors.New("cannot define a new channel with no Application section")
	}

	if conf.Consortium == "" {
		return nil, errors.New("cannot define a new channel with no Consortium value")
	}

	newChannelGroup, err := NewChannelGroup(conf)
	if err != nil {
		return nil, fmt.Errorf("could not turn parse profile into channel group %w", err)
	}

	updt, err := update.Compute(&cb.Config{ChannelGroup: templateConfig}, &cb.Config{ChannelGroup: newChannelGroup})
	if err != nil {
		return nil, fmt.Errorf("could not compute update %w", err)
	}

	// Add the consortium name to create the channel for into the write set as required.
	updt.ChannelId = channelID
	updt.ReadSet.Values[icc.ConsortiumKey] = &cb.ConfigValue{Version: 0}
	updt.WriteSet.Values[icc.ConsortiumKey] = &cb.ConfigValue{
		Version: 0,
		Value: protoutil.MarshalOrPanic(&cb.Consortium{
			Name: conf.Consortium,
		}),
	}

	return updt, nil
}

// DefaultConfigTemplate generates a config template based on the assumption that
// the input profile is a channel creation template and no system channel context
// is available.
func DefaultConfigTemplate(conf *genesisconfig.Profile) (*cb.ConfigGroup, error) {
	channelGroup, err := NewChannelGroup(conf)
	if err != nil {
		return nil, fmt.Errorf("error parsing configuration %w", err)
	}

	if _, ok := channelGroup.Groups[icc.ApplicationGroupKey]; !ok {
		return nil, errors.New("channel template configs must contain an application section")
	}

	channelGroup.Groups[icc.ApplicationGroupKey].Values = nil
	channelGroup.Groups[icc.ApplicationGroupKey].Policies = nil

	return channelGroup, nil
}

func ConfigTemplateFromGroup(conf *genesisconfig.Profile, cg *cb.ConfigGroup) (*cb.ConfigGroup, error) {
	template := proto.Clone(cg).(*cb.ConfigGroup)
	if template.Groups == nil {
		return nil, fmt.Errorf("supplied system channel group has no sub-groups")
	}

	template.Groups[icc.ApplicationGroupKey] = &cb.ConfigGroup{
		Groups: map[string]*cb.ConfigGroup{},
		Policies: map[string]*cb.ConfigPolicy{
			icc.AdminsPolicyKey: {},
		},
	}

	consortiums, ok := template.Groups[icc.ConsortiumsGroupKey]
	if !ok {
		return nil, fmt.Errorf("supplied system channel group does not appear to be system channel (missing consortiums group)")
	}

	if consortiums.Groups == nil {
		return nil, fmt.Errorf("system channel consortiums group appears to have no consortiums defined")
	}

	consortium, ok := consortiums.Groups[conf.Consortium]
	if !ok {
		return nil, fmt.Errorf("supplied system channel group is missing '%s' consortium", conf.Consortium)
	}

	if conf.Application == nil {
		return nil, fmt.Errorf("supplied channel creation profile does not contain an application section")
	}

	for _, organization := range conf.Application.Organizations {
		var ok bool
		template.Groups[icc.ApplicationGroupKey].Groups[organization.Name], ok = consortium.Groups[organization.Name]
		if !ok {
			return nil, fmt.Errorf("consortium %s does not contain member org %s", conf.Consortium, organization.Name)
		}
	}
	delete(template.Groups, icc.ConsortiumsGroupKey)

	addValue(template, icc.ConsortiumValue(conf.Consortium), icc.AdminsPolicyKey)

	return template, nil
}

// MakeChannelCreationTransaction is a handy utility function for creating transactions for channel creation.
// It assumes the invoker has no system channel context so ignores all but the application section.
func MakeChannelCreationTransaction(
	channelID string,
	signer identity.SignerSerializer,
	conf *genesisconfig.Profile,
) (*cb.Envelope, error) {
	template, err := DefaultConfigTemplate(conf)
	if err != nil {
		return nil, fmt.Errorf("could not generate default config template %w", err)
	}
	return MakeChannelCreationTransactionFromTemplate(channelID, signer, conf, template)
}

// MakeChannelCreationTransactionWithSystemChannelContext is a utility function for creating channel creation txes.
// It requires a configuration representing the orderer system channel to allow more sophisticated channel creation
// transactions modifying pieces of the configuration like the orderer set.
func MakeChannelCreationTransactionWithSystemChannelContext(
	channelID string,
	signer identity.SignerSerializer,
	conf,
	systemChannelConf *genesisconfig.Profile,
) (*cb.Envelope, error) {
	cg, err := NewChannelGroup(systemChannelConf)
	if err != nil {
		return nil, fmt.Errorf("could not parse system channel config %w", err)
	}

	template, err := ConfigTemplateFromGroup(conf, cg)
	if err != nil {
		return nil, fmt.Errorf("could not create config template %w", err)
	}

	return MakeChannelCreationTransactionFromTemplate(channelID, signer, conf, template)
}

// MakeChannelCreationTransactionFromTemplate creates a transaction for creating a channel.  It uses
// the given template to produce the config update set.  Usually, the caller will want to invoke
// MakeChannelCreationTransaction or MakeChannelCreationTransactionWithSystemChannelContext.
func MakeChannelCreationTransactionFromTemplate(
	channelID string,
	signer identity.SignerSerializer,
	conf *genesisconfig.Profile,
	template *cb.ConfigGroup,
) (*cb.Envelope, error) {
	newChannelConfigUpdate, err := NewChannelCreateConfigUpdate(channelID, conf, template)
	if err != nil {
		return nil, fmt.Errorf("config update generation failure %w", err)
	}

	newConfigUpdateEnv := &cb.ConfigUpdateEnvelope{
		ConfigUpdate: protoutil.MarshalOrPanic(newChannelConfigUpdate),
	}

	if signer != nil {
		sigHeader, err := protoutil.NewSignatureHeader(signer)
		if err != nil {
			return nil, fmt.Errorf("creating signature header failed %w", err)
		}

		newConfigUpdateEnv.Signatures = []*cb.ConfigSignature{{
			SignatureHeader: protoutil.MarshalOrPanic(sigHeader),
		}}

		newConfigUpdateEnv.Signatures[0].Signature, err = signer.Sign(util.Concatenate(newConfigUpdateEnv.Signatures[0].SignatureHeader, newConfigUpdateEnv.ConfigUpdate))
		if err != nil {
			return nil, fmt.Errorf("signature failure over config update %w", err)
		}

	}

	return protoutil.CreateSignedEnvelope(cb.HeaderType_CONFIG_UPDATE, channelID, signer, newConfigUpdateEnv, msgVersion, epoch)
}

// HasSkippedForeignOrgs is used to detect whether a configuration includes
// org definitions which should not be parsed because this tool is being
// run in a context where the user does not have access to that org's info
func HasSkippedForeignOrgs(conf *genesisconfig.Profile) error {
	var organizations []*genesisconfig.Organization

	if conf.Orderer != nil {
		organizations = append(organizations, conf.Orderer.Organizations...)
	}

	if conf.Application != nil {
		organizations = append(organizations, conf.Application.Organizations...)
	}

	for _, consortium := range conf.Consortiums {
		organizations = append(organizations, consortium.Organizations...)
	}

	for _, org := range organizations {
		if org.SkipAsForeign {
			return fmt.Errorf("organization '%s' is marked to be skipped as foreign", org.Name)
		}
	}

	return nil
}

// Bootstrapper is a wrapper around NewChannelConfigGroup which can produce genesis blocks
type Bootstrapper struct {
	channelGroup *cb.ConfigGroup
}

// NewBootstrapper creates a bootstrapper but returns an error instead of panic-ing
func NewBootstrapper(config *genesisconfig.Profile) (*Bootstrapper, error) {
	if err := HasSkippedForeignOrgs(config); err != nil {
		return nil, fmt.Errorf("all org definitions must be local during bootstrapping %w", err)
	}

	channelGroup, err := NewChannelGroup(config)
	if err != nil {
		return nil, fmt.Errorf("could not create channel group %w", err)
	}

	return &Bootstrapper{
		channelGroup: channelGroup,
	}, nil
}

// New creates a new Bootstrapper for generating genesis blocks
func New(config *genesisconfig.Profile) *Bootstrapper {
	bs, err := NewBootstrapper(config)
	if err != nil {
		panic(err)
	}
	return bs
}

// GenesisBlockForChannel produces a genesis block for a given channel ID
func (bs *Bootstrapper) GenesisBlockForChannel(channelID string) *cb.Block {
	return genesis.NewFactoryImpl(bs.channelGroup).Block(channelID)
}
