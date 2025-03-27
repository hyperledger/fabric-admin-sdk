package test

import (
	"errors"
	"os"
	"testing"

	"github.com/hyperledger/fabric-admin-sdk/internal/configtxgen/encoder"
	"github.com/hyperledger/fabric-admin-sdk/internal/configtxgen/genesisconfig"
	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-admin-sdk/internal/util"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	msgVersion = int32(0)
	epoch      = 0
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
	_ = os.RemoveAll(tmpDir)
})

// ConfigTxGen based on Profile return block
func ConfigTxGen(config *genesisconfig.Profile, channelID string) (*cb.Block, error) {
	pgen, err := encoder.NewBootstrapper(config)
	if err != nil {
		return nil, err
	}
	genesisBlock := pgen.GenesisBlockForChannel(channelID)
	return genesisBlock, nil
}

func SignConfigTx(channelID string, envConfigUpdate *cb.Envelope, signer identity.SigningIdentity) (*cb.Envelope, error) {
	payload, err := protoutil.UnmarshalPayload(envConfigUpdate.Payload)
	if err != nil {
		return nil, errors.New("bad payload")
	}

	if payload.Header == nil || payload.Header.ChannelHeader == nil {
		return nil, errors.New("bad header")
	}

	ch, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, errors.New("could not unmarshall channel header")
	}

	if ch.Type != int32(cb.HeaderType_CONFIG_UPDATE) {
		return nil, errors.New("bad type")
	}

	if ch.ChannelId == "" {
		return nil, errors.New("empty channel id")
	}

	configUpdateEnv, err := protoutil.UnmarshalConfigUpdateEnvelope(payload.Data)
	if err != nil {
		return nil, errors.New("bad config update env")
	}

	sigHeader, err := protoutil.NewSignatureHeader(signer)
	if err != nil {
		return nil, err
	}

	configSig := &cb.ConfigSignature{
		SignatureHeader: protoutil.MarshalOrPanic(sigHeader),
	}

	configSig.Signature, err = signer.Sign(util.Concatenate(configSig.SignatureHeader, configUpdateEnv.ConfigUpdate))
	if err != nil {
		return nil, err
	}

	configUpdateEnv.Signatures = append(configUpdateEnv.Signatures, configSig)

	return protoutil.CreateSignedEnvelope(cb.HeaderType_CONFIG_UPDATE, channelID, signer, configUpdateEnv, msgVersion, epoch)
}
