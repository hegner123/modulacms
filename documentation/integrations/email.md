# Email

Configure email delivery so ModulaCMS can send password reset links to your users.

## When Email Is Used

ModulaCMS sends email for one purpose: **password reset flows**. When a user requests a password reset, the CMS generates a time-limited token and emails a reset link. The token expires after 1 hour.

Email is optional. Without it, password resets must be handled by an administrator directly.

## Supported Providers

| Provider | Best for |
|----------|----------|
| SMTP | Any mail server (Gmail, Outlook, self-hosted) |
| SendGrid | Transactional email at scale |
| AWS SES | AWS-native infrastructure |
| Postmark | Deliverability-focused transactional email |

## Common Configuration

These fields apply to all providers. Set them in `modula.config.json`:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `email_enabled` | bool | `false` | Enable email sending |
| `email_provider` | string | `""` | Provider: `smtp`, `sendgrid`, `ses`, `postmark` |
| `email_from_address` | string | `""` | Sender email address (required when enabled) |
| `email_from_name` | string | `""` | Sender display name |
| `email_reply_to` | string | `""` | Default reply-to address |
| `password_reset_url` | string | `""` | Base URL for password reset links |

All email fields are hot-reloadable. Changes take effect without restarting the server.

## Set Up SMTP

```json
{
  "email_enabled": true,
  "email_provider": "smtp",
  "email_from_address": "noreply@example.com",
  "email_from_name": "My CMS",
  "password_reset_url": "https://app.example.com/reset-password",
  "email_host": "smtp.example.com",
  "email_port": 587,
  "email_username": "noreply@example.com",
  "email_password": "your-smtp-password",
  "email_tls": true
}
```

### SMTP Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `email_host` | string | `""` | SMTP server hostname |
| `email_port` | integer | `587` | SMTP server port |
| `email_username` | string | `""` | SMTP auth username |
| `email_password` | string | `""` | SMTP auth password |
| `email_tls` | bool | `true` | Require TLS |

> **Good to know**: Port 587 with TLS is the standard for authenticated SMTP submission. Use port 465 for implicit TLS (SMTPS). Port 25 is typically for server-to-server relay and is blocked by most cloud providers.

## Set Up SendGrid

```json
{
  "email_enabled": true,
  "email_provider": "sendgrid",
  "email_from_address": "noreply@example.com",
  "email_from_name": "My CMS",
  "password_reset_url": "https://app.example.com/reset-password",
  "email_api_key": "SG.your-sendgrid-api-key"
}
```

Create an API key in SendGrid under **Settings > API Keys**. The key needs the **Mail Send** permission.

## Set Up AWS SES

```json
{
  "email_enabled": true,
  "email_provider": "ses",
  "email_from_address": "noreply@example.com",
  "email_from_name": "My CMS",
  "password_reset_url": "https://app.example.com/reset-password",
  "bucket_region": "us-east-1",
  "email_aws_access_key_id": "AKIAIOSFODNN7EXAMPLE",
  "email_aws_secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
}
```

> **Good to know**: When you omit `email_aws_access_key_id` and `email_aws_secret_access_key`, ModulaCMS falls back to the default AWS credential chain (environment variables, IAM role). This is the recommended approach for EC2 or ECS deployments.

### SES Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `email_aws_access_key_id` | string | `""` | AWS access key (optional -- falls back to credential chain) |
| `email_aws_secret_access_key` | string | `""` | AWS secret key (optional -- falls back to credential chain) |

SES uses the `bucket_region` field to determine the AWS region for sending. Verify your sender address in the SES console before sending.

## Set Up Postmark

```json
{
  "email_enabled": true,
  "email_provider": "postmark",
  "email_from_address": "noreply@example.com",
  "email_from_name": "My CMS",
  "password_reset_url": "https://app.example.com/reset-password",
  "email_api_key": "your-postmark-server-token"
}
```

Find your server token in Postmark under **Servers > (your server) > API Tokens**.

### API Provider Fields

These fields apply to SendGrid and Postmark:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `email_api_key` | string | `""` | API key or server token |
| `email_api_endpoint` | string | `""` | Custom API endpoint URL (leave empty to use the provider default) |

## Password Reset URL

The `password_reset_url` field sets the base URL for reset links sent to users. ModulaCMS appends the reset token as a query parameter.

For example, with `"password_reset_url": "https://app.example.com/reset-password"`, the email contains a link like:

```
https://app.example.com/reset-password?token=a1b2c3d4...
```

Your frontend handles this URL, reads the token, and calls the confirm-reset endpoint:

```bash
curl -X POST http://localhost:8080/api/v1/auth/confirm-password-reset \
  -H "Content-Type: application/json" \
  -d '{"token": "a1b2c3d4...", "password": "new-secure-password"}'
```

## Sensitive Fields

ModulaCMS redacts `email_password`, `email_api_key`, `email_aws_access_key_id`, and `email_aws_secret_access_key` when returning configuration through the API or the `config show` CLI command. Updating config via the API skips redacted values automatically to prevent overwriting secrets with the placeholder.

## Next Steps

- [Authentication guide](/docs/custom-admin/authentication) -- password reset flow details
- [OAuth](/docs/integrations/oauth) -- set up third-party login as an alternative to password-based auth
- [Configuration reference](/docs/getting-started/configuration) -- all config fields
