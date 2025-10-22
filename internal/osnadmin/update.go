package osnadmin

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"mime/multipart"
	"net/http"
)

// Update channel configuration using presented config envelope.
func Update(ctx context.Context, osnURL, channelID string, caCertPool *x509.CertPool, tlsClientCert tls.Certificate, configEnvelope []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s/participation/v1/channels", osnURL)

	req, err := createUpdateRequest(ctx, url, configEnvelope)
	if err != nil {
		return nil, fmt.Errorf("create update request: %w", err)
	}

	return httpDo(req, caCertPool, tlsClientCert)
}

func createUpdateRequest(ctx context.Context, url string, configEnvelope []byte) (*http.Request, error) {
	joinBody := new(bytes.Buffer)
	writer := multipart.NewWriter(joinBody)
	part, err := writer.CreateFormFile("config-update-envelope", "config_update.pb")
	if err != nil {
		return nil, err
	}
	_, err = part.Write(configEnvelope)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, joinBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}
