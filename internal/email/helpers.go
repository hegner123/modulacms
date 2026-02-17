package email

import (
	"encoding/base64"
	"net"
	"net/http"
	"time"
)

// encodeBase64 encodes raw bytes to a standard base64 string.
// Used by sendgrid and postmark senders for attachment encoding.
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// newHTTPClient creates an HTTP client with its own transport, isolated from
// other HTTP clients in the process (S3/bucket, etc.). This ensures that
// Close()/CloseIdleConnections() only affects the email sender's connections.
func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:         (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}
