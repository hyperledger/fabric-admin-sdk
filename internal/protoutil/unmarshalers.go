package protoutil

import (
	"fmt"

	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	"google.golang.org/protobuf/proto"
)

// UnmarshalEnvelope unmarshals bytes to a Envelope
func UnmarshalEnvelope(encoded []byte) (*cb.Envelope, error) {
	envelope := &cb.Envelope{}
	if err := proto.Unmarshal(encoded, envelope); err != nil {
		return nil, fmt.Errorf("error unmarshaling Envelope: %w", err)
	}
	return envelope, nil
}

// UnmarshalPayload unmarshals bytes to a Payload
func UnmarshalPayload(encoded []byte) (*cb.Payload, error) {
	payload := &cb.Payload{}
	if err := proto.Unmarshal(encoded, payload); err != nil {
		return nil, fmt.Errorf("error unmarshaling Payload: %w", err)
	}
	return payload, nil
}

// UnmarshalChannelHeader unmarshals bytes to a ChannelHeader
func UnmarshalChannelHeader(bytes []byte) (*cb.ChannelHeader, error) {
	chdr := &cb.ChannelHeader{}
	if err := proto.Unmarshal(bytes, chdr); err != nil {
		return nil, fmt.Errorf("error unmarshaling ChannelHeader: %w", err)
	}
	return chdr, nil
}

// UnmarshalConfigUpdateEnvelope attempts to unmarshal bytes to a *cb.ConfigUpdate
func UnmarshalConfigUpdateEnvelope(data []byte) (*cb.ConfigUpdateEnvelope, error) {
	configUpdateEnvelope := &cb.ConfigUpdateEnvelope{}
	if err := proto.Unmarshal(data, configUpdateEnvelope); err != nil {
		return nil, fmt.Errorf("error unmarshaling ConfigUpdateEnvelope: %w", err)
	}
	return configUpdateEnvelope, nil
}
