package osnadmin

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"google.golang.org/protobuf/proto"
)

func Fetch(osnURL, channelID string, blockID string, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (*common.Block, error) {
	url := fmt.Sprintf("%s/participation/v1/channels/%s/blocks/%s", osnURL, channelID, blockID)

	resp, err := httpGet(url, caCertPool, tlsClientCert)
	if err != nil {
		return nil, fmt.Errorf("process request: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var block common.Block
	if err := proto.Unmarshal(body, &block); err != nil {
		return nil, fmt.Errorf("unmarshal response body: %w", err)
	}
	return &block, nil
}
