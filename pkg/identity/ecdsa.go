/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto/ecdsa"

	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

func ecdsaPrivateKeySign(privateKey *ecdsa.PrivateKey) (signFn, error) {
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, err
	}

	result := func(message []byte) ([]byte, error) {
		digest := hash.SHA256(message)
		return sign(digest)
	}
	return result, nil
}
