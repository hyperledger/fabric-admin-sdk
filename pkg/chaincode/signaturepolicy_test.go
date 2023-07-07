package chaincode

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("signaturepolicyenvelope to string",
	func(expression string) {
		//gen a SignaturePolicyEnvelope from expression
		applicationPolicy, err := NewApplicationPolicy(expression, "")
		Expect(err).NotTo(HaveOccurred())
		policy := applicationPolicy.GetSignaturePolicy()

		//parse the SignaturePolicyEnvelope back to expression
		dstExpression, err := SignaturePolicyEnvelopeToString(policy)
		Expect(err).NotTo(HaveOccurred())

		fmt.Println("src Expression:", expression)
		fmt.Println("dst Expression:", dstExpression)

		Expect(dstExpression).To(Equal(expression))
	},
	Entry("When keyword has OR", `OR('Org3MSP.peer','Org1MSP.admin','Org2MSP.member')`),
	Entry("When keyword has AND", `AND('Org3MSP.peer','Org1MSP.admin','Org2MSP.member')`),
	Entry("When keyword has AND,OR", `AND('Org3MSP.peer',OR('Org1MSP.admin','Org2MSP.member'))`),
	Entry("When keyword has OutOf", `OutOf(2,'Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')`),
)
