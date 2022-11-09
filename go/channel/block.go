package channel

import (
	"context"
	"fabric-admin-sdk/internal/pkg/identity"
	"fabric-admin-sdk/tools"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

// GetConfigBlock get block config
func GetConfigBlock(signCert, priKey, MSPID, channelID string, connection pb.EndorserClient) (*cb.Block, error) {

	signer, err := tools.CreateSigner(priKey, signCert, MSPID)
	if err != nil {
		return nil, errors.WithMessage(err, "create signer")
	}
	proposalResp, err := getSignedProposal(channelID, "cscc", "GetConfigBlock", signer, connection)
	if err != nil {
		return nil, errors.WithMessage(err, "get signed proposal")
	}

	block := &cb.Block{}
	if err := proto.Unmarshal(proposalResp.Response.Payload, block); err != nil {
		return nil, errors.WithMessage(err, "block unmarshal")
	}
	return block, nil
}

// GetBlockChainInfo get chain info
func GetBlockChainInfo(signCert, priKey, MSPID, channelID string, connection pb.EndorserClient) (*cb.BlockchainInfo, error) {

	signer, err := tools.CreateSigner(priKey, signCert, MSPID)
	if err != nil {
		return nil, errors.WithMessage(err, "get signer")
	}
	proposalResp, err := getSignedProposal(channelID, "qscc", "GetChainInfo", signer, connection)
	if err != nil {
		return nil, errors.WithMessage(err, "get signed proposal")
	}

	blockChainInfo := &cb.BlockchainInfo{}
	err = proto.Unmarshal(proposalResp.Response.Payload, blockChainInfo)
	if err != nil {
		return nil, errors.WithMessage(err, "block unmarshal")
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
		return nil, errors.WithMessage(err, "signer serialize")
	}
	prop, _, err := protoutil.CreateProposalFromCIS(cb.HeaderType_ENDORSER_TRANSACTION, channelID, spec, c)
	if err != nil {
		return nil, errors.WithMessage(err, "create proposal")
	}

	var signedProp *pb.SignedProposal
	signedProp, err = protoutil.GetSignedProposal(prop, signer)
	if err != nil {
		return nil, errors.WithMessage(err, "get signe proposal")
	}

	var proposalResp *pb.ProposalResponse
	proposalResp, err = connection.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return nil, errors.WithMessage(err, "process proposal")
	}

	return proposalResp, nil
}
