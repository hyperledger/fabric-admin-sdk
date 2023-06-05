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

func PeerMembershipQuery(ctx context.Context, conn *grpc.ClientConn, signer identity.SigningIdentity, channel string, filter *peer.ChaincodeInterest) (*discovery.PeerMembershipResult, error) {
	id := &msp.SerializedIdentity{
		Mspid:   signer.MspID(),
		IdBytes: signer.Credentials(),
	}

	idBytes, err := proto.Marshal(id)
	if err != nil {
		return nil, err
	}

	querys := []*discovery.Query{
		&discovery.Query{
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
			ClientTlsCertHash: signer.Credentials(),
		},
		Queries: querys,
	}

	payload, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(payload)
	if err != nil {
		return nil, err
	}

	signedRequest := discovery.SignedRequest{
		Payload:   payload,
		Signature: sig,
	}

	cli := discovery.NewDiscoveryClient(conn)

	rs, err := cli.Discover(ctx, &signedRequest)
	if err != nil {
		return nil, err
	}

	for _, qrs := range rs.Results {
		return qrs.GetMembers(), nil
	}
	return nil, nil
}
