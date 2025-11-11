package channel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hyperledger/fabric-admin-sdk/internal/osnadmin"
	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/internal/proposal"

	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type ChannelList struct {
	SystemChannel interface{} `json:"systemChannel"`
	Channels      []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"channels"`
}

func CreateChannel(osnURL string, block *cb.Block, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (*http.Response, error) {
	blockBytes := protoutil.MarshalOrPanic(block)
	return osnadmin.Join(osnURL, blockBytes, caCertPool, tlsClientCert)
}

func JoinChannel(ctx context.Context, connection grpc.ClientConnInterface, id identity.SigningIdentity, block *cb.Block) error {
	blockBytes, err := proto.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	prop, err := proposal.NewProposal(id, "cscc", "JoinChain", proposal.WithArguments(blockBytes), proposal.WithType(cb.HeaderType_CONFIG))
	if err != nil {
		return err
	}

	signedProp, err := proposal.NewSignedProposal(prop, id)
	if err != nil {
		return err
	}

	endorser := peer.NewEndorserClient(connection)
	proposalResp, err := endorser.ProcessProposal(ctx, signedProp)
	if err != nil {
		return err
	}

	if err := proposal.CheckSuccessfulResponse(proposalResp); err != nil {
		return err
	}

	return nil
}

func ListChannel(osnURL string, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (ChannelList, error) {

	var channels ChannelList
	resp, err := osnadmin.ListAllChannels(osnURL, caCertPool, tlsClientCert)
	if err != nil {
		return channels, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return channels, err
	}

	if err := json.Unmarshal(body, &channels); err != nil {
		return channels, err
	}

	return channels, nil
}

func ListChannelOnPeer(ctx context.Context, connection grpc.ClientConnInterface, id identity.SigningIdentity) ([]*peer.ChannelInfo, error) {
	prop, err := proposal.NewProposal(id, "cscc", "GetChannels", proposal.WithType(cb.HeaderType_ENDORSER_TRANSACTION))
	if err != nil {
		return nil, err
	}

	signedProp, err := proposal.NewSignedProposal(prop, id)
	if err != nil {
		return nil, err
	}

	endorser := peer.NewEndorserClient(connection)

	proposalResp, err := endorser.ProcessProposal(ctx, signedProp)
	if err != nil {
		return nil, err
	}

	if err := proposal.CheckSuccessfulResponse(proposalResp); err != nil {
		return nil, err
	}

	var channelQueryResponse peer.ChannelQueryResponse
	err = proto.Unmarshal(proposalResp.GetResponse().GetPayload(), &channelQueryResponse)
	if err != nil {
		return nil, err
	}
	return channelQueryResponse.GetChannels(), nil
}

func UpdateChannel(ctx context.Context, osnURL string, caCertPool *x509.CertPool, tlsClientCert tls.Certificate, channelName string, env *cb.Envelope) (*http.Response, error) {
	envelopeBytes := protoutil.MarshalOrPanic(env)
	return osnadmin.Update(ctx, osnURL, channelName, caCertPool, tlsClientCert, envelopeBytes)
}
