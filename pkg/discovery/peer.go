package discovery

import (
	"context"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/discovery"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type Peer struct {
	client      discovery.DiscoveryClient
	id          identity.SigningIdentity
	tlsCertHash []byte
}

func NewPeer(connection grpc.ClientConnInterface, id identity.SigningIdentity, options ...PeerOption) *Peer {
	result := &Peer{
		client: discovery.NewDiscoveryClient(connection),
		id:     id,
	}

	for _, option := range options {
		option(result)
	}

	return result
}

// PeerOption implements an option for creating a new Peer.
type PeerOption func(*Peer)

// WithTLSClientCertificateHash specifies the SHA-256 hash of the TLS client certificate. This option is required only
// if mutual TLS authentication is used for the gRPC connection to the peer.
func WithTLSClientCertificateHash(certificateHash []byte) PeerOption {
	return func(p *Peer) {
		p.tlsCertHash = certificateHash
	}
}

// PeerMembershipQuery returns information on peers that belong to the specified channel. If no filtering of results
// is required, nil can be supplied as the filter argument.
func (p *Peer) PeerMembershipQuery(ctx context.Context, channel string, filter *peer.ChaincodeInterest) (*discovery.PeerMembershipResult, error) {
	serializedID := &msp.SerializedIdentity{
		Mspid:   p.id.MspID(),
		IdBytes: p.id.Credentials(),
	}

	idBytes, err := proto.Marshal(serializedID)
	if err != nil {
		return nil, err
	}

	querys := []*discovery.Query{
		{
			Channel: channel,
			Query: &discovery.Query_PeerQuery{
				PeerQuery: &discovery.PeerMembershipQuery{
					Filter: filter,
				},
			},
		},
	}

	request := &discovery.Request{
		Authentication: &discovery.AuthInfo{
			ClientIdentity:    idBytes,
			ClientTlsCertHash: p.tlsCertHash,
		},
		Queries: querys,
	}

	payload, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	sig, err := p.id.Sign(payload)
	if err != nil {
		return nil, err
	}

	signedRequest := discovery.SignedRequest{
		Payload:   payload,
		Signature: sig,
	}

	rs, err := p.client.Discover(ctx, &signedRequest)
	if err != nil {
		return nil, err
	}

	for _, qrs := range rs.Results {
		return qrs.GetMembers(), nil
	}
	return nil, nil
}
