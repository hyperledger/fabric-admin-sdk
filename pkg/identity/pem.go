/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
)

func certificateToPEM(certificate *x509.Certificate) ([]byte, error) {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate.Raw,
	}
	return pemEncode(block)
}

func pemEncode(block *pem.Block) ([]byte, error) {
	var buffer bytes.Buffer
	if err := pem.Encode(&buffer, block); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
