-- user_ssh_keys table for storing SSH public keys linked to user accounts
CREATE TABLE IF NOT EXISTS user_ssh_keys (
    ssh_key_id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL,
    public_key TEXT NOT NULL,
    key_type VARCHAR(50) NOT NULL, -- "ssh-rsa", "ssh-ed25519", "ecdsa-sha2-nistp256", etc.
    fingerprint VARCHAR(255) NOT NULL UNIQUE,
    label VARCHAR(255), -- User-friendly label: "laptop", "work desktop", etc.
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Index for fast lookup by fingerprint during SSH auth
CREATE INDEX IF NOT EXISTS idx_ssh_keys_fingerprint ON user_ssh_keys(fingerprint);

-- Index for listing user's keys
CREATE INDEX IF NOT EXISTS idx_ssh_keys_user_id ON user_ssh_keys(user_id);
