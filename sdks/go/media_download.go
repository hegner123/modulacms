package modula

import (
	"context"
	"fmt"
	"net/http"
)

// MediaDownloadResource provides media file download operations. The server
// generates a pre-signed S3 URL and returns a 302 redirect. These methods
// extract the redirect URL without following it, so the caller can use the
// URL directly (e.g. for streaming downloads or passing to a frontend).
// It is accessed via [Client].MediaDownload.
type MediaDownloadResource struct {
	http *httpClient
}

// GetURL returns a pre-signed download URL for the given media item.
// The URL is valid for approximately 15 minutes. Returns the URL string
// from the server's 302 redirect Location header.
func (d *MediaDownloadResource) GetURL(ctx context.Context, id MediaID) (string, error) {
	return d.getDownloadURL(ctx, "/api/v1/media/"+string(id)+"/download")
}

// AdminGetURL returns a pre-signed download URL for the given admin media item.
// The URL is valid for approximately 15 minutes. Returns the URL string
// from the server's 302 redirect Location header.
func (d *MediaDownloadResource) AdminGetURL(ctx context.Context, id AdminMediaID) (string, error) {
	return d.getDownloadURL(ctx, "/api/v1/adminmedia/"+string(id)+"/download")
}

func (d *MediaDownloadResource) getDownloadURL(ctx context.Context, path string) (string, error) {
	fullURL := d.http.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("modula: %w", err)
	}

	// Use doRaw so we can inspect the redirect response without following it.
	// We need a client that does not follow redirects.
	noRedirectClient := *d.http.httpClient
	noRedirectClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	d.http.setAuth(req)
	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("modula: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusTemporaryRedirect {
		loc := resp.Header.Get("Location")
		if loc == "" {
			return "", fmt.Errorf("modula: redirect response missing Location header")
		}
		return loc, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", d.http.buildError(resp)
	}

	// Unexpected 2xx without redirect — fall back to Location if present.
	loc := resp.Header.Get("Location")
	if loc != "" {
		return loc, nil
	}

	return "", fmt.Errorf("modula: unexpected status %d from download endpoint", resp.StatusCode)
}
