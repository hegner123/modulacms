package email

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hegner123/modulacms/internal/config"
)

// SESSender sends email via AWS SES using the v1 SDK SendRawEmail API.
type SESSender struct {
	sess   *session.Session
	client *ses.SES
	region string
}

// newSESSender creates an SESSender from config. Uses static credentials if
// Email_AWS_Access_Key_ID and Email_AWS_Secret_Access_Key are provided;
// otherwise falls back to the default AWS credential chain (env vars, IAM role, etc.).
func newSESSender(cfg config.Config) (*SESSender, error) {
	region := detectSESRegion(cfg.Email_API_Endpoint)

	awsCfg := aws.NewConfig().WithRegion(region)

	// Use static credentials if provided, otherwise default credential chain.
	if cfg.Email_AWS_Access_Key_ID != "" && cfg.Email_AWS_Secret_Access_Key != "" {
		awsCfg.WithCredentials(credentials.NewStaticCredentials(
			cfg.Email_AWS_Access_Key_ID,
			cfg.Email_AWS_Secret_Access_Key,
			"",
		))
	}

	// Override endpoint if configured.
	if cfg.Email_API_Endpoint != "" {
		awsCfg.WithEndpoint(cfg.Email_API_Endpoint)
	}

	sess, err := session.NewSession(awsCfg)
	if err != nil {
		return nil, &ProviderError{Provider: "ses", Err: fmt.Errorf("create session: %w", err)}
	}

	return &SESSender{
		sess:   sess,
		client: ses.New(sess),
		region: region,
	}, nil
}

// Send delivers a message via SES SendRawEmail with full MIME bytes.
func (s *SESSender) Send(ctx context.Context, msg Message) error {
	mimeData, err := buildMIME(msg)
	if err != nil {
		return &ProviderError{Provider: "ses", Err: fmt.Errorf("build MIME: %w", err)}
	}

	recipients := msg.allRecipients()
	destinations := make([]*string, len(recipients))
	for i, r := range recipients {
		destinations[i] = aws.String(r)
	}

	input := &ses.SendRawEmailInput{
		RawMessage:   &ses.RawMessage{Data: mimeData},
		Destinations: destinations,
	}

	_, err = s.client.SendRawEmailWithContext(ctx, input)
	if err != nil {
		return wrapAWSError(err)
	}

	return nil
}

// Close releases idle HTTP connections held by the SES client.
func (s *SESSender) Close() error {
	if s.client != nil && s.client.Config.HTTPClient != nil {
		s.client.Config.HTTPClient.CloseIdleConnections()
	}
	return nil
}

// detectSESRegion extracts the AWS region from an SES endpoint URL.
// If the endpoint contains ".amazonaws.com", splits on dots and takes the
// second segment (e.g., "email.us-west-2.amazonaws.com" -> "us-west-2").
// Otherwise defaults to "us-east-1".
func detectSESRegion(endpoint string) string {
	if endpoint == "" {
		return "us-east-1"
	}

	// Strip scheme if present.
	host := endpoint
	if idx := strings.Index(host, "://"); idx >= 0 {
		host = host[idx+3:]
	}
	// Strip path.
	if idx := strings.IndexByte(host, '/'); idx >= 0 {
		host = host[:idx]
	}

	if strings.Contains(host, ".amazonaws.com") {
		parts := strings.Split(host, ".")
		if len(parts) >= 3 {
			return parts[1]
		}
	}

	return "us-east-1"
}

// wrapAWSError wraps an AWS SDK error into a ProviderError, extracting
// the HTTP status code from awserr.RequestFailure if available.
func wrapAWSError(err error) *ProviderError {
	var reqErr awserr.RequestFailure
	if ok := errors.As(err, &reqErr); ok {
		return &ProviderError{
			Provider: "ses",
			Code:     reqErr.StatusCode(),
			Err:      err,
		}
	}
	return &ProviderError{Provider: "ses", Err: err}
}
