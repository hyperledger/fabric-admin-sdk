package osnadmin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"google.golang.org/protobuf/proto"
)

func Fetch(ctx context.Context, osnURL, channelID string, blockID string, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (*common.Block, error) {
	url := fmt.Sprintf("%s/participation/v1/channels/%s/blocks/%s", osnURL, channelID, blockID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := httpDo(req, caCertPool, tlsClientCert)
	if err != nil {
		return nil, fmt.Errorf("process request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
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
