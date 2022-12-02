/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package identity_test

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"

	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func DecodeCertificatePEM(certificatePEM []byte) *x509.Certificate {
	block, _ := pem.Decode(certificatePEM)
	certificate, err := x509.ParseCertificate(block.Bytes)
	Expect(err).NotTo(HaveOccurred())

	return certificate
}

var _ = Describe("SigningIdentity", func() {
	var certificate *x509.Certificate
	var privateKey *ecdsa.PrivateKey

	BeforeEach(func() {
		var err error

		privateKey, err = NewECDSAPrivateKey()
		Expect(err).NotTo(HaveOccurred())

		certificate, err = NewCertificate(privateKey)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Has MSP ID", func(specCtx SpecContext) {
		id, err := identity.NewPrivateKeySigningIdentity("MSP_ID", certificate, privateKey)
		Expect(err).NotTo(HaveOccurred())

		Expect(id.MspID()).To(Equal("MSP_ID"))
	})

	It("Has certificate", func(specCtx SpecContext) {
		id, err := identity.NewPrivateKeySigningIdentity("MSP_ID", certificate, privateKey)
		Expect(err).NotTo(HaveOccurred())

		actual := DecodeCertificatePEM(id.Credentials())
		Expect(actual).To(Equal(certificate))
	})

	It("Creates valid signature", func(specCtx SpecContext) {
		id, err := identity.NewPrivateKeySigningIdentity("", certificate, privateKey)
		Expect(err).NotTo(HaveOccurred())

		message := []byte("MESSAGE")

		signature, err := id.Sign(message)
		Expect(err).NotTo(HaveOccurred())

		hash := sha256.Sum256(message)

		valid := ecdsa.VerifyASN1(&privateKey.PublicKey, hash[:], signature)
		Expect(valid).To(BeTrue())
	})
})
