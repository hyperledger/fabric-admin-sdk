package channel

import (
	"crypto/tls"
	"crypto/x509"
	"fabric-admin-sdk/internal/osnadmin"
	"net/http"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"
)

func CreateChannel(osnURL string, block *cb.Block, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (*http.Response, error) {
	block_byte := protoutil.MarshalOrPanic(block)
	return osnadmin.Join(osnURL, block_byte, caCertPool, tlsClientCert)
}
