package email

import (
	"net/mail"
	"strings"
)

// MaxMessageSize is the maximum total size of attachment data in bytes.
// Set to 7.5 MB because base64 encoding inflates data by ~33%, and
// Postmark's 10 MB limit applies to the encoded API payload.
const MaxMessageSize = 7.5 * 1024 * 1024 // 7,864,320 bytes

// MaxRecipients is the maximum number of recipients (To + CC + BCC).
// SES limits to 50 recipients per call; SMTP servers typically reject
// beyond 100 RCPT TO commands. 50 is the safe lowest common denominator.
const MaxRecipients = 50

// Address wraps net/mail.Address with validation.
type Address struct {
	Name    string
	Address string
}

// NewAddress creates an Address with the given display name and email.
func NewAddress(name, email string) Address {
	return Address{Name: name, Address: email}
}

// Validate checks that the Address is a valid RFC 5322 email address.
// Checks are explicit string operations, not regex.
func (a Address) Validate() error {
	if a.Address == "" {
		return &InvalidMessageError{Field: "Address", Problem: "email address is empty"}
	}

	atIdx := strings.IndexByte(a.Address, '@')
	if atIdx < 0 {
		return &InvalidMessageError{Field: "Address", Problem: "email address missing @"}
	}

	local := a.Address[:atIdx]
	domain := a.Address[atIdx+1:]

	if local == "" {
		return &InvalidMessageError{Field: "Address", Problem: "email address has empty local part"}
	}
	if domain == "" {
		return &InvalidMessageError{Field: "Address", Problem: "email address has empty domain part"}
	}

	// Validate RFC 5322 syntax via stdlib.
	formatted := a.mailAddress().String()
	if _, err := mail.ParseAddress(formatted); err != nil {
		return &InvalidMessageError{Field: "Address", Problem: "invalid RFC 5322 address: " + err.Error()}
	}

	return nil
}

// mailAddress converts to a net/mail.Address.
func (a Address) mailAddress() *mail.Address {
	return &mail.Address{Name: a.Name, Address: a.Address}
}

// String returns the RFC 5322 formatted address.
func (a Address) String() string {
	return a.mailAddress().String()
}

// Attachment represents a file attachment on an email message.
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// Validate checks that the Attachment has required fields.
func (a Attachment) Validate() error {
	if a.Filename == "" {
		return &InvalidMessageError{Field: "Attachment", Problem: "filename is required"}
	}
	if a.ContentType == "" {
		return &InvalidMessageError{Field: "Attachment", Problem: "content-type is required"}
	}
	if len(a.Data) == 0 {
		return &InvalidMessageError{Field: "Attachment", Problem: "data is empty"}
	}
	return nil
}

// Message is a complete email envelope ready for sending.
type Message struct {
	From        Address
	To          []Address
	CC          []Address
	BCC         []Address
	ReplyTo     *Address
	Subject     string
	PlainBody   string
	HTMLBody    string
	Attachments []Attachment
	Headers     map[string]string
}

// Validate checks all message fields before any send attempt.
func (m Message) Validate() error {
	if err := m.From.Validate(); err != nil {
		return &InvalidMessageError{Field: "From", Problem: err.Error()}
	}

	if len(m.To) == 0 {
		return &InvalidMessageError{Field: "To", Problem: "at least one recipient is required"}
	}
	for i, addr := range m.To {
		if err := addr.Validate(); err != nil {
			return &InvalidMessageError{Field: "To", Problem: "recipient " + itoa(i) + ": " + err.Error()}
		}
	}

	for i, addr := range m.CC {
		if err := addr.Validate(); err != nil {
			return &InvalidMessageError{Field: "CC", Problem: "recipient " + itoa(i) + ": " + err.Error()}
		}
	}

	for i, addr := range m.BCC {
		if err := addr.Validate(); err != nil {
			return &InvalidMessageError{Field: "BCC", Problem: "recipient " + itoa(i) + ": " + err.Error()}
		}
	}

	if m.ReplyTo != nil {
		if err := m.ReplyTo.Validate(); err != nil {
			return &InvalidMessageError{Field: "ReplyTo", Problem: err.Error()}
		}
	}

	if strings.TrimSpace(m.Subject) == "" {
		return &InvalidMessageError{Field: "Subject", Problem: "subject is required"}
	}

	if strings.TrimSpace(m.PlainBody) == "" && strings.TrimSpace(m.HTMLBody) == "" {
		return &InvalidMessageError{Field: "Body", Problem: "at least one of PlainBody or HTMLBody is required"}
	}

	// Validate attachments and check total size.
	var totalSize int64
	for _, att := range m.Attachments {
		if err := att.Validate(); err != nil {
			return err
		}
		totalSize += int64(len(att.Data))
	}
	if totalSize > int64(MaxMessageSize) {
		return &InvalidMessageError{Field: "Attachments", Problem: "total attachment size exceeds 7.5 MB limit"}
	}

	// Check total recipient count.
	recipientCount := len(m.To) + len(m.CC) + len(m.BCC)
	if recipientCount > MaxRecipients {
		return &InvalidMessageError{Field: "Recipients", Problem: "total recipients exceed 50 limit"}
	}

	// Validate custom headers.
	for key, value := range m.Headers {
		if err := validateHeaderKey(key); err != nil {
			return err
		}
		if err := validateHeaderValue(value); err != nil {
			return err
		}
	}

	return nil
}

// allRecipients flattens To+CC+BCC into a single slice of email address strings
// for use as SMTP RCPT TO recipients.
func (m Message) allRecipients() []string {
	total := len(m.To) + len(m.CC) + len(m.BCC)
	recipients := make([]string, 0, total)
	for _, addr := range m.To {
		recipients = append(recipients, addr.Address)
	}
	for _, addr := range m.CC {
		recipients = append(recipients, addr.Address)
	}
	for _, addr := range m.BCC {
		recipients = append(recipients, addr.Address)
	}
	return recipients
}

// validateHeaderKey checks that a header key contains only printable ASCII
// (0x21-0x7E) excluding colon, per RFC 5322 Section 2.2.
func validateHeaderKey(key string) error {
	for i := 0; i < len(key); i++ {
		b := key[i]
		if b == ':' || b == ' ' || b == 0 || b == '\r' || b == '\n' || b < 0x21 || b > 0x7E {
			return &InvalidMessageError{Field: "Headers", Problem: "header key contains invalid character"}
		}
	}
	return nil
}

// validateHeaderValue checks that a header value does not contain CR or LF
// characters, which would enable header injection attacks.
func validateHeaderValue(value string) error {
	for i := 0; i < len(value); i++ {
		b := value[i]
		if b == '\r' || b == '\n' {
			return &InvalidMessageError{Field: "Headers", Problem: "header value contains CR or LF (header injection)"}
		}
	}
	return nil
}

// itoa converts an integer to a string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	negative := i < 0
	if negative {
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
