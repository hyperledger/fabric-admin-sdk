package chaincode

import (
	"context"
	"encoding/json"
	"fabric-admin-sdk/internal/pkg/identity"
	"fmt"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

const approveFuncName = "ApproveChaincodeDefinitionForMyOrg"
const lifecycleName = "_lifecycle"
const checkCommitReadinessFuncName = "CheckCommitReadiness"

func Approve(CCDefine CCDefine, Signer identity.CryptoImpl, EndorserClients []pb.EndorserClient, BroadcastClient ab.AtomicBroadcast_BroadcastClient) error {
	proposal, _, err := createProposal(CCDefine, Signer)
	if err != nil {
		return err
	}
	return processProposalWithBroadcast(proposal, Signer, EndorserClients, BroadcastClient)
}

func createProposal(CCDefine CCDefine, Signer identity.CryptoImpl) (proposal *pb.Proposal, txID string, err error) {
	var ccsrc *lb.ChaincodeSource
	if CCDefine.PackageID != "" {
		ccsrc = &lb.ChaincodeSource{
			Type: &lb.ChaincodeSource_LocalPackage{
				LocalPackage: &lb.ChaincodeSource_Local{
					PackageId: CCDefine.PackageID,
				},
			},
		}
	} else {
		ccsrc = &lb.ChaincodeSource{
			Type: &lb.ChaincodeSource_Unavailable_{
				Unavailable: &lb.ChaincodeSource_Unavailable{},
			},
		}
	}

	args := &lb.ApproveChaincodeDefinitionForMyOrgArgs{
		Name:                CCDefine.Name,
		Version:             CCDefine.Version,
		Sequence:            CCDefine.Sequence,
		EndorsementPlugin:   CCDefine.EndorsementPlugin,
		ValidationPlugin:    CCDefine.ValidationPlugin,
		ValidationParameter: CCDefine.ValidationParameterBytes,
		InitRequired:        CCDefine.InitRequired,
		Collections:         CCDefine.CollectionConfigPackage,
		Source:              ccsrc,
	}

	argsBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, "", err
	}
	ccInput := &pb.ChaincodeInput{Args: [][]byte{[]byte(approveFuncName), argsBytes}}

	cis := &pb.ChaincodeInvocationSpec{
		ChaincodeSpec: &pb.ChaincodeSpec{
			ChaincodeId: &pb.ChaincodeID{Name: lifecycleName},
			Input:       ccInput,
		},
	}

	creatorBytes, err := Signer.Serialize()
	if err != nil {
		return nil, "", errors.WithMessage(err, "failed to serialize identity")
	}

	proposal, txID, err = protoutil.CreateChaincodeProposalWithTxIDAndTransient(cb.HeaderType_ENDORSER_TRANSACTION, CCDefine.ChannelID, cis, creatorBytes, CCDefine.InputTxID, nil)
	if err != nil {
		return nil, "", errors.WithMessage(err, "failed to create ChaincodeInvocationSpec proposal")
	}

	return proposal, txID, nil
}

func signProposal(proposal *pb.Proposal, signer identity.CryptoImpl) (*pb.SignedProposal, error) {
	// check for nil argument
	if proposal == nil {
		return nil, errors.New("proposal cannot be nil")
	}

	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling proposal")
	}

	signature, err := signer.Sign(proposalBytes)
	if err != nil {
		return nil, err
	}

	return &pb.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}, nil
}

func processProposalWithBroadcast(proposal *pb.Proposal, Signer identity.CryptoImpl, EndorserClients []pb.EndorserClient, BroadcastClient ab.AtomicBroadcast_BroadcastClient) error {
	//sign
	signedProposal, err := signProposal(proposal, Signer)
	if err != nil {
		return err
	}
	//ProcessProposal
	var responses []*pb.ProposalResponse
	for _, endorser := range EndorserClients {
		proposalResponse, err := endorser.ProcessProposal(context.Background(), signedProposal)
		if err != nil {
			return errors.WithMessage(err, "failed to endorse proposal")
		}
		responses = append(responses, proposalResponse)
	}
	//CreateSignedTx
	env, err := protoutil.CreateSignedTx(proposal, Signer, responses...)
	if err != nil {
		return errors.WithMessage(err, "failed to create signed transaction")
	}
	//Send Broadcast
	if err = BroadcastClient.Send(env); err != nil {
		return errors.WithMessage(err, "failed to send transaction")
	}
	return nil
}

func ReadinessCheck(CCDefine CCDefine, Signer identity.CryptoImpl, EndorserClient pb.EndorserClient) error {
	proposal, err := createReadinessCheckProposal(CCDefine, Signer)
	if err != nil {
		return err
	}
	return processProposal(proposal, Signer, EndorserClient)
}

func processProposal(proposal *pb.Proposal, Signer identity.CryptoImpl, EndorserClient pb.EndorserClient) error {
	signedProposal, err := signProposal(proposal, Signer)
	if err != nil {
		return errors.WithMessage(err, "failed to create signed proposal")
	}

	// checkcommitreadiness currently only supports a single peer
	proposalResponse, err := EndorserClient.ProcessProposal(context.Background(), signedProposal)
	if err != nil {
		return errors.WithMessage(err, "failed to endorse proposal")
	}

	if proposalResponse == nil {
		return errors.New("received nil proposal response")
	}

	if proposalResponse.Response == nil {
		return errors.New("received proposal response with nil response")
	}

	if proposalResponse.Response.Status != int32(cb.Status_SUCCESS) {
		return errors.Errorf("query failed with status: %d - %s", proposalResponse.Response.Status, proposalResponse.Response.Message)
	}

	return printResponseAsJSON(proposalResponse, &lb.CheckCommitReadinessResult{})
}

func printResponseAsJSON(proposalResponse *pb.ProposalResponse, msg proto.Message) error {
	err := proto.Unmarshal(proposalResponse.Response.Payload, msg)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal proposal response's response payload as type %T", msg)
	}

	bytes, err := json.MarshalIndent(msg, "", "\t")
	if err != nil {
		return errors.Wrap(err, "failed to marshal output")
	}

	fmt.Println(string(bytes))
	return nil
}

func createReadinessCheckProposal(CCDefine CCDefine, Signer identity.CryptoImpl) (*pb.Proposal, error) {
	args := &lb.CheckCommitReadinessArgs{
		Name:                CCDefine.Name,
		Version:             CCDefine.Version,
		Sequence:            CCDefine.Sequence,
		EndorsementPlugin:   CCDefine.EndorsementPlugin,
		ValidationPlugin:    CCDefine.ValidationPlugin,
		ValidationParameter: CCDefine.ValidationParameterBytes,
		InitRequired:        CCDefine.InitRequired,
		Collections:         CCDefine.CollectionConfigPackage,
	}

	argsBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, err
	}
	ccInput := &pb.ChaincodeInput{Args: [][]byte{[]byte(checkCommitReadinessFuncName), argsBytes}}

	cis := &pb.ChaincodeInvocationSpec{
		ChaincodeSpec: &pb.ChaincodeSpec{
			ChaincodeId: &pb.ChaincodeID{Name: lifecycleName},
			Input:       ccInput,
		},
	}

	creatorBytes, err := Signer.Serialize()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to serialize identity")
	}

	proposal, _, err := protoutil.CreateChaincodeProposalWithTxIDAndTransient(cb.HeaderType_ENDORSER_TRANSACTION, CCDefine.ChannelID, cis, creatorBytes, CCDefine.InputTxID, nil)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create ChaincodeInvocationSpec proposal")
	}

	return proposal, nil
}
