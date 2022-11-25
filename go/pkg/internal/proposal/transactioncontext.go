/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package proposal

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fabric-admin-sdk/internal/pkg/identity"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
)

type transactionContext struct {
	TransactionID   string
	SignatureHeader *common.SignatureHeader
}

func newTransactionContext(serializer identity.Serializer) (*transactionContext, error) {
	nonce := make([]byte, 24)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	creator, err := serializer.Serialize()
	if err != nil {
		return nil, err
	}

	saltedCreator := append(nonce, creator...)
	rawTransactionID := sha256.Sum256(saltedCreator)
	transactionID := hex.EncodeToString(rawTransactionID[:])

	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}

	transactionCtx := &transactionContext{
		TransactionID:   transactionID,
		SignatureHeader: signatureHeader,
	}
	return transactionCtx, nil
}
