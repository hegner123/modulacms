package email

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

func TestSendGridSender_CorrectPayload(t *testing.T) {
	t.Parallel()

	var receivedBody []byte
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(202)
	}))
	defer server.Close()

	sender := &SendGridSender{
		apiKey:   "test-key",
		endpoint: server.URL,
		client:   server.Client(),
	}

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Hello",
		PlainBody: "body",
	}

	err := sender.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if receivedAuth != "Bearer test-key" {
		t.Errorf("auth header = %q, want %q", receivedAuth, "Bearer test-key")
	}

	var payload sendGridPayload
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.From.Email != "sender@example.com" {
		t.Errorf("from = %q, want %q", payload.From.Email, "sender@example.com")
	}
	if payload.Subject != "Hello" {
		t.Errorf("subject = %q, want %q", payload.Subject, "Hello")
	}
}

func TestSendGridSender_4xxReturnsProviderError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"errors":[{"message":"bad request"}]}`))
	}))
	defer server.Close()

	sender := &SendGridSender{
		apiKey:   "test-key",
		endpoint: server.URL,
		client:   server.Client(),
	}

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Hello",
		PlainBody: "body",
	}

	err := sender.Send(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}

	var provErr *ProviderError
	if !errors.As(err, &provErr) {
		t.Fatalf("expected *ProviderError, got %T", err)
	}
	if provErr.Code != 400 {
		t.Errorf("Code = %d, want 400", provErr.Code)
	}
}

func TestPostmarkSender_CorrectHeaders(t *testing.T) {
	t.Parallel()

	var receivedToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedToken = r.Header.Get("X-Postmark-Server-Token")
		w.WriteHeader(200)
		w.Write([]byte(`{"MessageID":"test"}`))
	}))
	defer server.Close()

	sender := &PostmarkSender{
		serverToken: "pm-token-123",
		endpoint:    server.URL,
		client:      server.Client(),
	}

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Hello",
		PlainBody: "body",
	}

	err := sender.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if receivedToken != "pm-token-123" {
		t.Errorf("token header = %q, want %q", receivedToken, "pm-token-123")
	}
}

func TestPostmarkSender_AttachmentBase64(t *testing.T) {
	t.Parallel()

	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(200)
		w.Write([]byte(`{"MessageID":"test"}`))
	}))
	defer server.Close()

	sender := &PostmarkSender{
		serverToken: "pm-token",
		endpoint:    server.URL,
		client:      server.Client(),
	}

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Hello",
		PlainBody: "body",
		Attachments: []Attachment{
			{Filename: "test.txt", ContentType: "text/plain", Data: []byte("hello world")},
		},
	}

	err := sender.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	var payload postmarkPayload
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(payload.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(payload.Attachments))
	}
	att := payload.Attachments[0]
	if att.Name != "test.txt" {
		t.Errorf("attachment name = %q, want %q", att.Name, "test.txt")
	}
	// Base64 of "hello world" = "aGVsbG8gd29ybGQ="
	if att.Content != "aGVsbG8gd29ybGQ=" {
		t.Errorf("attachment content = %q, want base64 of 'hello world'", att.Content)
	}
}

func TestSESSender_SendsRawMIME(t *testing.T) {
	t.Parallel()

	// We cannot easily mock the AWS SDK client in a unit test, so we verify
	// that the MIME builder integration works by checking buildMIME output.
	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "Hello from SES",
	}

	mimeData, err := buildMIME(msg)
	if err != nil {
		t.Fatalf("buildMIME() error = %v", err)
	}

	// Verify MIME output contains expected components.
	content := string(mimeData)
	if !strings.Contains(content, "From:") {
		t.Error("MIME missing From header")
	}
	if !strings.Contains(content, "Hello from SES") {
		t.Error("MIME missing body text")
	}
}

func TestSESSender_WithAttachments(t *testing.T) {
	t.Parallel()

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "See attached",
		Attachments: []Attachment{
			{Filename: "doc.pdf", ContentType: "application/pdf", Data: []byte("fake-pdf-data")},
		},
	}

	mimeData, err := buildMIME(msg)
	if err != nil {
		t.Fatalf("buildMIME() error = %v", err)
	}

	content := string(mimeData)
	if !strings.Contains(content, "multipart/mixed") {
		t.Error("expected multipart/mixed for message with attachments")
	}
	if !strings.Contains(content, "doc.pdf") {
		t.Error("expected attachment filename in MIME output")
	}
}

func TestSESSender_ProviderErrorWrapping(t *testing.T) {
	t.Parallel()

	// Verify wrapAWSError produces correct ProviderError structure.
	innerErr := errors.New("some AWS error")
	provErr := wrapAWSError(innerErr)

	if provErr.Provider != "ses" {
		t.Errorf("Provider = %q, want %q", provErr.Provider, "ses")
	}
	if provErr.Code != 0 {
		t.Errorf("Code = %d, want 0 (non-request error)", provErr.Code)
	}
	if !errors.Is(provErr, innerErr) {
		t.Error("expected Unwrap() to return the inner error")
	}
}

func TestSESSender_DefaultCredentialChain(t *testing.T) {
	t.Parallel()

	// Both AWS key fields empty — should succeed, falling back to default chain.
	sender, err := newSESSender(defaultSESConfig())
	if err != nil {
		t.Fatalf("newSESSender() error = %v", err)
	}
	if sender == nil {
		t.Fatal("expected non-nil sender")
	}
	if sender.region != "us-east-1" {
		t.Errorf("region = %q, want %q", sender.region, "us-east-1")
	}
}

func TestDetectSESRegion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{name: "empty", endpoint: "", want: "us-east-1"},
		{name: "amazonaws standard", endpoint: "https://email.us-west-2.amazonaws.com", want: "us-west-2"},
		{name: "amazonaws no scheme", endpoint: "email.eu-west-1.amazonaws.com", want: "eu-west-1"},
		{name: "custom endpoint", endpoint: "https://localstack:4566", want: "us-east-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := detectSESRegion(tt.endpoint)
			if got != tt.want {
				t.Errorf("detectSESRegion(%q) = %q, want %q", tt.endpoint, got, tt.want)
			}
		})
	}
}

func defaultSESConfig() config.Config {
	return config.Config{
		Email_Enabled:               true,
		Email_Provider:              config.EmailSES,
		Email_AWS_Access_Key_ID:     "",
		Email_AWS_Secret_Access_Key: "",
	}
}
