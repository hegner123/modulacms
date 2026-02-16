package config

import "encoding/json"

const redactedValue = "********"

// RedactedConfig returns a copy of the config with sensitive fields replaced
// by a redaction placeholder.
func RedactedConfig(c Config) Config {
	redacted := c
	sensitive := SensitiveKeys()

	if sensitive["auth_salt"] {
		redacted.Auth_Salt = redactedValue
	}
	if sensitive["db_password"] {
		redacted.Db_Password = redactedValue
	}
	if sensitive["bucket_access_key"] {
		redacted.Bucket_Access_Key = redactedValue
	}
	if sensitive["bucket_secret_key"] {
		redacted.Bucket_Secret_Key = redactedValue
	}
	if sensitive["oauth_client_id"] {
		redacted.Oauth_Client_Id = redactedValue
	}
	if sensitive["oauth_client_secret"] {
		redacted.Oauth_Client_Secret = redactedValue
	}
	if sensitive["observability_dsn"] {
		redacted.Observability_DSN = redactedValue
	}

	return redacted
}

// RedactedJSON marshals a redacted copy of the config to JSON.
func RedactedJSON(c Config) ([]byte, error) {
	return json.MarshalIndent(RedactedConfig(c), "", "  ")
}

// IsRedactedValue reports whether v is the redaction placeholder.
// Used by the update path to skip fields that were not actually changed.
func IsRedactedValue(v string) bool {
	return v == redactedValue
}
