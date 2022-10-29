package chaincode

import (
	"context"
	"fabric-admin-sdk/internal/pkg/identity"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

const approveFuncName = "ApproveChaincodeDefinitionForMyOrg"
const lifecycleName = "_lifecycle"

func Approve(Signer identity.CryptoImpl, ChannelID, inputTxID, PackageID, Name, Version, EndorsementPlugin, ValidationPlugin string,
	Sequence int64, ValidationParameterBytes []byte, InitRequired bool, CollectionConfigPackage *peer.CollectionConfigPackage,
	EndorserClients []pb.EndorserClient, BroadcastClient ab.AtomicBroadcast_BroadcastClient) error {
	proposal, _, err := createProposal(Signer, ChannelID, inputTxID, PackageID, Name, Version, EndorsementPlugin, ValidationPlugin,
		Sequence, ValidationParameterBytes, InitRequired, CollectionConfigPackage)
	if err != nil {
		return err
	}
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

func createProposal(Signer identity.CryptoImpl, ChannelID, inputTxID, PackageID, Name, Version, EndorsementPlugin, ValidationPlugin string,
	Sequence int64, ValidationParameterBytes []byte, InitRequired bool, CollectionConfigPackage *peer.CollectionConfigPackage) (proposal *pb.Proposal, txID string, err error) {
	var ccsrc *lb.ChaincodeSource
	if PackageID != "" {
		ccsrc = &lb.ChaincodeSource{
			Type: &lb.ChaincodeSource_LocalPackage{
				LocalPackage: &lb.ChaincodeSource_Local{
					PackageId: PackageID,
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
		Name:                Name,
		Version:             Version,
		Sequence:            Sequence,
		EndorsementPlugin:   EndorsementPlugin,
		ValidationPlugin:    ValidationPlugin,
		ValidationParameter: ValidationParameterBytes,
		InitRequired:        InitRequired,
		Collections:         CollectionConfigPackage,
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

	proposal, txID, err = protoutil.CreateChaincodeProposalWithTxIDAndTransient(cb.HeaderType_ENDORSER_TRANSACTION, ChannelID, cis, creatorBytes, inputTxID, nil)
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
