package channel

import (
	"context"
	"fabric-admin-sdk/internal/pkg/identity"
	"fabric-admin-sdk/internal/protoutil"
	"fabric-admin-sdk/pkg/tools"
	"fmt"

	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	pb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
)

// GetConfigBlock get block config
func GetConfigBlock(signCert, priKey, MSPID, channelID string, connection pb.EndorserClient) (*cb.Block, error) {

	signer, err := tools.CreateSigner(priKey, signCert, MSPID)
	if err != nil {
		return nil, fmt.Errorf("create signer %w", err)
	}
	proposalResp, err := getSignedProposal(channelID, "cscc", "GetConfigBlock", signer, connection)
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
func GetBlockChainInfo(signCert, priKey, MSPID, channelID string, connection pb.EndorserClient) (*cb.BlockchainInfo, error) {

	signer, err := tools.CreateSigner(priKey, signCert, MSPID)
	if err != nil {
		return nil, fmt.Errorf("get signer %w", err)
	}
	proposalResp, err := getSignedProposal(channelID, "qscc", "GetChainInfo", signer, connection)
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

func getSignedProposal(channelID, ccName, funcName string, signer *identity.CryptoImpl, connection pb.EndorserClient) (*pb.ProposalResponse, error) {

	spec := &pb.ChaincodeInvocationSpec{
		ChaincodeSpec: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
			ChaincodeId: &pb.ChaincodeID{Name: ccName},
			Input:       &pb.ChaincodeInput{Args: [][]byte{[]byte(funcName), []byte(channelID)}},
		},
	}

	c, err := signer.Serialize()
	if err != nil {
		return nil, fmt.Errorf("signer serialize %w", err)
	}
	prop, _, err := protoutil.CreateProposalFromCIS(cb.HeaderType_ENDORSER_TRANSACTION, channelID, spec, c)
	if err != nil {
		return nil, fmt.Errorf("create proposal %w", err)
	}

	var signedProp *pb.SignedProposal
	signedProp, err = protoutil.GetSignedProposal(prop, signer)
	if err != nil {
		return nil, fmt.Errorf("get signe proposal %w", err)
	}

	var proposalResp *pb.ProposalResponse
	proposalResp, err = connection.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return nil, fmt.Errorf("process proposal %w", err)
	}

	return proposalResp, nil
}
