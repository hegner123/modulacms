package email

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
	"time"
)

// buildMIME builds RFC 2045 MIME bytes from a Message.
// Uses stdlib only -- no internal package imports.
func buildMIME(msg Message) ([]byte, error) {
	var buf bytes.Buffer

	// Write standard headers.
	writeHeader(&buf, "From", msg.From.String())
	writeHeader(&buf, "To", joinAddresses(msg.To))
	if len(msg.CC) > 0 {
		writeHeader(&buf, "Cc", joinAddresses(msg.CC))
	}
	if msg.ReplyTo != nil {
		writeHeader(&buf, "Reply-To", msg.ReplyTo.String())
	}
	writeHeader(&buf, "Subject", encodeSubject(msg.Subject))

	// Write required headers if not already set by the caller.
	headers := msg.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	if _, ok := headers["MIME-Version"]; !ok {
		writeHeader(&buf, "MIME-Version", "1.0")
	}
	if _, ok := headers["Date"]; !ok {
		writeHeader(&buf, "Date", time.Now().UTC().Format(time.RFC1123Z))
	}
	if _, ok := headers["Message-ID"]; !ok {
		msgID, err := generateMessageID(msg.From.Address)
		if err != nil {
			return nil, fmt.Errorf("generate message-id: %w", err)
		}
		writeHeader(&buf, "Message-ID", msgID)
	}

	// Write custom headers.
	for key, value := range headers {
		writeHeader(&buf, key, value)
	}

	// Determine body structure.
	hasPlain := strings.TrimSpace(msg.PlainBody) != ""
	hasHTML := strings.TrimSpace(msg.HTMLBody) != ""
	hasAttachments := len(msg.Attachments) > 0

	switch {
	case hasAttachments:
		writeMixedBody(&buf, msg, hasPlain, hasHTML)
	case hasPlain && hasHTML:
		writeAlternativeBody(&buf, msg)
	case hasHTML:
		writeHeader(&buf, "Content-Type", "text/html; charset=utf-8")
		writeHeader(&buf, "Content-Transfer-Encoding", "quoted-printable")
		buf.WriteString("\r\n")
		buf.WriteString(msg.HTMLBody)
	default:
		writeHeader(&buf, "Content-Type", "text/plain; charset=utf-8")
		writeHeader(&buf, "Content-Transfer-Encoding", "quoted-printable")
		buf.WriteString("\r\n")
		buf.WriteString(msg.PlainBody)
	}

	return buf.Bytes(), nil
}

// joinAddresses formats a slice of Addresses as comma-separated RFC 5322 strings.
func joinAddresses(addrs []Address) string {
	parts := make([]string, len(addrs))
	for i, a := range addrs {
		parts[i] = a.String()
	}
	return strings.Join(parts, ", ")
}

// writeHeader writes a single MIME header line.
func writeHeader(buf *bytes.Buffer, key, value string) {
	buf.WriteString(key)
	buf.WriteString(": ")
	buf.WriteString(value)
	buf.WriteString("\r\n")
}

// generateMessageID creates a unique Message-ID in the format <hex@domain>.
// Uses 16 random bytes from crypto/rand, hex-encoded to 32 characters.
func generateMessageID(fromAddress string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		return "", err
	}

	domain := "localhost"
	atIdx := strings.IndexByte(fromAddress, '@')
	if atIdx >= 0 && atIdx+1 < len(fromAddress) {
		domain = fromAddress[atIdx+1:]
	}

	return "<" + hex.EncodeToString(randomBytes) + "@" + domain + ">", nil
}

// encodeSubject returns the subject as-is for ASCII-only strings, or
// Q-encoded (RFC 2047) for subjects containing non-ASCII characters.
func encodeSubject(subject string) string {
	for i := 0; i < len(subject); i++ {
		if subject[i] > 127 {
			return mime.QEncoding.Encode("utf-8", subject)
		}
	}
	return subject
}

// writeAlternativeBody writes a multipart/alternative body with plain + HTML parts.
func writeAlternativeBody(buf *bytes.Buffer, msg Message) {
	writer := multipart.NewWriter(buf)
	writeHeader(buf, "Content-Type", "multipart/alternative; boundary="+writer.Boundary())
	buf.WriteString("\r\n")

	plainHeader := make(textproto.MIMEHeader)
	plainHeader.Set("Content-Type", "text/plain; charset=utf-8")
	plainHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	part, err := writer.CreatePart(plainHeader)
	if err == nil {
		part.Write([]byte(msg.PlainBody))
	}

	htmlHeader := make(textproto.MIMEHeader)
	htmlHeader.Set("Content-Type", "text/html; charset=utf-8")
	htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	part, err = writer.CreatePart(htmlHeader)
	if err == nil {
		part.Write([]byte(msg.HTMLBody))
	}

	writer.Close()
}

// writeMixedBody writes a multipart/mixed body wrapping text/alternative + attachments.
func writeMixedBody(buf *bytes.Buffer, msg Message, hasPlain, hasHTML bool) {
	writer := multipart.NewWriter(buf)
	writeHeader(buf, "Content-Type", "multipart/mixed; boundary="+writer.Boundary())
	buf.WriteString("\r\n")

	// Write the text body part(s).
	if hasPlain && hasHTML {
		// Create an inner multipart/alternative part.
		altBuf := &bytes.Buffer{}
		altWriter := multipart.NewWriter(altBuf)

		plainHeader := make(textproto.MIMEHeader)
		plainHeader.Set("Content-Type", "text/plain; charset=utf-8")
		plainHeader.Set("Content-Transfer-Encoding", "quoted-printable")
		part, err := altWriter.CreatePart(plainHeader)
		if err == nil {
			part.Write([]byte(msg.PlainBody))
		}

		htmlHeader := make(textproto.MIMEHeader)
		htmlHeader.Set("Content-Type", "text/html; charset=utf-8")
		htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")
		part, err = altWriter.CreatePart(htmlHeader)
		if err == nil {
			part.Write([]byte(msg.HTMLBody))
		}
		altWriter.Close()

		// Write the alternative part into the mixed writer.
		altHeader := make(textproto.MIMEHeader)
		altHeader.Set("Content-Type", "multipart/alternative; boundary="+altWriter.Boundary())
		part, err = writer.CreatePart(altHeader)
		if err == nil {
			part.Write(altBuf.Bytes())
		}
	} else if hasHTML {
		htmlHeader := make(textproto.MIMEHeader)
		htmlHeader.Set("Content-Type", "text/html; charset=utf-8")
		htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")
		part, err := writer.CreatePart(htmlHeader)
		if err == nil {
			part.Write([]byte(msg.HTMLBody))
		}
	} else {
		plainHeader := make(textproto.MIMEHeader)
		plainHeader.Set("Content-Type", "text/plain; charset=utf-8")
		plainHeader.Set("Content-Transfer-Encoding", "quoted-printable")
		part, err := writer.CreatePart(plainHeader)
		if err == nil {
			part.Write([]byte(msg.PlainBody))
		}
	}

	// Write attachment parts.
	for _, att := range msg.Attachments {
		attHeader := make(textproto.MIMEHeader)
		attHeader.Set("Content-Type", att.ContentType)
		attHeader.Set("Content-Transfer-Encoding", "base64")
		attHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", att.Filename))
		part, err := writer.CreatePart(attHeader)
		if err == nil {
			encoded := base64.StdEncoding.EncodeToString(att.Data)
			part.Write([]byte(encoded))
		}
	}

	writer.Close()
}
