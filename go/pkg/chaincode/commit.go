package chaincode

import (
	"fabric-admin-sdk/internal/pkg/identity"
	"fabric-admin-sdk/internal/protoutil"
	"fmt"

	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	ab "github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	pb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
	lb "github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/protobuf/proto"
)

const commitFuncName = "CommitChaincodeDefinition"

type CCDefine struct {
	ChannelID                string
	InputTxID                string
	PackageID                string
	Name                     string
	Version                  string
	EndorsementPlugin        string
	ValidationPlugin         string
	Sequence                 int64
	ValidationParameterBytes []byte
	InitRequired             bool
	CollectionConfigPackage  *peer.CollectionConfigPackage
}

func Commit(CCDefine CCDefine, Signer identity.CryptoImpl, EndorserClients []pb.EndorserClient, BroadcastClient ab.AtomicBroadcast_BroadcastClient) error {
	proposal, _, err := createCommitProposal(CCDefine, Signer)
	if err != nil {
		return err
	}
	return processProposalWithBroadcast(proposal, Signer, EndorserClients, BroadcastClient)
}

func createCommitProposal(CCDefine CCDefine, Signer identity.CryptoImpl) (proposal *pb.Proposal, txID string, err error) {
	args := &lb.CommitChaincodeDefinitionArgs{
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
		return nil, "", fmt.Errorf("failed to serialize identity %w", err)
	}

	proposal, txID, err = protoutil.CreateChaincodeProposalWithTxIDAndTransient(cb.HeaderType_ENDORSER_TRANSACTION, CCDefine.ChannelID, cis, creatorBytes, CCDefine.InputTxID, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create ChaincodeInvocationSpec proposal %w", err)
	}

	return proposal, txID, nil
}
