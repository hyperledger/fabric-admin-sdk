package protoutil

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/internal/pkg/identity"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

// CreateSignedTx assembles an Envelope message from proposal, endorsements,
// and a signer. This function should be called by a client when it has
// collected enough endorsements for a proposal to create a transaction and
// submit it to peers for ordering
func CreateSignedTx(
	proposal *peer.Proposal,
	signer identity.Signer,
	resps ...*peer.ProposalResponse,
) (*common.Envelope, error) {
	if err := ensureValidResponses(resps); err != nil {
		return nil, err
	}

	// the original header
	hdr, err := UnmarshalHeader(proposal.Header)
	if err != nil {
		return nil, err
	}

	// the original payload
	pPayl, err := UnmarshalChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, err
	}

	endorsements := fillEndorsements(resps)

	// create ChaincodeEndorsedAction
	cea := &peer.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].Payload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := GetBytesProposalPayloadForTx(pPayl)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &peer.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &peer.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*peer.TransactionAction, 1)
	taas[0] = taa
	tx := &peer.Transaction{Actions: taas}

	// serialize the tx
	txBytes, err := GetBytesTransaction(tx)
	if err != nil {
		return nil, err
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	paylBytes, err := GetBytesPayload(payl)
	if err != nil {
		return nil, err
	}

	// sign the payload
	sig, err := signer.Sign(paylBytes)
	if err != nil {
		return nil, err
	}

	// here's the envelope
	return &common.Envelope{Payload: paylBytes, Signature: sig}, nil
}

// ensureValidResponses checks that all actions are bitwise equal and that they are successful.
func ensureValidResponses(responses []*peer.ProposalResponse) error {
	if len(responses) == 0 {
		return errors.New("at least one proposal response is required")
	}

	var firstResponse []byte
	for n, r := range responses {
		if r.Response.Status < 200 || r.Response.Status >= 400 {
			return fmt.Errorf("proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
		}

		if n == 0 {
			firstResponse = r.Payload
		} else if !bytes.Equal(firstResponse, r.Payload) {
			return errors.New("ProposalResponsePayloads do not match")
		}
	}

	return nil
}

func fillEndorsements(responses []*peer.ProposalResponse) []*peer.Endorsement {
	endorsements := make([]*peer.Endorsement, len(responses))
	for n, r := range responses {
		endorsements[n] = r.Endorsement
	}
	return endorsements
}
