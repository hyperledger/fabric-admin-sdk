package util

import (
	"errors"

	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
)

func Concatenate[T any](slices ...[]T) []T {
	size := 0
	for _, slice := range slices {
		size += len(slice)
	}

	result := make([]T, size)
	i := 0
	for _, slice := range slices {
		copy(result[i:], slice)
		i += len(slice)
	}

	return result
}

const (
	msgVersion = int32(0)
	epoch      = 0
)

func SignConfigTx(channelID string, envConfigUpdate *cb.Envelope, signer identity.SigningIdentity) (*cb.Envelope, error) {
	payload, err := protoutil.UnmarshalPayload(envConfigUpdate.GetPayload())
	if err != nil {
		return nil, errors.New("bad payload")
	}

	if payload.GetHeader() == nil || payload.Header.ChannelHeader == nil {
		return nil, errors.New("bad header")
	}

	ch, err := protoutil.UnmarshalChannelHeader(payload.GetHeader().GetChannelHeader())
	if err != nil {
		return nil, errors.New("could not unmarshall channel header")
	}

	if ch.GetType() != int32(cb.HeaderType_CONFIG_UPDATE) {
		return nil, errors.New("bad type")
	}

	if ch.GetChannelId() == "" {
		return nil, errors.New("empty channel id")
	}

	configUpdateEnv, err := protoutil.UnmarshalConfigUpdateEnvelope(payload.GetData())
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

	configSig.Signature, err = signer.Sign(Concatenate(configSig.GetSignatureHeader(), configUpdateEnv.GetConfigUpdate()))
	if err != nil {
		return nil, err
	}

	configUpdateEnv.Signatures = append(configUpdateEnv.Signatures, configSig)

	return protoutil.CreateSignedEnvelope(cb.HeaderType_CONFIG_UPDATE, channelID, signer, configUpdateEnv, msgVersion, epoch)
}
