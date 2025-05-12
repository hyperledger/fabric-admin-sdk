/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package osnadmin

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
)

// ListAllChannels that an OSN is a member of.
func ListAllChannels(osnURL string, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (*http.Response, error) {
	url := osnURL + "/participation/v1/channels"

	return httpGet(url, caCertPool, tlsClientCert)
}

// ListSingleChannel that an OSN is a member of.
func ListSingleChannel(osnURL, channelID string, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (*http.Response, error) {
	url := fmt.Sprintf("%s/participation/v1/channels/%s", osnURL, channelID)

	return httpGet(url, caCertPool, tlsClientCert)
}
