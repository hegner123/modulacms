package email

import (
	"errors"
	"net/mail"
	"strings"
	"testing"
	"time"
)

func TestAddress_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		addr    Address
		wantErr bool
	}{
		{name: "empty", addr: NewAddress("", ""), wantErr: true},
		{name: "missing @", addr: NewAddress("", "noatsign"), wantErr: true},
		{name: "empty local part", addr: NewAddress("", "@example.com"), wantErr: true},
		{name: "empty domain", addr: NewAddress("", "user@"), wantErr: true},
		{name: "valid bare", addr: NewAddress("", "user@example.com"), wantErr: false},
		{name: "valid with name", addr: NewAddress("Test User", "user@example.com"), wantErr: false},
		{name: "valid dotless domain (localhost)", addr: NewAddress("", "user@localhost"), wantErr: false},
		{name: "invalid format", addr: NewAddress("", "not an email <><>"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.addr.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Address.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAttachment_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		att     Attachment
		wantErr bool
	}{
		{name: "missing filename", att: Attachment{ContentType: "text/plain", Data: []byte("x")}, wantErr: true},
		{name: "missing content-type", att: Attachment{Filename: "a.txt", Data: []byte("x")}, wantErr: true},
		{name: "empty data", att: Attachment{Filename: "a.txt", ContentType: "text/plain"}, wantErr: true},
		{name: "valid", att: Attachment{Filename: "a.txt", ContentType: "text/plain", Data: []byte("hello")}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.att.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Attachment.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessage_Validate(t *testing.T) {
	t.Parallel()

	validFrom := NewAddress("Sender", "sender@example.com")
	validTo := []Address{NewAddress("Recip", "recip@example.com")}

	tests := []struct {
		name    string
		msg     Message
		wantErr bool
		errField string
	}{
		{
			name:    "missing from",
			msg:     Message{To: validTo, Subject: "Hi", PlainBody: "body"},
			wantErr: true,
			errField: "From",
		},
		{
			name:    "empty to",
			msg:     Message{From: validFrom, Subject: "Hi", PlainBody: "body"},
			wantErr: true,
			errField: "To",
		},
		{
			name:    "invalid to addr",
			msg:     Message{From: validFrom, To: []Address{NewAddress("", "bad")}, Subject: "Hi", PlainBody: "body"},
			wantErr: true,
			errField: "To",
		},
		{
			name:    "empty subject",
			msg:     Message{From: validFrom, To: validTo, Subject: "", PlainBody: "body"},
			wantErr: true,
			errField: "Subject",
		},
		{
			name:    "no body",
			msg:     Message{From: validFrom, To: validTo, Subject: "Hi"},
			wantErr: true,
			errField: "Body",
		},
		{
			name:    "both bodies valid",
			msg:     Message{From: validFrom, To: validTo, Subject: "Hi", PlainBody: "text", HTMLBody: "<p>html</p>"},
			wantErr: false,
		},
		{
			name:    "plain only valid",
			msg:     Message{From: validFrom, To: validTo, Subject: "Hi", PlainBody: "text"},
			wantErr: false,
		},
		{
			name:    "HTML only valid",
			msg:     Message{From: validFrom, To: validTo, Subject: "Hi", HTMLBody: "<p>html</p>"},
			wantErr: false,
		},
		{
			name: "invalid attachment",
			msg: Message{
				From: validFrom, To: validTo, Subject: "Hi", PlainBody: "body",
				Attachments: []Attachment{{Filename: "", ContentType: "text/plain", Data: []byte("x")}},
			},
			wantErr: true,
			errField: "Attachment",
		},
		{
			name: "attachments exceed 7.5MB",
			msg: Message{
				From: validFrom, To: validTo, Subject: "Hi", PlainBody: "body",
				Attachments: []Attachment{
					{Filename: "big.bin", ContentType: "application/octet-stream", Data: make([]byte, 8*1024*1024)},
				},
			},
			wantErr: true,
			errField: "Attachments",
		},
		{
			name: "recipients exceed 50",
			msg: func() Message {
				to := make([]Address, 51)
				for i := range to {
					to[i] = NewAddress("", "user"+itoa(i)+"@example.com")
				}
				return Message{From: validFrom, To: to, Subject: "Hi", PlainBody: "body"}
			}(),
			wantErr: true,
			errField: "Recipients",
		},
		{
			name: "header key with colon",
			msg: Message{
				From: validFrom, To: validTo, Subject: "Hi", PlainBody: "body",
				Headers: map[string]string{"X-Bad:Key": "value"},
			},
			wantErr: true,
			errField: "Headers",
		},
		{
			name: "header key with null byte",
			msg: Message{
				From: validFrom, To: validTo, Subject: "Hi", PlainBody: "body",
				Headers: map[string]string{"X-Bad\x00Key": "value"},
			},
			wantErr: true,
			errField: "Headers",
		},
		{
			name: "header value with CRLF",
			msg: Message{
				From: validFrom, To: validTo, Subject: "Hi", PlainBody: "body",
				Headers: map[string]string{"X-Good": "value\r\nInjected: header"},
			},
			wantErr: true,
			errField: "Headers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.msg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errField != "" {
				var invErr *InvalidMessageError
				if errors.As(err, &invErr) {
					if invErr.Field != tt.errField {
						t.Errorf("expected error field %q, got %q", tt.errField, invErr.Field)
					}
				}
			}
		})
	}
}

func TestMessage_AllRecipients(t *testing.T) {
	t.Parallel()

	msg := Message{
		To:  []Address{NewAddress("", "to1@example.com"), NewAddress("", "to2@example.com")},
		CC:  []Address{NewAddress("", "cc@example.com")},
		BCC: []Address{NewAddress("", "bcc@example.com")},
	}

	recipients := msg.allRecipients()
	if len(recipients) != 4 {
		t.Fatalf("expected 4 recipients, got %d", len(recipients))
	}
	expected := []string{"to1@example.com", "to2@example.com", "cc@example.com", "bcc@example.com"}
	for i, want := range expected {
		if recipients[i] != want {
			t.Errorf("recipient[%d] = %q, want %q", i, recipients[i], want)
		}
	}
}

func TestBuildMIME_PlainOnly(t *testing.T) {
	t.Parallel()

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "Hello, world!",
	}

	data, err := buildMIME(msg)
	if err != nil {
		t.Fatalf("buildMIME() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Content-Type: text/plain") {
		t.Error("expected Content-Type: text/plain header")
	}
	if !strings.Contains(content, "Hello, world!") {
		t.Error("expected plain body in output")
	}
}

func TestBuildMIME_HTMLOnly(t *testing.T) {
	t.Parallel()

	msg := Message{
		From:     NewAddress("Sender", "sender@example.com"),
		To:       []Address{NewAddress("Recip", "recip@example.com")},
		Subject:  "Test",
		HTMLBody: "<p>Hello</p>",
	}

	data, err := buildMIME(msg)
	if err != nil {
		t.Fatalf("buildMIME() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Content-Type: text/html") {
		t.Error("expected Content-Type: text/html header")
	}
}

func TestBuildMIME_MultipartAlternative(t *testing.T) {
	t.Parallel()

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "Plain text",
		HTMLBody:  "<p>HTML</p>",
	}

	data, err := buildMIME(msg)
	if err != nil {
		t.Fatalf("buildMIME() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "multipart/alternative") {
		t.Error("expected multipart/alternative content type")
	}
	if !strings.Contains(content, "Plain text") {
		t.Error("expected plain body in output")
	}
	if !strings.Contains(content, "<p>HTML</p>") {
		t.Error("expected HTML body in output")
	}
}

func TestBuildMIME_WithAttachments(t *testing.T) {
	t.Parallel()

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "See attached",
		Attachments: []Attachment{
			{Filename: "test.txt", ContentType: "text/plain", Data: []byte("attachment content")},
		},
	}

	data, err := buildMIME(msg)
	if err != nil {
		t.Fatalf("buildMIME() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "multipart/mixed") {
		t.Error("expected multipart/mixed content type")
	}
	if !strings.Contains(content, "Content-Transfer-Encoding: base64") {
		t.Error("expected base64 attachment encoding")
	}
	if !strings.Contains(content, "test.txt") {
		t.Error("expected attachment filename in disposition")
	}
}

func TestBuildMIME_SubjectEncoding(t *testing.T) {
	t.Parallel()

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Héllo Wörld",
		PlainBody: "body",
	}

	data, err := buildMIME(msg)
	if err != nil {
		t.Fatalf("buildMIME() error = %v", err)
	}

	content := string(data)
	// Non-ASCII subject should be Q-encoded.
	if !strings.Contains(content, "=?utf-8?q?") {
		t.Error("expected Q-encoded subject for non-ASCII characters")
	}
}

func TestBuildMIME_RequiredHeaders(t *testing.T) {
	t.Parallel()

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "body",
	}

	data, err := buildMIME(msg)
	if err != nil {
		t.Fatalf("buildMIME() error = %v", err)
	}

	content := string(data)

	// MIME-Version must be present.
	if !strings.Contains(content, "MIME-Version: 1.0") {
		t.Error("expected MIME-Version: 1.0 header")
	}

	// Date must be present and parseable.
	dateIdx := strings.Index(content, "Date: ")
	if dateIdx < 0 {
		t.Fatal("expected Date header")
	}
	dateStart := dateIdx + len("Date: ")
	dateEnd := strings.Index(content[dateStart:], "\r\n")
	if dateEnd < 0 {
		t.Fatal("could not find end of Date header")
	}
	dateStr := content[dateStart : dateStart+dateEnd]
	if _, err := time.Parse(time.RFC1123Z, dateStr); err != nil {
		t.Errorf("Date header %q does not parse as RFC1123Z: %v", dateStr, err)
	}

	// Message-ID must be present and match <hex@domain> format.
	msgIDIdx := strings.Index(content, "Message-ID: ")
	if msgIDIdx < 0 {
		t.Fatal("expected Message-ID header")
	}
	msgIDStart := msgIDIdx + len("Message-ID: ")
	msgIDEnd := strings.Index(content[msgIDStart:], "\r\n")
	if msgIDEnd < 0 {
		t.Fatal("could not find end of Message-ID header")
	}
	msgID := content[msgIDStart : msgIDStart+msgIDEnd]
	if !strings.HasPrefix(msgID, "<") || !strings.HasSuffix(msgID, ">") {
		t.Errorf("Message-ID %q not in <...> format", msgID)
	}
	inner := msgID[1 : len(msgID)-1]
	atIdx := strings.IndexByte(inner, '@')
	if atIdx < 0 {
		t.Errorf("Message-ID %q missing @", msgID)
	} else {
		hexPart := inner[:atIdx]
		if len(hexPart) != 32 {
			t.Errorf("Message-ID hex part length = %d, want 32", len(hexPart))
		}
		domain := inner[atIdx+1:]
		if domain != "example.com" {
			t.Errorf("Message-ID domain = %q, want %q", domain, "example.com")
		}
	}
}

func TestBuildMIME_PreservesCustomHeaders(t *testing.T) {
	t.Parallel()

	customDate := "Mon, 01 Jan 2024 00:00:00 +0000"
	customMsgID := "<custom-id@example.com>"

	msg := Message{
		From:      NewAddress("Sender", "sender@example.com"),
		To:        []Address{NewAddress("Recip", "recip@example.com")},
		Subject:   "Test",
		PlainBody: "body",
		Headers: map[string]string{
			"Date":       customDate,
			"Message-ID": customMsgID,
		},
	}

	data, err := buildMIME(msg)
	if err != nil {
		t.Fatalf("buildMIME() error = %v", err)
	}

	content := string(data)

	// Parse with net/mail to verify headers are properly set.
	parsed, parseErr := mail.ReadMessage(strings.NewReader(content))
	if parseErr != nil {
		t.Fatalf("mail.ReadMessage() error = %v", parseErr)
	}

	if got := parsed.Header.Get("Date"); got != customDate {
		t.Errorf("Date = %q, want %q", got, customDate)
	}
	if got := parsed.Header.Get("Message-Id"); got != customMsgID {
		t.Errorf("Message-ID = %q, want %q", got, customMsgID)
	}
}
