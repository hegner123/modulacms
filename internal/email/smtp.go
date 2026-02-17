// Package email provides transactional email sending with multiple provider backends.
//
// SMTP sender: uses PLAIN auth over TLS only. LOGIN and XOAUTH2 are not supported.
// smtp.PlainAuth refuses to authenticate over non-TLS connections (Go stdlib safety check).
//
// Compatible: Amazon SES SMTP, Mailgun SMTP, Postfix, Sendmail, MailPit/Mailhog (no auth),
// any relay that accepts PLAIN over STARTTLS or implicit TLS.
//
// Incompatible: Microsoft 365 (deprecated PLAIN, requires XOAUTH2), Google Workspace
// (deprecated PLAIN, requires XOAUTH2 or app passwords).
// Operators using M365 or Google Workspace must use an HTTP API provider (SendGrid, SES,
// Postmark) instead of SMTP.
package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/hegner123/modulacms/internal/config"
)

// SMTPSender sends email via SMTP with PLAIN authentication.
type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	useTLS   bool
}

// newSMTPSender creates an SMTPSender from config. Validates that Email_Host
// is non-empty. Defaults port to 587 (STARTTLS) or 465 (TLS).
func newSMTPSender(cfg config.Config) (*SMTPSender, error) {
	if cfg.Email_Host == "" {
		return nil, fmt.Errorf("email_host is required for SMTP provider")
	}

	port := cfg.Email_Port
	if port == 0 {
		if cfg.Email_TLS {
			port = 465
		} else {
			port = 587
		}
	}

	return &SMTPSender{
		host:     cfg.Email_Host,
		port:     port,
		username: cfg.Email_Username,
		password: cfg.Email_Password,
		useTLS:   cfg.Email_TLS,
	}, nil
}

// Send validates the message, builds MIME bytes, and delivers via SMTP.
// Uses context-aware dialing and per-connection deadlines. Does NOT use
// smtp.SendMail because it ignores context.Context and has no timeout support.
func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	// Defense-in-depth: validate independently so SMTPSender is safe without Service wrapper.
	if err := msg.Validate(); err != nil {
		return err
	}

	mimeData, err := buildMIME(msg)
	if err != nil {
		return &ProviderError{Provider: "smtp", Err: fmt.Errorf("build MIME: %w", err)}
	}

	// Compute timeout: 30s without attachments, 120s with attachments.
	timeout := 30 * time.Second
	if len(msg.Attachments) > 0 {
		timeout = 120 * time.Second
	}

	addr := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))

	// Dial with context awareness.
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return &ProviderError{Provider: "smtp", Err: fmt.Errorf("dial: %w", err)}
	}

	// Set deadline for the entire SMTP conversation. Use the earlier of the
	// computed timeout or the context deadline so cancellation is respected.
	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return &ProviderError{Provider: "smtp", Err: fmt.Errorf("set deadline: %w", err)}
	}

	tlsConfig := &tls.Config{ServerName: s.host}

	var client *smtp.Client

	if s.useTLS {
		// Implicit TLS (port 465): wrap the connection in TLS before SMTP.
		tlsConn := tls.Client(conn, tlsConfig)
		client, err = smtp.NewClient(tlsConn, s.host)
	} else {
		// STARTTLS (port 587): start plain, upgrade to TLS.
		client, err = smtp.NewClient(conn, s.host)
	}
	if err != nil {
		conn.Close()
		return &ProviderError{Provider: "smtp", Err: fmt.Errorf("new client: %w", err)}
	}
	defer client.Close()

	// STARTTLS upgrade for non-implicit TLS connections.
	if !s.useTLS {
		if err := client.StartTLS(tlsConfig); err != nil {
			return &ProviderError{Provider: "smtp", Err: fmt.Errorf("STARTTLS: %w", err)}
		}
	}

	// Authenticate if credentials are provided.
	if s.username != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return &ProviderError{Provider: "smtp", Err: fmt.Errorf("auth: %w", err)}
		}
	}

	// Set sender.
	if err := client.Mail(msg.From.Address); err != nil {
		return &ProviderError{Provider: "smtp", Err: fmt.Errorf("MAIL FROM: %w", err)}
	}

	// Set recipients.
	for _, rcpt := range msg.allRecipients() {
		if err := client.Rcpt(rcpt); err != nil {
			return &ProviderError{Provider: "smtp", Err: fmt.Errorf("RCPT TO %s: %w", rcpt, err)}
		}
	}

	// Write DATA.
	w, err := client.Data()
	if err != nil {
		return &ProviderError{Provider: "smtp", Err: fmt.Errorf("DATA: %w", err)}
	}
	if _, err := w.Write(mimeData); err != nil {
		return &ProviderError{Provider: "smtp", Err: fmt.Errorf("write data: %w", err)}
	}
	if err := w.Close(); err != nil {
		return &ProviderError{Provider: "smtp", Err: fmt.Errorf("close data: %w", err)}
	}

	return client.Quit()
}

// Close returns nil. SMTPSender dials fresh on each Send.
func (s *SMTPSender) Close() error { return nil }
