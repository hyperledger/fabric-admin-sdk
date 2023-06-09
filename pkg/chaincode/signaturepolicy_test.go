package chaincode

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("signaturepolicy", func() {
	Context("signaturepolicyenvelope to string", func() {
		It("should equal", func(specCtx SpecContext) {
			expression1 := `OR('Org3MSP.peer','Org1MSP.admin','Org2MSP.member')`
			expression2 := `AND('Org3MSP.peer','Org1MSP.admin','Org2MSP.member')`
			expression3 := `AND('Org3MSP.peer',OR('Org1MSP.admin','Org2MSP.member'))`
			expression4 := `OutOf(2,'Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')`

			expressions := []string{expression1, expression2, expression3, expression4}

			for _, expression := range expressions {
				//gen a SignaturePolicyEnvelope
				applicationPolicy, err := NewApplicationPolicy(expression, "")
				Expect(err).NotTo(HaveOccurred())
				policy := applicationPolicy.GetSignaturePolicy()

				//parse the SignaturePolicyEnvelope back to expression
				dstExpression := SignaturePolicyEnvelopeToString(policy)

				fmt.Println("src Expression", expression)
				fmt.Println("dst Expression", dstExpression)

				Expect(dstExpression).To(Equal(expression))

			}
		})
	})
})

func TestSignaturePolicy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "signaturepolicy")
}
