package common

import (
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/network"
	"google.golang.org/grpc"
)

type ConnectionDetail struct {
	ID         identity.SigningIdentity
	Connection grpc.ClientConnInterface
}

func ConstructConnectionDetail(mspID string, SignCert string, PrivKeyPath string, orgPeerAddress string, TLSCACert string) (*ConnectionDetail, error) {
	peer := network.Node{
		Addr:      orgPeerAddress,
		TLSCACert: TLSCACert,
	}

	err := peer.LoadConfig()
	if err != nil {
		return nil, err
	}
	peerConnection, err := network.DialConnection(peer)
	if err != nil {
		return nil, err
	}

	cert, _, err := identity.ReadCertificate(SignCert)
	if err != nil {
		return nil, err
	}

	priv, err := identity.ReadECDSAPrivateKey(PrivKeyPath)
	if err != nil {
		return nil, err
	}

	orgMSP, err := identity.NewPrivateKeySigningIdentity(mspID, cert, priv)
	if err != nil {
		return nil, err
	}
	return &ConnectionDetail{
		Connection: peerConnection,
		ID:         orgMSP,
	}, nil
}
