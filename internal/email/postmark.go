package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/config"
)

const postmarkDefaultEndpoint = "https://api.postmarkapp.com/email"

// PostmarkSender sends email via the Postmark API.
type PostmarkSender struct {
	serverToken string
	endpoint    string
	client      *http.Client
}

// newPostmarkSender creates a PostmarkSender from config.
func newPostmarkSender(cfg config.Config) (*PostmarkSender, error) {
	if cfg.Email_API_Key == "" {
		return nil, fmt.Errorf("email_api_key is required for Postmark provider")
	}

	endpoint := cfg.Email_API_Endpoint
	if endpoint == "" {
		endpoint = postmarkDefaultEndpoint
	}

	return &PostmarkSender{
		serverToken: cfg.Email_API_Key,
		endpoint:    endpoint,
		client:      newHTTPClient(30 * time.Second),
	}, nil
}

// postmarkPayload is the JSON structure for the Postmark email API.
type postmarkPayload struct {
	From        string                `json:"From"`
	To          string                `json:"To"`
	Cc          string                `json:"Cc,omitempty"`
	Bcc         string                `json:"Bcc,omitempty"`
	Subject     string                `json:"Subject"`
	TextBody    string                `json:"TextBody,omitempty"`
	HtmlBody    string                `json:"HtmlBody,omitempty"`
	ReplyTo     string                `json:"ReplyTo,omitempty"`
	Attachments []postmarkAttachment  `json:"Attachments,omitempty"`
}

type postmarkAttachment struct {
	Name        string `json:"Name"`
	Content     string `json:"Content"`
	ContentType string `json:"ContentType"`
}

// Send delivers a message via the Postmark API.
func (p *PostmarkSender) Send(ctx context.Context, msg Message) error {
	payload := postmarkPayload{
		From:    msg.From.String(),
		To:      joinAddresses(msg.To),
		Subject: msg.Subject,
	}

	if len(msg.CC) > 0 {
		payload.Cc = joinAddresses(msg.CC)
	}
	if len(msg.BCC) > 0 {
		payload.Bcc = joinAddresses(msg.BCC)
	}
	if msg.PlainBody != "" {
		payload.TextBody = msg.PlainBody
	}
	if msg.HTMLBody != "" {
		payload.HtmlBody = msg.HTMLBody
	}
	if msg.ReplyTo != nil {
		payload.ReplyTo = msg.ReplyTo.String()
	}

	for _, att := range msg.Attachments {
		payload.Attachments = append(payload.Attachments, postmarkAttachment{
			Name:        att.Filename,
			Content:     encodeBase64(att.Data),
			ContentType: att.ContentType,
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &ProviderError{Provider: "postmark", Err: fmt.Errorf("marshal payload: %w", err)}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return &ProviderError{Provider: "postmark", Err: fmt.Errorf("create request: %w", err)}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Postmark-Server-Token", p.serverToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return &ProviderError{Provider: "postmark", Err: err}
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return &ProviderError{
			Provider: "postmark",
			Code:     resp.StatusCode,
			Err:      fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	return nil
}

// Close releases idle connections in the HTTP client.
func (p *PostmarkSender) Close() error {
	p.client.CloseIdleConnections()
	return nil
}
