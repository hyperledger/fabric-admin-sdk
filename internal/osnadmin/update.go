package osnadmin

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"mime/multipart"
	"net/http"
)

// Update channel configuration using presented config envelope.
func Update(osnURL, channelID string, caCertPool *x509.CertPool, tlsClientCert tls.Certificate, configEnvelope []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s/participation/v1/channels/%s", osnURL, channelID)

	req, err := createUpdateRequest(url, configEnvelope)
	if err != nil {
		return nil, fmt.Errorf("create update request: %w", err)
	}

	return httpDo(req, caCertPool, tlsClientCert)
}

func createUpdateRequest(url string, configEnvelope []byte) (*http.Request, error) {
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

	req, err := http.NewRequest(http.MethodPut, url, joinBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}
