package chaincode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/internal/proposal"
	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer/lifecycle"
	"google.golang.org/protobuf/proto"
)

func processProposalWithBroadcast(proposalProto *peer.Proposal, id identity.SigningIdentity, EndorserClients []peer.EndorserClient, BroadcastClient orderer.AtomicBroadcast_BroadcastClient) error {
	signedProposal, err := proposal.NewSignedProposal(proposalProto, id)
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
	env, err := protoutil.CreateSignedTx(proposalProto, id, responses...)
	if err != nil {
		return fmt.Errorf("failed to create signed transaction %w", err)
	}
	//Send Broadcast
	if err = BroadcastClient.Send(env); err != nil {
		return fmt.Errorf("failed to send transaction %w", err)
	}
	return nil
}

func ReadinessCheck(Definition Definition, id identity.SigningIdentity, EndorserClient peer.EndorserClient) error {
	proposal, err := createReadinessCheckProposal(Definition, id)
	if err != nil {
		return err
	}
	return processProposal(proposal, id, EndorserClient)
}

func processProposal(proposalProto *peer.Proposal, signer identity.Signer, EndorserClient peer.EndorserClient) error {
	signedProposal, err := proposal.NewSignedProposal(proposalProto, signer)
	if err != nil {
		return err
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

func createReadinessCheckProposal(Definition Definition, id identity.Identity) (*peer.Proposal, error) {
	args := &lifecycle.CheckCommitReadinessArgs{
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

	return proposal.NewProposal(id, lifecycleChaincodeName, checkCommitReadinessTransactionName, proposal.WithChannel(Definition.ChannelName), proposal.WithArguments(argsBytes))
}
