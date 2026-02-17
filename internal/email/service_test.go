package email

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/config"
)

func TestNewService_DisabledConfig(t *testing.T) {
	t.Parallel()

	cfg := config.Config{Email_Enabled: false}
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if svc.Enabled() {
		t.Error("expected Enabled() = false for disabled config")
	}
}

func TestNewService_InvalidFromAddress(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Email_Enabled:      true,
		Email_Provider:     config.EmailSmtp,
		Email_Host:         "smtp.example.com",
		Email_From_Address: "not-an-email",
	}
	_, err := NewService(cfg)
	if err == nil {
		t.Fatal("expected error for invalid from address")
	}
}

func TestService_Send_InjectsFrom(t *testing.T) {
	t.Parallel()

	mock := &mockSender{}
	svc := &Service{
		sender: mock,
		from:   NewAddress("Default", "default@example.com"),
	}

	msg := Message{
		To:        []Address{NewAddress("", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "body",
	}
	err := svc.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if mock.lastMsg.From.Address != "default@example.com" {
		t.Errorf("From = %q, want %q", mock.lastMsg.From.Address, "default@example.com")
	}
}

func TestService_Send_InjectsReplyTo(t *testing.T) {
	t.Parallel()

	rt := NewAddress("", "reply@example.com")
	mock := &mockSender{}
	svc := &Service{
		sender:  mock,
		from:    NewAddress("Default", "default@example.com"),
		replyTo: &rt,
	}

	msg := Message{
		To:        []Address{NewAddress("", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "body",
	}
	err := svc.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if mock.lastMsg.ReplyTo == nil {
		t.Fatal("expected ReplyTo to be injected")
	}
	if mock.lastMsg.ReplyTo.Address != "reply@example.com" {
		t.Errorf("ReplyTo = %q, want %q", mock.lastMsg.ReplyTo.Address, "reply@example.com")
	}
}

func TestService_Send_ValidationFailure(t *testing.T) {
	t.Parallel()

	mock := &mockSender{}
	svc := &Service{
		sender: mock,
		from:   NewAddress("Default", "default@example.com"),
	}

	// Empty To should fail validation.
	msg := Message{
		Subject:   "Test",
		PlainBody: "body",
	}
	err := svc.Send(context.Background(), msg)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var invErr *InvalidMessageError
	if !errors.As(err, &invErr) {
		t.Errorf("expected *InvalidMessageError, got %T", err)
	}
}

func TestService_Reload_SwapsSender(t *testing.T) {
	t.Parallel()

	cfg := config.Config{Email_Enabled: false}
	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if svc.Enabled() {
		t.Fatal("expected initially disabled")
	}

	newCfg := config.Config{
		Email_Enabled:      true,
		Email_Provider:     config.EmailSmtp,
		Email_Host:         "smtp.example.com",
		Email_From_Address: "from@example.com",
	}
	if err := svc.Reload(newCfg); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}
	if !svc.Enabled() {
		t.Error("expected Enabled() = true after reload")
	}
}

func TestService_Reload_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	mock := &mockSender{}
	svc := &Service{
		sender: mock,
		from:   NewAddress("Default", "default@example.com"),
	}

	var wg sync.WaitGroup
	const goroutines = 10

	// Concurrent sends.
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 10 {
				msg := Message{
					To:        []Address{NewAddress("", "recip@example.com")},
					Subject:   "Test",
					PlainBody: "body",
				}
				svc.Send(context.Background(), msg)
			}
		}()
	}

	// Concurrent reloads.
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := config.Config{Email_Enabled: false}
			svc.Reload(cfg)
		}()
	}

	wg.Wait()
}

func TestService_Reload_InFlightCompletes(t *testing.T) {
	t.Parallel()

	var sendStarted atomic.Bool
	var sendCompleted atomic.Bool

	slowSender := &slowMockSender{
		delay:     200 * time.Millisecond,
		onStart:   func() { sendStarted.Store(true) },
		onComplete: func() { sendCompleted.Store(true) },
	}

	svc := &Service{
		sender: slowSender,
		from:   NewAddress("Default", "default@example.com"),
	}

	// Start a slow send in a goroutine.
	var sendErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		msg := Message{
			To:        []Address{NewAddress("", "recip@example.com")},
			Subject:   "Test",
			PlainBody: "body",
		}
		sendErr = svc.Send(context.Background(), msg)
	}()

	// Wait for the send to start, then reload.
	for !sendStarted.Load() {
		time.Sleep(5 * time.Millisecond)
	}

	cfg := config.Config{Email_Enabled: false}
	if err := svc.Reload(cfg); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	wg.Wait()

	if sendErr != nil {
		t.Errorf("in-flight Send() error = %v", sendErr)
	}
	if !sendCompleted.Load() {
		t.Error("in-flight send did not complete")
	}
}

// mockSender records the last message sent.
type mockSender struct {
	mu      sync.Mutex
	lastMsg Message
}

func (m *mockSender) Send(_ context.Context, msg Message) error {
	m.mu.Lock()
	m.lastMsg = msg
	m.mu.Unlock()
	return nil
}

func (m *mockSender) Close() error { return nil }

// slowMockSender simulates a slow send for testing in-flight behavior.
type slowMockSender struct {
	delay      time.Duration
	onStart    func()
	onComplete func()
}

func (s *slowMockSender) Send(_ context.Context, _ Message) error {
	if s.onStart != nil {
		s.onStart()
	}
	time.Sleep(s.delay)
	if s.onComplete != nil {
		s.onComplete()
	}
	return nil
}

func (s *slowMockSender) Close() error { return nil }
