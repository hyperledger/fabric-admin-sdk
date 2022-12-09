package channel

import (
	"context"
	"fabric-admin-sdk/pkg/identity"
	"fabric-admin-sdk/pkg/internal/proposal"
	"fmt"

	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	pb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
)

// GetConfigBlock get block config
func GetConfigBlock(id identity.SigningIdentity, channelID string, connection pb.EndorserClient) (*cb.Block, error) {
	proposalResp, err := getSignedProposal(channelID, "cscc", "GetConfigBlock", id, connection)
	if err != nil {
		return nil, fmt.Errorf("get signed proposal %w", err)
	}

	block := &cb.Block{}
	if err := proto.Unmarshal(proposalResp.Response.Payload, block); err != nil {
		return nil, fmt.Errorf("block unmarshal %w", err)
	}
	return block, nil
}

// GetBlockChainInfo get chain info
func GetBlockChainInfo(id identity.SigningIdentity, channelID string, connection pb.EndorserClient) (*cb.BlockchainInfo, error) {
	proposalResp, err := getSignedProposal(channelID, "qscc", "GetChainInfo", id, connection)
	if err != nil {
		return nil, fmt.Errorf("get signed proposal %w", err)
	}

	blockChainInfo := &cb.BlockchainInfo{}
	err = proto.Unmarshal(proposalResp.Response.Payload, blockChainInfo)
	if err != nil {
		return nil, fmt.Errorf("block unmarshal %w", err)
	}
	return blockChainInfo, nil
}

func getSignedProposal(channelID, ccName, funcName string, id identity.SigningIdentity, connection pb.EndorserClient) (*pb.ProposalResponse, error) {
	prop, err := proposal.NewProposal(id, ccName, funcName, proposal.WithChannel(channelID), proposal.WithArguments([]byte(channelID)))
	if err != nil {
		return nil, err
	}

	signedProp, err := proposal.NewSignedProposal(prop, id)
	if err != nil {
		return nil, err
	}

	var proposalResp *pb.ProposalResponse
	proposalResp, err = connection.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return nil, fmt.Errorf("process proposal %w", err)
	}

	return proposalResp, nil
}
