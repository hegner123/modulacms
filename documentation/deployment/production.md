# Production Deployment

ModulaCMS runs as a single binary with built-in HTTP, HTTPS (with automatic Let's Encrypt certificates), and SSH servers. No reverse proxy is required for HTTPS -- the binary handles TLS termination directly.

This guide covers deploying ModulaCMS to a Linux server with automatic CI/CD via GitHub Actions and manual deployment as a fallback.

## Prerequisites

- A Linux server (AMD64) with root or sudo access
- A domain name with DNS pointing to your server's IP address
- Ports 80, 443, and your SSH port (default 2233) open in the firewall
- Go 1.24+ with CGO enabled (for building from source)

## Server Setup

### Create the Deployment Directory

```bash
mkdir -p /root/app/modula
mkdir -p /root/app/modula/certs
mkdir -p /root/app/modula/logs
mkdir -p /root/app/modula/backups

chmod 755 /root/app/modula
chmod 700 /root/app/modula/certs
chmod 755 /root/app/modula/logs
chmod 755 /root/app/modula/backups
```

### Create the systemd Service

Create `/etc/systemd/system/modulacms.service`:

```ini
[Unit]
Description=ModulaCMS Headless CMS
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/app/modula
ExecStart=/root/app/modula/modulacms-amd
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

Environment="MODULACMS_ENV=production"

LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable modulacms
sudo systemctl start modulacms
sudo systemctl status modulacms
```

### Configure the Firewall

```bash
sudo ufw allow 22/tcp    # SSH management
sudo ufw allow 80/tcp    # HTTP (required for Let's Encrypt challenge)
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 2233/tcp  # ModulaCMS SSH TUI (adjust to match your ssh_port)
sudo ufw enable
```

### Configure DNS

Before ModulaCMS can obtain Let's Encrypt certificates, your domain must resolve to the server:

```bash
dig +short your-domain.com
# Should return your server's IP address

dig +short admin.your-domain.com
# Should also return your server's IP address
```

Let's Encrypt requires your domain to be publicly accessible on port 80 for HTTP-01 challenge verification.

## Configuration

Create `/root/app/modula/modula.config.json`:

```json
{
  "environment": "production",
  "os": "linux",
  "environment_hosts": {
    "local": "localhost",
    "development": "localhost",
    "staging": "staging.your-domain.com",
    "production": "your-domain.com"
  },
  "port": ":80",
  "ssl_port": ":443",
  "cert_dir": "/root/app/modula/certs/",
  "client_site": "your-domain.com",
  "admin_site": "admin.your-domain.com",
  "ssh_host": "localhost",
  "ssh_port": "2233",
  "log_path": "/root/app/modula/logs/",
  "db_driver": "sqlite",
  "db_url": "/root/app/modula/modula.db",
  "db_name": "modula.db",
  "cors_origins": ["https://your-domain.com", "https://admin.your-domain.com"],
  "cors_methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
  "cors_headers": ["Content-Type", "Authorization"],
  "cors_credentials": true,
  "cookie_secure": true,
  "cookie_samesite": "lax",
  "backup_option": "/root/app/modula/backups/",
  "update_auto_enabled": false,
  "update_check_interval": "startup",
  "update_channel": "stable"
}
```

### Key Configuration Fields

**Network and TLS:**

| Field | Value | Notes |
|-------|-------|-------|
| `environment` | `"production"` | Enables automatic Let's Encrypt certificates |
| `port` | `":80"` | Standard HTTP port |
| `ssl_port` | `":443"` | Standard HTTPS port |
| `cert_dir` | Path to cert directory | Let's Encrypt stores certificates here |
| `client_site` | Your domain | Whitelisted for Let's Encrypt |
| `admin_site` | Your admin subdomain | Also whitelisted for Let's Encrypt |

**Database:**

| Field | Value | Notes |
|-------|-------|-------|
| `db_driver` | `"sqlite"`, `"mysql"`, or `"postgres"` | Choose your database backend |
| `db_url` | File path or connection string | SQLite: file path. MySQL/PostgreSQL: connection string |
| `db_name` | Database name | Used with MySQL and PostgreSQL |

**Security:**

| Field | Value | Notes |
|-------|-------|-------|
| `auth_salt` | Auto-generated on first run | Can be set manually for consistency across instances |
| `cookie_secure` | `true` | Required for HTTPS |
| `cors_origins` | Array of allowed origins | Use your actual domains |

**Optional -- S3 Storage:**
Configure `bucket_region`, `bucket_media`, `bucket_endpoint`, `bucket_access_key`, `bucket_secret_key`, and related fields if using S3-compatible storage for media.

**Optional -- OAuth:**
Configure `oauth_client_id`, `oauth_client_secret`, and `oauth_endpoint` fields if enabling OAuth authentication with Google, GitHub, or Azure.

## How HTTPS Works

When `environment` is set to anything other than `"local"`, ModulaCMS automatically:

1. Uses `golang.org/x/crypto/acme/autocert` for Let's Encrypt integration
2. Whitelists domains from `environment_hosts[environment]`, `client_site`, and `admin_site`
3. Obtains SSL certificates on the first HTTPS request
4. Stores certificates in `cert_dir`
5. Renews certificates automatically before expiration
6. Serves both HTTP (port 80) and HTTPS (port 443) concurrently

No reverse proxy or manual certificate renewal is needed.

## Deploying

### Automatic Deployment via GitHub Actions

Push to the `dev` or `develop` branch to trigger automatic deployment. The workflow runs tests, builds a Linux AMD64 binary, deploys it to your server via SSH, and restarts the service.

Configure these GitHub repository secrets (Settings > Secrets and variables > Actions):

**Required:**

| Secret | Description | Example |
|--------|-------------|---------|
| `DEPLOY_SSH_KEY` | SSH private key for the deployment server | Contents of `~/.ssh/modulacms_deploy` |
| `DEPLOY_HOST` | Server hostname or IP | `your-server.com` |
| `DEPLOY_USER` | SSH username | `root` |

**Optional:**

| Secret | Description | Default |
|--------|-------------|---------|
| `DEPLOY_PATH` | Deployment directory on server | `/root/app/modula` |
| `HEALTH_CHECK_URL` | URL for post-deploy health check | (none) |

Generate a deployment SSH key:

```bash
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/modulacms_deploy
ssh-copy-id -i ~/.ssh/modulacms_deploy.pub root@your-server.com
```

Copy the private key contents into the `DEPLOY_SSH_KEY` secret.

### Manual Deployment

Build and deploy from your local machine:

```bash
just build
```

Or build and transfer manually:

```bash
# Build for Linux AMD64
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -mod vendor -o modulacms-amd ./cmd

# Transfer to server
scp modulacms-amd root@your-server.com:/root/app/modula/

# Restart the service
ssh root@your-server.com "sudo systemctl restart modulacms"
```

### Creating Releases

Tag a version to create a GitHub release with binaries for all platforms:

```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

The CI workflow builds for darwin/linux on amd64/arm64, creates a GitHub release, and generates release notes.

## Monitoring

### Service Status

```bash
sudo systemctl status modulacms
```

### Logs

```bash
# Follow live logs
sudo journalctl -u modulacms -f

# Last 100 lines
sudo journalctl -u modulacms -n 100

# Logs from the last hour
sudo journalctl -u modulacms --since "1 hour ago"
```

### Rollback

If a deployment introduces problems, restore the backup binary:

```bash
ssh root@your-server.com
sudo systemctl stop modulacms
cd /root/app/modula
cp modulacms-amd.backup modulacms-amd
sudo systemctl start modulacms
sudo systemctl status modulacms
```

## Troubleshooting

### SSH Authentication Failure During Deploy

`Permission denied (publickey)` during CI deployment.

- Verify the SSH private key is correctly stored in the `DEPLOY_SSH_KEY` GitHub secret
- Confirm the public key is in `~/.ssh/authorized_keys` on the server
- Check permissions: `chmod 600 ~/.ssh/authorized_keys`

### Service Fails to Start

Check the journal for details:

```bash
sudo journalctl -u modulacms -n 50 --no-pager
```

Common causes:
- **Ports in use:** `sudo lsof -i :80` and `sudo lsof -i :443`
- **Missing database file:** `ls -la /root/app/modula/modula.db`
- **Permission issues:** `chmod +x /root/app/modula/modulacms-amd`

### Build Fails with CGO Errors

Cross-compiling CGO code (required for SQLite) needs a C cross-compiler:

```bash
# On macOS targeting Linux
brew install FiloSottile/musl-cross/musl-cross
```

Alternatively, build directly on the target Linux server.

## Security Recommendations

- Use a dedicated deployment user instead of `root`
- Restrict the deployment SSH key to deployment operations only
- Rotate deployment SSH keys every 6 months
- Never commit credentials to the repository
- Enable the firewall and only allow necessary ports
- Monitor logs for unauthorized access attempts

## Optional: Reverse Proxy

ModulaCMS handles HTTPS natively, so a reverse proxy is not required. If you need one for load balancing, additional security layers, or serving multiple applications on the same server, you can place Caddy or Nginx in front of ModulaCMS. See the `deploy/Caddyfile` in the repository for an example configuration.
