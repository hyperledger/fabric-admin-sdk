/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package proposal

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"google.golang.org/protobuf/proto"
)

type transactionContext struct {
	TransactionID   string
	SignatureHeader *common.SignatureHeader
}

func newTransactionContext(id identity.Identity) (*transactionContext, error) {
	nonce := make([]byte, 24)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   id.MspID(),
		IdBytes: id.Credentials(),
	}
	creator, err := proto.Marshal(serializedIdentity)
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
