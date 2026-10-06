/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode_test

import (
	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseCollectionConfig", func() {
	It("converts a collection configuration JSON string", func() {
		collections, err := chaincode.ParseCollectionConfig(`[
			{
				"name": "collectionMarbles",
				"policy": "OR('Org1MSP.member', 'Org2MSP.member')",
				"requiredPeerCount": 0,
				"maxPeerCount": 3,
				"blockToLive": 1000000,
				"memberOnlyRead": true,
				"memberOnlyWrite": true
			}
		]`)
		Expect(err).NotTo(HaveOccurred())
		Expect(collections.GetConfig()).To(HaveLen(1))

		config := collections.GetConfig()[0].GetStaticCollectionConfig()
		Expect(config.GetName()).To(Equal("collectionMarbles"))
		Expect(config.GetRequiredPeerCount()).To(Equal(int32(0)))
		Expect(config.GetMaximumPeerCount()).To(Equal(int32(3)))
		Expect(config.GetBlockToLive()).To(Equal(uint64(1000000)))
		Expect(config.GetMemberOnlyRead()).To(BeTrue())
		Expect(config.GetMemberOnlyWrite()).To(BeTrue())
		Expect(config.GetEndorsementPolicy()).To(BeNil())

		policy, err := chaincode.SignaturePolicyEnvelopeToString(config.GetMemberOrgsPolicy().GetSignaturePolicy())
		Expect(err).NotTo(HaveOccurred())
		Expect(policy).To(Equal("OR('Org1MSP.member','Org2MSP.member')"))
	})

	It("converts multiple collections", func() {
		collections, err := chaincode.ParseCollectionConfig(`[
			{
				"name": "collectionOne",
				"policy": "OR('Org1MSP.member')"
			},
			{
				"name": "collectionTwo",
				"policy": "OR('Org2MSP.member')"
			}
		]`)
		Expect(err).NotTo(HaveOccurred())
		Expect(collections.GetConfig()).To(HaveLen(2))
		Expect(collections.GetConfig()[0].GetStaticCollectionConfig().GetName()).To(Equal("collectionOne"))
		Expect(collections.GetConfig()[1].GetStaticCollectionConfig().GetName()).To(Equal("collectionTwo"))
	})

	It("defaults the peer counts when they are omitted", func() {
		collections, err := chaincode.ParseCollectionConfig(`[
			{
				"name": "collectionPrivate",
				"policy": "OR('Org1MSP.member')"
			}
		]`)
		Expect(err).NotTo(HaveOccurred())

		config := collections.GetConfig()[0].GetStaticCollectionConfig()
		Expect(config.GetRequiredPeerCount()).To(Equal(int32(0)))
		Expect(config.GetMaximumPeerCount()).To(Equal(int32(1)))
	})

	It("supports an endorsement policy", func() {
		collections, err := chaincode.ParseCollectionConfig(`[
			{
				"name": "collectionMarbles",
				"policy": "OR('Org1MSP.member')",
				"endorsementPolicy": {
					"signaturePolicy": "OR('Org1MSP.member', 'Org2MSP.member')"
				}
			}
		]`)
		Expect(err).NotTo(HaveOccurred())

		endorsementPolicy := collections.GetConfig()[0].GetStaticCollectionConfig().GetEndorsementPolicy()
		Expect(endorsementPolicy).NotTo(BeNil())

		policy, err := chaincode.SignaturePolicyEnvelopeToString(endorsementPolicy.GetSignaturePolicy())
		Expect(err).NotTo(HaveOccurred())
		Expect(policy).To(Equal("OR('Org1MSP.member','Org2MSP.member')"))
	})

	It("returns an error for invalid JSON", func() {
		_, err := chaincode.ParseCollectionConfig(`not json`)
		Expect(err).To(HaveOccurred())
	})

	It("returns an error naming the collection for an invalid policy", func() {
		_, err := chaincode.ParseCollectionConfig(`[
			{
				"name": "collectionBad",
				"policy": "BOGUS('Org1MSP.member')"
			}
		]`)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("collectionBad"))
	})
})
