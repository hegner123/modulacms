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

const sendGridDefaultEndpoint = "https://api.sendgrid.com/v3/mail/send"

// SendGridSender sends email via the SendGrid v3 API.
type SendGridSender struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

// newSendGridSender creates a SendGridSender from config.
func newSendGridSender(cfg config.Config) (*SendGridSender, error) {
	if cfg.Email_API_Key == "" {
		return nil, fmt.Errorf("email_api_key is required for SendGrid provider")
	}

	endpoint := cfg.Email_API_Endpoint
	if endpoint == "" {
		endpoint = sendGridDefaultEndpoint
	}

	return &SendGridSender{
		apiKey:   cfg.Email_API_Key,
		endpoint: endpoint,
		client:   newHTTPClient(30 * time.Second),
	}, nil
}

// sendGridPayload is the JSON structure for the SendGrid v3 API.
type sendGridPayload struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
	ReplyTo          *sendGridEmail            `json:"reply_to,omitempty"`
	Attachments      []sendGridAttachment      `json:"attachments,omitempty"`
}

type sendGridPersonalization struct {
	To  []sendGridEmail `json:"to"`
	CC  []sendGridEmail `json:"cc,omitempty"`
	BCC []sendGridEmail `json:"bcc,omitempty"`
}

type sendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type sendGridAttachment struct {
	Content     string `json:"content"`
	Type        string `json:"type"`
	Filename    string `json:"filename"`
	Disposition string `json:"disposition"`
}

// Send delivers a message via the SendGrid v3 API.
func (s *SendGridSender) Send(ctx context.Context, msg Message) error {
	payload := sendGridPayload{
		From:    sendGridEmail{Email: msg.From.Address, Name: msg.From.Name},
		Subject: msg.Subject,
	}

	pers := sendGridPersonalization{}
	for _, addr := range msg.To {
		pers.To = append(pers.To, sendGridEmail{Email: addr.Address, Name: addr.Name})
	}
	for _, addr := range msg.CC {
		pers.CC = append(pers.CC, sendGridEmail{Email: addr.Address, Name: addr.Name})
	}
	for _, addr := range msg.BCC {
		pers.BCC = append(pers.BCC, sendGridEmail{Email: addr.Address, Name: addr.Name})
	}
	payload.Personalizations = []sendGridPersonalization{pers}

	if msg.PlainBody != "" {
		payload.Content = append(payload.Content, sendGridContent{Type: "text/plain", Value: msg.PlainBody})
	}
	if msg.HTMLBody != "" {
		payload.Content = append(payload.Content, sendGridContent{Type: "text/html", Value: msg.HTMLBody})
	}

	if msg.ReplyTo != nil {
		payload.ReplyTo = &sendGridEmail{Email: msg.ReplyTo.Address, Name: msg.ReplyTo.Name}
	}

	for _, att := range msg.Attachments {
		payload.Attachments = append(payload.Attachments, sendGridAttachment{
			Content:     encodeBase64(att.Data),
			Type:        att.ContentType,
			Filename:    att.Filename,
			Disposition: "attachment",
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &ProviderError{Provider: "sendgrid", Err: fmt.Errorf("marshal payload: %w", err)}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return &ProviderError{Provider: "sendgrid", Err: fmt.Errorf("create request: %w", err)}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return &ProviderError{Provider: "sendgrid", Err: err}
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return &ProviderError{
			Provider: "sendgrid",
			Code:     resp.StatusCode,
			Err:      fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	return nil
}

// Close releases idle connections in the HTTP client.
func (s *SendGridSender) Close() error {
	s.client.CloseIdleConnections()
	return nil
}
