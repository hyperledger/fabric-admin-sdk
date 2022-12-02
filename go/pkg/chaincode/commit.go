package chaincode

import (
	"fabric-admin-sdk/internal/pkg/identity"
	"fabric-admin-sdk/pkg/internal/proposal"

	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
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

func Commit(CCDefine CCDefine, signer identity.CryptoImpl, EndorserClients []peer.EndorserClient, BroadcastClient orderer.AtomicBroadcast_BroadcastClient) error {
	proposal, err := createCommitProposal(CCDefine, signer)
	if err != nil {
		return err
	}
	return processProposalWithBroadcast(proposal, signer, EndorserClients, BroadcastClient)
}

func createCommitProposal(CCDefine CCDefine, signer identity.CryptoImpl) (*peer.Proposal, error) {
	args := &lifecycle.CommitChaincodeDefinitionArgs{
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

	return proposal.NewProposal(signer, lifecycleName, commitFuncName, proposal.WithChannel(CCDefine.ChannelID), proposal.WithArguments(argsBytes))
}
