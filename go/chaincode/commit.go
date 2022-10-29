package chaincode

import (
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

const commitFuncName = "CommitChaincodeDefinition"

func Commit(ChannelID, inputTxID, PackageID, Name, Version, EndorsementPlugin, ValidationPlugin string,
	Sequence int64, ValidationParameterBytes []byte, InitRequired bool, CollectionConfigPackage *peer.CollectionConfigPackage,
	Signer identity.CryptoImpl, EndorserClients []pb.EndorserClient, BroadcastClient ab.AtomicBroadcast_BroadcastClient) error {
	proposal, _, err := createCommitProposal(ChannelID, inputTxID, PackageID, Name, Version, EndorsementPlugin, ValidationPlugin,
		Sequence, ValidationParameterBytes, InitRequired, CollectionConfigPackage, Signer)
	if err != nil {
		return err
	}
	return processProposal(proposal, Signer, EndorserClients, BroadcastClient)
}

func createCommitProposal(ChannelID, inputTxID, PackageID, Name, Version, EndorsementPlugin, ValidationPlugin string,
	Sequence int64, ValidationParameterBytes []byte, InitRequired bool, CollectionConfigPackage *peer.CollectionConfigPackage,
	Signer identity.CryptoImpl) (proposal *pb.Proposal, txID string, err error) {
	args := &lb.CommitChaincodeDefinitionArgs{
		Name:                Name,
		Version:             Version,
		Sequence:            Sequence,
		EndorsementPlugin:   EndorsementPlugin,
		ValidationPlugin:    ValidationPlugin,
		ValidationParameter: ValidationParameterBytes,
		InitRequired:        InitRequired,
		Collections:         CollectionConfigPackage,
	}

	argsBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, "", err
	}
	ccInput := &pb.ChaincodeInput{Args: [][]byte{[]byte(commitFuncName), argsBytes}}

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
