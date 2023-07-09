/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
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

func CreateSigner(privateKeyPath, certificatePath, mspID string) (SigningIdentity, error) {
	priv, err := GetPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}

	cert, _, err := GetCertificate(certificatePath)
	if err != nil {
		return nil, err
	}

	return NewPrivateKeySigningIdentity(mspID, cert, priv)
}

func GetPrivateKey(f string) (*ecdsa.PrivateKey, error) {
	in, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	k, err := PEMtoPrivateKey(in)
	if err != nil {
		return nil, err
	}

	key, ok := k.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("expecting ecdsa key")
	}

	return key, nil
}

func GetCertificate(f string) (*x509.Certificate, []byte, error) {
	in, err := os.ReadFile(f)
	if err != nil {
		return nil, nil, err
	}

	block, _ := pem.Decode(in)

	c, err := x509.ParseCertificate(block.Bytes)
	return c, in, err
}

// PEMtoPrivateKey unmarshals a pem to private key
func PEMtoPrivateKey(raw []byte) (interface{}, error) {
	if len(raw) == 0 {
		return nil, errors.New("invalid PEM. It must be different from nil")
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("failed decoding PEM. Block must be different from nil. [% x]", raw)
	}
	cert, err := DERToPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, err
}

// DERToPrivateKey unmarshals a der to private key
func DERToPrivateKey(der []byte) (key interface{}, err error) {

	if key, err = x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}

	if key, err = x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key.(type) {
		case *ecdsa.PrivateKey:
			return
		default:
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err = x509.ParseECPrivateKey(der); err == nil {
		return
	}

	return nil, errors.New("invalid key type. The DER must contain an ecdsa.PrivateKey")
}
