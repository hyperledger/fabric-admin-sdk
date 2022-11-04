package e2e_test

import (
	"crypto/tls"
	"crypto/x509"
	"fabric-admin-sdk/channel"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("channel", func() {
	Context("get channel list", func() {
		It("should work", func() {
			_, err := os.Stat("../../../fabric-samples/test-network")
			if err != nil {
				ginkgo.Skip("skip for unit test")
			}
			var caFile, clientCert, clientKey, osnURL string
			osnURL = "https://orderer.example.com:7053"
			caFile = "../../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
			clientCert = "../../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt"
			clientKey = "../../../fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key"
			caCertPool := x509.NewCertPool()
			caFilePEM, err := ioutil.ReadFile(caFile)
			caCertPool.AppendCertsFromPEM(caFilePEM)
			Expect(err).NotTo(HaveOccurred())
			tlsClientCert, err := tls.LoadX509KeyPair(clientCert, clientKey)
			Expect(err).NotTo(HaveOccurred())
			channels, err := channel.ListChannel(osnURL, caCertPool, tlsClientCert)
			Expect(err).NotTo(HaveOccurred())
			for _, v := range channels.Channels {
				fmt.Println("channel name: ", v.Name)
			}
		})
	})
})
