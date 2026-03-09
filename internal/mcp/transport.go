package mcp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
)

// directTransport implements http.RoundTripper by routing requests directly
// to an in-process http.Handler. This eliminates the HTTP loopback where
// MCP tools would otherwise call the CMS API over localhost.
//
// The handler is typically the full middleware-wrapped ServeMux, so requests
// go through the same auth, permission, CORS, and audit middleware as
// external HTTP requests — just without network overhead or serialization
// through a TCP socket.
type directTransport struct {
	handler http.Handler
}

// RoundTrip satisfies http.RoundTripper. It captures the handler's response
// using httptest.ResponseRecorder and converts it to an *http.Response.
func (t *directTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	t.handler.ServeHTTP(rec, req)

	resp := &http.Response{
		StatusCode:    rec.Code,
		Status:        strconv.Itoa(rec.Code) + " " + http.StatusText(rec.Code),
		Header:        rec.Header(),
		Body:          io.NopCloser(bytes.NewReader(rec.Body.Bytes())),
		ContentLength: int64(rec.Body.Len()),
		Request:       req,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
	}

	return resp, nil
}
