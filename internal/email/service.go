package email

import (
	"context"
	"sync"

	"github.com/hegner123/modulacms/internal/config"
)

// Service is a hot-reloadable email sending service. It holds a Sender,
// default From/ReplyTo addresses, and swaps them atomically on config reload.
type Service struct {
	mu      sync.RWMutex
	sender  Sender
	from    Address
	replyTo *Address
}

// NewService constructs a Service from the given config. Returns an error
// if the config specifies an enabled provider but the from address is invalid.
func NewService(cfg config.Config) (*Service, error) {
	sender, err := NewSender(cfg)
	if err != nil {
		return nil, err
	}

	from := NewAddress(cfg.Email_From_Name, cfg.Email_From_Address)

	// Validate the from address only if email is enabled (noop sender
	// doesn't need a valid from address).
	if cfg.Email_Enabled && cfg.Email_Provider != config.EmailDisabled {
		if err := from.Validate(); err != nil {
			return nil, err
		}
	}

	var replyTo *Address
	if cfg.Email_Reply_To != "" {
		rt := NewAddress("", cfg.Email_Reply_To)
		if err := rt.Validate(); err != nil {
			return nil, err
		}
		replyTo = &rt
	}

	return &Service{
		sender:  sender,
		from:    from,
		replyTo: replyTo,
	}, nil
}

// Send sends a message, injecting default From and ReplyTo if not set on msg.
// Validates the message before delegating to the underlying Sender.
func (s *Service) Send(ctx context.Context, msg Message) error {
	s.mu.RLock()
	sender := s.sender
	from := s.from
	replyTo := s.replyTo
	s.mu.RUnlock()

	// Inject defaults.
	if msg.From.Address == "" {
		msg.From = from
	}
	if msg.ReplyTo == nil && replyTo != nil {
		msg.ReplyTo = replyTo
	}

	if err := msg.Validate(); err != nil {
		return err
	}

	return sender.Send(ctx, msg)
}

// Reload builds a new sender from the updated config and swaps it in.
// In-flight Send calls that already acquired RLock will complete using the
// old sender. Close() on the old sender only releases idle connections,
// not active ones -- in-flight HTTP requests continue to completion.
// There is no automatic retry mechanism in v1; if an in-flight send fails
// during reload, the calling handler receives the error.
func (s *Service) Reload(cfg config.Config) error {
	newSender, err := NewSender(cfg)
	if err != nil {
		return err
	}

	from := NewAddress(cfg.Email_From_Name, cfg.Email_From_Address)
	var replyTo *Address
	if cfg.Email_Reply_To != "" {
		rt := NewAddress("", cfg.Email_Reply_To)
		replyTo = &rt
	}

	s.mu.Lock()
	oldSender := s.sender
	s.sender = newSender
	s.from = from
	s.replyTo = replyTo
	s.mu.Unlock()

	// Release idle connections from the old sender.
	return oldSender.Close()
}

// Close releases resources held by the current sender.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sender.Close()
}

// Enabled reports whether the service has an active (non-noop) sender.
func (s *Service) Enabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, isNoop := s.sender.(*noopSender)
	return !isNoop
}
