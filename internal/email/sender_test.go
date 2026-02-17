package email

import (
	"context"
	"errors"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

func TestNewSender_Disabled(t *testing.T) {
	t.Parallel()

	cfg := config.Config{Email_Enabled: false}
	sender, err := NewSender(cfg)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}
	if _, ok := sender.(*noopSender); !ok {
		t.Error("expected noopSender when Email_Enabled=false")
	}
}

func TestNewSender_DisabledProvider(t *testing.T) {
	t.Parallel()

	cfg := config.Config{Email_Enabled: true, Email_Provider: ""}
	sender, err := NewSender(cfg)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}
	if _, ok := sender.(*noopSender); !ok {
		t.Error("expected noopSender when provider is empty")
	}
}

func TestNewSender_UnknownProvider(t *testing.T) {
	t.Parallel()

	cfg := config.Config{Email_Enabled: true, Email_Provider: "carrier-pigeon"}
	_, err := NewSender(cfg)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestNoopSender_ReturnsDisabledError(t *testing.T) {
	t.Parallel()

	sender := &noopSender{}
	err := sender.Send(context.Background(), Message{})
	if err == nil {
		t.Fatal("expected error from noopSender")
	}
	var disabledErr *DisabledError
	if !errors.As(err, &disabledErr) {
		t.Errorf("expected *DisabledError, got %T", err)
	}
}

func TestNewSender_SMTPMissingHost(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Email_Enabled:  true,
		Email_Provider: config.EmailSmtp,
		Email_Host:     "",
	}
	_, err := NewSender(cfg)
	if err == nil {
		t.Fatal("expected error for SMTP with missing host")
	}
}

func TestNewSender_SendGridMissingKey(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Email_Enabled:  true,
		Email_Provider: config.EmailSendGrid,
		Email_API_Key:  "",
	}
	_, err := NewSender(cfg)
	if err == nil {
		t.Fatal("expected error for SendGrid with missing API key")
	}
}

func TestNewSender_PostmarkMissingKey(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Email_Enabled:  true,
		Email_Provider: config.EmailPostmark,
		Email_API_Key:  "",
	}
	_, err := NewSender(cfg)
	if err == nil {
		t.Fatal("expected error for Postmark with missing API key")
	}
}

func TestNewSender_SESDefaultCredentialChain(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Email_Enabled:               true,
		Email_Provider:              config.EmailSES,
		Email_AWS_Access_Key_ID:     "",
		Email_AWS_Secret_Access_Key: "",
	}
	sender, err := NewSender(cfg)
	if err != nil {
		t.Fatalf("NewSender() error = %v; expected success with default credential chain", err)
	}
	if _, ok := sender.(*SESSender); !ok {
		t.Errorf("expected *SESSender, got %T", sender)
	}
}
