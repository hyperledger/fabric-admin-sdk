package channel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fabric-admin-sdk/internal/osnadmin"
	"fabric-admin-sdk/internal/protoutil"
	"fabric-admin-sdk/pkg/identity"
	"fabric-admin-sdk/pkg/internal/proposal"
	"fmt"
	"io"
	"net/http"

	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	pb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
)

type ChannelList struct {
	SystemChannel interface{} `json:"systemChannel"`
	Channels      []struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"channels"`
}

func CreateChannel(osnURL string, block *cb.Block, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (*http.Response, error) {
	block_byte := protoutil.MarshalOrPanic(block)
	return osnadmin.Join(osnURL, block_byte, caCertPool, tlsClientCert)
}

func JoinChannel(block *cb.Block, id identity.SigningIdentity, connection pb.EndorserClient) error {
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

	proposalResp, err := connection.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return err
	}

	if err = proposal.CheckSuccessfulResponse(proposalResp); err != nil {
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
