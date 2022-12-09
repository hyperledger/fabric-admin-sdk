/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
)

// Identity used to interact with a Fabric network.
type Identity interface {
	MspID() string       // ID of the Membership Service Provider to which this identity belongs.
	Credentials() []byte // Implementation-specific credentials.
}

// Signer can sign messages using an identity's private credentials.
type Signer interface {
	Sign(message []byte) ([]byte, error)
}

// SigningIdentity represents an identity that is able to sign messages.
type SigningIdentity interface {
	Identity
	Signer
}

func NewPrivateKeySigningIdentity(mspID string, certificate *x509.Certificate, privateKey crypto.PrivateKey) (SigningIdentity, error) {
	credentials, err := certificateToPEM(certificate)
	if err != nil {
		return nil, err
	}

	sign, err := newPrivateKeySign(privateKey)
	if err != nil {
		return nil, err
	}

	id := &signingIdentity{
		mspID:       mspID,
		credentials: credentials,
		sign:        sign,
	}
	return id, nil
}

type signFn func([]byte) ([]byte, error)

func newPrivateKeySign(privateKey crypto.PrivateKey) (signFn, error) {
	switch key := privateKey.(type) {
	case *ecdsa.PrivateKey:
		return ecdsaPrivateKeySign(key), nil
	default:
		return nil, fmt.Errorf("unsupported key type: %T", privateKey)
	}
}

type signingIdentity struct {
	mspID       string
	credentials []byte
	sign        signFn
}

func (id *signingIdentity) MspID() string {
	return id.mspID
}

func (id *signingIdentity) Credentials() []byte {
	return id.credentials
}

func (id *signingIdentity) Sign(message []byte) ([]byte, error) {
	return id.sign(message)
}
