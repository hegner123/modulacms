package email

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/config"
)

// Sender is the interface for email transport backends.
type Sender interface {
	Send(ctx context.Context, msg Message) error
	Close() error
}

// noopSender is a no-op sender that returns DisabledError on every Send.
type noopSender struct{}

func (n *noopSender) Send(_ context.Context, _ Message) error {
	return &DisabledError{}
}

func (n *noopSender) Close() error { return nil }

// NewSender constructs a Sender based on the email provider in the config.
func NewSender(cfg config.Config) (Sender, error) {
	if !cfg.Email_Enabled || cfg.Email_Provider == config.EmailDisabled {
		return &noopSender{}, nil
	}

	switch cfg.Email_Provider {
	case config.EmailSmtp:
		return newSMTPSender(cfg)
	case config.EmailSendGrid:
		return newSendGridSender(cfg)
	case config.EmailSES:
		return newSESSender(cfg)
	case config.EmailPostmark:
		return newPostmarkSender(cfg)
	default:
		return nil, fmt.Errorf("unknown email provider: %q", cfg.Email_Provider)
	}
}
