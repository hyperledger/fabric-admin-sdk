/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tools

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-admin-sdk/internal/configtxgen/encoder"
	"github.com/hyperledger/fabric-admin-sdk/internal/configtxgen/genesisconfig"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"

	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
)

// configtxGen
// base on Profile return block
func ConfigTxGen(config *genesisconfig.Profile, channelID string) (*cb.Block, error) {
	pgen, err := encoder.NewBootstrapper(config)
	if err != nil {
		return nil, err
	}
	genesisBlock := pgen.GenesisBlockForChannel(channelID)
	return genesisBlock, nil
}

// load profile
// file as file path
// profile_name name
func LoadProfile(configName, FABRIC_CFG_PATH string) (*genesisconfig.Profile, error) {
	return genesisconfig.Load(configName, FABRIC_CFG_PATH)
}

func CreateSigner(privateKeyPath, certificatePath, mspID string) (identity.SigningIdentity, error) {
	priv, err := GetPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}

	cert, _, err := GetCertificate(certificatePath)
	if err != nil {
		return nil, err
	}

	return identity.NewPrivateKeySigningIdentity(mspID, cert, priv)
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
		return nil, errors.New("Invalid PEM. It must be different from nil.")
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("Failed decoding PEM. Block must be different from nil. [% x]", raw)
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
			return nil, errors.New("Found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err = x509.ParseECPrivateKey(der); err == nil {
		return
	}

	return nil, errors.New("Invalid key type. The DER must contain an ecdsa.PrivateKey")
}
