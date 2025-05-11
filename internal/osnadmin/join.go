/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package osnadmin

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"mime/multipart"
	"net/http"
)

// Join an OSN to a new or existing channel.
func Join(osnURL string, blockBytes []byte, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (*http.Response, error) {
	url := osnURL + "/participation/v1/channels"
	req, err := createJoinRequest(url, blockBytes)
	if err != nil {
		return nil, err
	}

	return httpDo(req, caCertPool, tlsClientCert)
}

func createJoinRequest(url string, blockBytes []byte) (*http.Request, error) {
	joinBody := new(bytes.Buffer)
	writer := multipart.NewWriter(joinBody)
	part, err := writer.CreateFormFile("config-block", "config.block")
	if err != nil {
		return nil, err
	}
	_, err = part.Write(blockBytes)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, joinBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}
