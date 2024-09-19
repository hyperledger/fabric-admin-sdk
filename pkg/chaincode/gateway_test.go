/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package chaincode_test

import (
	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Gateway", func() {
	It("ClientIdentity", func() {
		controller := gomock.NewController(GinkgoT())
		defer controller.Finish()

		mockConnection := NewMockClientConnInterface(controller)
		mockSigner := NewMockSigner(controller, "", nil, nil)
		gateway := chaincode.NewPeer(mockConnection, mockSigner)

		actual := gateway.ClientIdentity()

		Expect(actual).To(BeIdenticalTo(mockSigner))
	})
})
