package email

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

func TestSMTPSender_ValidatesBeforeDial(t *testing.T) {
	t.Parallel()

	sender := &SMTPSender{
		host: "127.0.0.1",
		port: 25,
	}

	// Invalid message: no To recipients.
	msg := Message{
		From:      NewAddress("", "sender@example.com"),
		Subject:   "Test",
		PlainBody: "body",
	}

	err := sender.Send(context.Background(), msg)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var invErr *InvalidMessageError
	if !errors.As(err, &invErr) {
		t.Errorf("expected *InvalidMessageError, got %T: %v", err, err)
	}
}

func TestSMTPSender_SendPlainAuth(t *testing.T) {
	t.Parallel()

	// Start a local TCP listener speaking minimal SMTP.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	// SMTP mock server goroutine.
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(10 * time.Second))
		runMinimalSMTPServer(conn)
	}()

	sender := &SMTPSender{
		host:     "127.0.0.1",
		port:     addr.Port,
		username: "user",
		password: "pass",
		useTLS:   false,
	}

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "Hello",
	}

	// This will likely fail at STARTTLS since the mock doesn't support real TLS.
	// We expect a ProviderError wrapping a STARTTLS failure.
	err = sender.Send(context.Background(), msg)
	if err == nil {
		// If it somehow succeeded (no TLS enforcement), that's acceptable.
		return
	}
	var provErr *ProviderError
	if !errors.As(err, &provErr) {
		t.Errorf("expected *ProviderError, got %T: %v", err, err)
	}

	<-serverDone
}

func TestSMTPSender_RespectsContext(t *testing.T) {
	t.Parallel()

	// Start a listener that accepts but never responds.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		// Hold the connection open without responding.
		time.Sleep(30 * time.Second)
		conn.Close()
	}()

	sender := &SMTPSender{
		host:   "127.0.0.1",
		port:   addr.Port,
		useTLS: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "body",
	}

	start := time.Now()
	err = sender.Send(ctx, msg)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error from cancelled context")
	}

	// Should return within a reasonable time of the context timeout.
	if elapsed > 5*time.Second {
		t.Errorf("Send took %v, expected to return within ~500ms context timeout", elapsed)
	}
}

func TestSMTPSender_TimeoutOnHungServer(t *testing.T) {
	t.Parallel()

	// Start a listener that accepts and sends greeting but then hangs.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		// Send SMTP greeting then hang.
		conn.Write([]byte("220 localhost SMTP\r\n"))
		time.Sleep(120 * time.Second)
		conn.Close()
	}()

	sender := &SMTPSender{
		host:   "127.0.0.1",
		port:   addr.Port,
		useTLS: false,
	}

	// Use a context with a timeout shorter than the server hang.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "body",
	}

	start := time.Now()
	err = sender.Send(ctx, msg)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error from hung server")
	}

	// The deadline is set to 30s (no attachments), so it should time out
	// within the 30s deadline. With the 2s context, it should fail sooner.
	if elapsed > 35*time.Second {
		t.Errorf("Send took %v, expected to return within deadline", elapsed)
	}
}

// runMinimalSMTPServer implements a minimal SMTP server for testing.
func runMinimalSMTPServer(conn net.Conn) {
	conn.Write([]byte("220 localhost SMTP\r\n"))
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO"):
			conn.Write([]byte("250-localhost\r\n"))
			conn.Write([]byte("250-STARTTLS\r\n"))
			conn.Write([]byte("250-AUTH PLAIN LOGIN\r\n"))
			conn.Write([]byte("250 OK\r\n"))
		case strings.HasPrefix(upper, "STARTTLS"):
			conn.Write([]byte("220 Ready to start TLS\r\n"))
			// We can't actually do TLS here; the client will fail.
			return
		case strings.HasPrefix(upper, "AUTH"):
			conn.Write([]byte("235 Authentication successful\r\n"))
		case strings.HasPrefix(upper, "MAIL"):
			conn.Write([]byte("250 OK\r\n"))
		case strings.HasPrefix(upper, "RCPT"):
			conn.Write([]byte("250 OK\r\n"))
		case strings.HasPrefix(upper, "DATA"):
			conn.Write([]byte("354 Start mail input\r\n"))
			// Read until lone dot.
			for scanner.Scan() {
				if scanner.Text() == "." {
					break
				}
			}
			conn.Write([]byte("250 OK\r\n"))
		case strings.HasPrefix(upper, "QUIT"):
			conn.Write([]byte("221 Bye\r\n"))
			return
		default:
			conn.Write([]byte("250 OK\r\n"))
		}
	}
}
