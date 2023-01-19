package chaincode

import (
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/internal/proposal"

	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/protobuf/proto"
)

// Commit a chaincode package to specific peer.
func Commit(Definition Definition, id identity.SigningIdentity, EndorserClients []peer.EndorserClient, BroadcastClient orderer.AtomicBroadcast_BroadcastClient) error {
	proposal, err := CreateCommitProposal(Definition, id)
	if err != nil {
		return err
	}
	return processProposalWithBroadcast(proposal, id, EndorserClients, BroadcastClient)
}

func CreateCommitProposal(Definition Definition, id identity.SigningIdentity) (*peer.Proposal, error) {
	if Definition.ChannelName == "" {
		return nil, fmt.Errorf("For chaincode commit channel id is needed.")
	}
	if Definition.Name == "" {
		return nil, fmt.Errorf("For chaincode commit chaincode name is needed.")
	}
	if Definition.Version == "" {
		return nil, fmt.Errorf("For chaincode commit chaincode version is needed.")
	}
	if Definition.Sequence == 0 {
		return nil, fmt.Errorf("For chaincode commit chaincode Sequence is needed.")
	}
	args := &lifecycle.CommitChaincodeDefinitionArgs{
		Name:                Definition.Name,
		Version:             Definition.Version,
		Sequence:            Definition.Sequence,
		EndorsementPlugin:   Definition.EndorsementPlugin,
		ValidationPlugin:    Definition.ValidationPlugin,
		ValidationParameter: Definition.ValidationParameter,
		InitRequired:        Definition.InitRequired,
		Collections:         Definition.Collections,
	}

	argsBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, err
	}

	return proposal.NewProposal(id, lifecycleChaincodeName, commitTransactionName, proposal.WithChannel(Definition.ChannelName), proposal.WithArguments(argsBytes))
}
