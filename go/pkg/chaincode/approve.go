package chaincode

import (
	"context"
	"encoding/json"
	"errors"
	"fabric-admin-sdk/internal/pkg/identity"
	"fabric-admin-sdk/internal/protoutil"
	"fmt"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/protobuf/proto"
)

const approveFuncName = "ApproveChaincodeDefinitionForMyOrg"
const lifecycleName = "_lifecycle"
const checkCommitReadinessFuncName = "CheckCommitReadiness"

func Approve(CCDefine CCDefine, Signer identity.CryptoImpl, EndorserClients []peer.EndorserClient, BroadcastClient orderer.AtomicBroadcast_BroadcastClient) error {
	proposal, _, err := createProposal(CCDefine, Signer)
	if err != nil {
		return err
	}
	return processProposalWithBroadcast(proposal, Signer, EndorserClients, BroadcastClient)
}

func createProposal(CCDefine CCDefine, Signer identity.CryptoImpl) (proposal *peer.Proposal, txID string, err error) {
	var ccsrc *lifecycle.ChaincodeSource
	if CCDefine.PackageID != "" {
		ccsrc = &lifecycle.ChaincodeSource{
			Type: &lifecycle.ChaincodeSource_LocalPackage{
				LocalPackage: &lifecycle.ChaincodeSource_Local{
					PackageId: CCDefine.PackageID,
				},
			},
		}
	} else {
		ccsrc = &lifecycle.ChaincodeSource{
			Type: &lifecycle.ChaincodeSource_Unavailable_{
				Unavailable: &lifecycle.ChaincodeSource_Unavailable{},
			},
		}
	}

	args := &lifecycle.ApproveChaincodeDefinitionForMyOrgArgs{
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
	ccInput := &peer.ChaincodeInput{Args: [][]byte{[]byte(approveFuncName), argsBytes}}

	cis := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			ChaincodeId: &peer.ChaincodeID{Name: lifecycleName},
			Input:       ccInput,
		},
	}

	creatorBytes, err := Signer.Serialize()
	if err != nil {
		return nil, "", fmt.Errorf("failed to serialize identity %w", err)
	}

	proposal, txID, err = protoutil.CreateChaincodeProposalWithTxIDAndTransient(common.HeaderType_ENDORSER_TRANSACTION, CCDefine.ChannelID, cis, creatorBytes, CCDefine.InputTxID, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create ChaincodeInvocationSpec proposal %w", err)
	}

	return proposal, txID, nil
}

func signProposal(proposal *peer.Proposal, signer identity.CryptoImpl) (*peer.SignedProposal, error) {
	// check for nil argument
	if proposal == nil {
		return nil, errors.New("proposal cannot be nil")
	}

	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, fmt.Errorf("error marshaling proposal %w", err)
	}

	signature, err := signer.Sign(proposalBytes)
	if err != nil {
		return nil, err
	}

	return &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}, nil
}

func processProposalWithBroadcast(proposal *peer.Proposal, Signer identity.CryptoImpl, EndorserClients []peer.EndorserClient, BroadcastClient orderer.AtomicBroadcast_BroadcastClient) error {
	//sign
	signedProposal, err := signProposal(proposal, Signer)
	if err != nil {
		return err
	}
	//ProcessProposal
	var responses []*peer.ProposalResponse
	for _, endorser := range EndorserClients {
		proposalResponse, err := endorser.ProcessProposal(context.Background(), signedProposal)
		if err != nil {
			return fmt.Errorf("failed to endorse proposal %w", err)
		}
		responses = append(responses, proposalResponse)
	}
	//CreateSignedTx
	env, err := protoutil.CreateSignedTx(proposal, Signer, responses...)
	if err != nil {
		return fmt.Errorf("failed to create signed transaction %w", err)
	}
	//Send Broadcast
	if err = BroadcastClient.Send(env); err != nil {
		return fmt.Errorf("failed to send transaction %w", err)
	}
	return nil
}

func ReadinessCheck(CCDefine CCDefine, Signer identity.CryptoImpl, EndorserClient peer.EndorserClient) error {
	proposal, err := createReadinessCheckProposal(CCDefine, Signer)
	if err != nil {
		return err
	}
	return processProposal(proposal, Signer, EndorserClient)
}

func processProposal(proposal *peer.Proposal, Signer identity.CryptoImpl, EndorserClient peer.EndorserClient) error {
	signedProposal, err := signProposal(proposal, Signer)
	if err != nil {
		return fmt.Errorf("failed to create signed proposal %w", err)
	}

	// checkcommitreadiness currently only supports a single peer
	proposalResponse, err := EndorserClient.ProcessProposal(context.Background(), signedProposal)
	if err != nil {
		return fmt.Errorf("failed to endorse proposal %w", err)
	}

	if proposalResponse == nil {
		return errors.New("received nil proposal response")
	}

	if proposalResponse.Response == nil {
		return errors.New("received proposal response with nil response")
	}

	if proposalResponse.Response.Status != int32(common.Status_SUCCESS) {
		return fmt.Errorf("query failed with status: %d - %s", proposalResponse.Response.Status, proposalResponse.Response.Message)
	}

	return printResponseAsJSON(proposalResponse, &lifecycle.CheckCommitReadinessResult{})
}

func printResponseAsJSON(proposalResponse *peer.ProposalResponse, msg proto.Message) error {
	err := proto.Unmarshal(proposalResponse.Response.Payload, msg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal proposal response's response payload as type %T %w", msg, err)
	}

	bytes, err := json.MarshalIndent(msg, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal output %w", err)
	}

	fmt.Println(string(bytes))
	return nil
}

func createReadinessCheckProposal(CCDefine CCDefine, Signer identity.CryptoImpl) (*peer.Proposal, error) {
	args := &lifecycle.CheckCommitReadinessArgs{
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
	ccInput := &peer.ChaincodeInput{Args: [][]byte{[]byte(checkCommitReadinessFuncName), argsBytes}}

	cis := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			ChaincodeId: &peer.ChaincodeID{Name: lifecycleName},
			Input:       ccInput,
		},
	}

	creatorBytes, err := Signer.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize identity %w", err)
	}

	proposal, _, err := protoutil.CreateChaincodeProposalWithTxIDAndTransient(common.HeaderType_ENDORSER_TRANSACTION, CCDefine.ChannelID, cis, creatorBytes, CCDefine.InputTxID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaincodeInvocationSpec proposal %w", err)
	}

	return proposal, nil
}
