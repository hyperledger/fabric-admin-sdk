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

// Client is a wrapper around snapshot client
type Client struct {
	snapshot peer.SnapshotClient
	id       identity.SigningIdentity
}

func New(conn grpc.ClientConnInterface, id identity.SigningIdentity) *Client {
	return &Client{
		snapshot: peer.NewSnapshotClient(conn),
		id:       id,
	}
}

// SubmitSnapshotRequest from a specific peer
func (s *Client) SubmitSnapshotRequest(ctx context.Context, channelID string, blockNum uint64) error {
	signedRequest, err := s.newSignedSnapshotRequest(channelID, blockNum)
	if err != nil {
		return fmt.Errorf("failed to create signed snapshot request: %w", err)
	}

	if _, err := s.snapshot.Generate(ctx, signedRequest); err != nil {
		return fmt.Errorf("failed to submit generate snapshot request: %w", err)
	}

	return nil
}

// CancelSnapshotRequest from a specific peer
func (s *Client) CancelSnapshotRequest(ctx context.Context, channelID string) error {
	signedRequest, err := s.newSignedSnapshotRequest(channelID, 0)
	if err != nil {
		return fmt.Errorf("failed to create signed snapshot request: %w", err)
	}

	if _, err := s.snapshot.Cancel(ctx, signedRequest); err != nil {
		return fmt.Errorf("failed to cancel snapshot request: %w", err)
	}

	return nil
}

// ListPendingSnapshots from a specific peer
func (s *Client) ListPendingSnapshots(ctx context.Context, channelID string) error {
	signedRequest, err := s.newSignedSnapshotRequest(channelID, 0)
	if err != nil {
		return fmt.Errorf("failed to create signed snapshot request: %w", err)
	}

	if _, err := s.snapshot.QueryPendings(ctx, signedRequest); err != nil {
		return fmt.Errorf("failed to list pending snapshots: %w", err)
	}

	return nil
}

// newSignedSnapshotRequest returns a signed snapshot request for
// given channel
func (s *Client) newSignedSnapshotRequest(channelID string, blockNum uint64) (*peer.SignedSnapshotRequest, error) {
	nonce, err := protoutil.CreateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to create nonce: %w", err)
	}

	request, err := proto.Marshal(&peer.SnapshotRequest{
		SignatureHeader: &common.SignatureHeader{
			Creator: s.id.Credentials(),
			Nonce:   nonce,
		},
		ChannelId:   channelID,
		BlockNumber: blockNum,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot request: %w", err)
	}

	signature, err := s.id.Sign(request)
	if err != nil {
		return nil, fmt.Errorf("failed to sign snapshot request: %w", err)
	}

	return &peer.SignedSnapshotRequest{
		Request:   request,
		Signature: signature,
	}, nil
}
