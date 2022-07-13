package e2e_test

import (
	"crypto/tls"
	"crypto/x509"
	"fabric-admin-sdk/channel"
	"fabric-admin-sdk/tools"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("e2e", func() {
	Context("Create channel", func() {
		It("should create channel with test-network", func() {
			_, err := os.Stat("../../fabric-samples/test-network")
			if err != nil {
				ginkgo.Skip("skip for unit test")
			}
			profile, err := tools.LoadProfile("TwoOrgsApplicationGenesis", "./")
			Expect(err).NotTo(HaveOccurred())
			Expect(profile).ToNot(BeNil())
			Expect(profile.Orderer.BatchSize.MaxMessageCount).To(Equal(uint32(10)))
			block, err := tools.ConfigTxGen(profile, "mychannel")
			Expect(err).NotTo(HaveOccurred())
			Expect(block).ToNot(BeNil())

			var caFile, clientCert, clientKey, osnURL string
			osnURL = "https://localhost:7053"
			caFile = "../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
			clientCert = "../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt"
			clientKey = "../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key"
			caCertPool := x509.NewCertPool()
			caFilePEM, err := ioutil.ReadFile(caFile)
			caCertPool.AppendCertsFromPEM(caFilePEM)
			Expect(err).NotTo(HaveOccurred())
			tlsClientCert, err := tls.LoadX509KeyPair(clientCert, clientKey)
			Expect(err).NotTo(HaveOccurred())
			resp, err := channel.CreateChannel(osnURL, block, caCertPool, tlsClientCert)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("myHttpGet error is ", err)
			}
			fmt.Println("response statuscode is ", resp.StatusCode,
				"\nhead[name]=", resp.Header["Name"],
				"\nbody is ", string(body))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusCreated))
		})
	})
})
