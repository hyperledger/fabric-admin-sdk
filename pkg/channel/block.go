package channel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/internal/osnadmin"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/internal/proposal"

	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	pb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// GetConfigBlock get block config
func GetConfigBlock(ctx context.Context, connection grpc.ClientConnInterface, id identity.SigningIdentity, channelID string) (*cb.Block, error) {
	proposalResp, err := getSignedProposal(ctx, connection, channelID, "cscc", "GetConfigBlock", id)
	if err != nil {
		return nil, fmt.Errorf("get signed proposal %w", err)
	}

	block := &cb.Block{}
	if err := proto.Unmarshal(proposalResp.GetResponse().GetPayload(), block); err != nil {
		return nil, fmt.Errorf("block unmarshal %w", err)
	}
	return block, nil
}

// GetBlockChainInfo get chain info
func GetBlockChainInfo(ctx context.Context, connection grpc.ClientConnInterface, id identity.SigningIdentity, channelID string) (*cb.BlockchainInfo, error) {
	proposalResp, err := getSignedProposal(ctx, connection, channelID, "qscc", "GetChainInfo", id)
	if err != nil {
		return nil, fmt.Errorf("get signed proposal %w", err)
	}

	blockChainInfo := &cb.BlockchainInfo{}
	err = proto.Unmarshal(proposalResp.GetResponse().GetPayload(), blockChainInfo)
	if err != nil {
		return nil, fmt.Errorf("block unmarshal %w", err)
	}
	return blockChainInfo, nil
}

func getSignedProposal(ctx context.Context, connection grpc.ClientConnInterface, channelID, ccName, funcName string, id identity.SigningIdentity) (*pb.ProposalResponse, error) {
	prop, err := proposal.NewProposal(id, ccName, funcName, proposal.WithChannel(channelID), proposal.WithArguments([]byte(channelID)))
	if err != nil {
		return nil, err
	}

	signedProp, err := proposal.NewSignedProposal(prop, id)
	if err != nil {
		return nil, err
	}

	endorser := pb.NewEndorserClient(connection)
	proposalResp, err := endorser.ProcessProposal(ctx, signedProp)
	if err != nil {
		return nil, fmt.Errorf("process proposal %w", err)
	}

	return proposalResp, nil
}

func GetBlock(ctx context.Context, osnURL, channelID, blockID string, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (*cb.Block, error) {
	return osnadmin.Fetch(ctx, osnURL, channelID, blockID, caCertPool, tlsClientCert)
}
