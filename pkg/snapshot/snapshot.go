package snapshot

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// Peer is a wrapper around snapshot client
type Peer struct {
	snapshot peer.SnapshotClient
	id       identity.SigningIdentity
}

func NewPeer(conn grpc.ClientConnInterface, id identity.SigningIdentity) *Peer {
	return &Peer{
		snapshot: peer.NewSnapshotClient(conn),
		id:       id,
	}
}

// SubmitRequest snapshot from a specific peer
func (p *Peer) SubmitRequest(ctx context.Context, channelID string, blockNum uint64) error {
	signedRequest, err := p.newSignedSnapshotRequest(channelID, blockNum)
	if err != nil {
		return fmt.Errorf("failed to create signed snapshot request: %w", err)
	}

	if _, err := p.snapshot.Generate(ctx, signedRequest); err != nil {
		return fmt.Errorf("failed to submit generate snapshot request: %w", err)
	}

	return nil
}

// CancelRequest snapshot from a specific peer
func (p *Peer) CancelRequest(ctx context.Context, channelID string) error {
	signedRequest, err := p.newSignedSnapshotRequest(channelID, 0)
	if err != nil {
		return fmt.Errorf("failed to create signed snapshot request: %w", err)
	}

	if _, err := p.snapshot.Cancel(ctx, signedRequest); err != nil {
		return fmt.Errorf("failed to cancel snapshot request: %w", err)
	}

	return nil
}

// QueryPending snapshots from a specific peer
func (p *Peer) QueryPending(ctx context.Context, channelID string) (*peer.QueryPendingSnapshotsResponse, error) {
	signedRequest, err := p.newSignedSnapshotRequest(channelID, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed snapshot request: %w", err)
	}

	res, err := p.snapshot.QueryPendings(ctx, signedRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending snapshots: %w", err)
	}

	return res, nil
}

// newSignedSnapshotRequest returns a signed snapshot request for
// given channel
func (p *Peer) newSignedSnapshotRequest(channelID string, blockNum uint64) (*peer.SignedSnapshotRequest, error) {
	nonce, err := protoutil.CreateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to create nonce: %w", err)
	}

	request, err := proto.Marshal(&peer.SnapshotRequest{
		SignatureHeader: &common.SignatureHeader{
			Creator: p.id.Credentials(),
			Nonce:   nonce,
		},
		ChannelId:   channelID,
		BlockNumber: blockNum,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot request: %w", err)
	}

	signature, err := p.id.Sign(request)
	if err != nil {
		return nil, fmt.Errorf("failed to sign snapshot request: %w", err)
	}

	return &peer.SignedSnapshotRequest{
		Request:   request,
		Signature: signature,
	}, nil
}
