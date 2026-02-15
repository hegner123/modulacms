# ModulaCMS Deployment Guide

This guide explains how to set up automatic deployment for ModulaCMS when pushing to the `dev` or `develop` branches.

## Overview

The GitHub Actions workflow automatically:
1. Runs tests on every push
2. Builds and deploys to your dev server when pushing to `dev` or `develop` branches
3. Creates GitHub releases when you push version tags (e.g., `v1.0.0`)

## Setup Instructions

### 1. Configure GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions, and add these secrets:

#### Required Secrets:

- **`DEPLOY_SSH_KEY`**: Your SSH private key for the deployment server
  ```bash
  # Generate a new SSH key pair (on your local machine)
  ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/modulacms_deploy

  # Copy the private key content
  cat ~/.ssh/modulacms_deploy
  # Paste the entire output into DEPLOY_SSH_KEY secret

  # Copy the public key to your server
  ssh-copy-id -i ~/.ssh/modulacms_deploy.pub root@your-server.com
  ```

- **`DEPLOY_HOST`**: Your server hostname or IP
  ```
  Example: your-server.com
  Or: 123.45.67.89
  ```

- **`DEPLOY_USER`**: SSH username for deployment
  ```
  Example: root
  Or: deploy
  ```

#### Optional Secrets:

- **`DEPLOY_PATH`**: Custom deployment path (default: `/root/app/modula`)
  ```
  Example: /opt/modulacms
  ```

- **`HEALTH_CHECK_URL`**: URL for health checks after deployment
  ```
  Example: https://your-domain.com/health
  ```

### 2. Setup Systemd Service on Server

SSH into your server and create the systemd service file:

```bash
sudo nano /etc/systemd/system/modulacms.service
```

Paste this configuration:

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

# Environment variables
Environment="MODULACMS_ENV=production"

# Resource limits
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

**Enable and start the service:**

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable modulacms

# Start the service
sudo systemctl start modulacms

# Check status
sudo systemctl status modulacms
```

### 3. Setup Deployment Directory

Create the deployment directory on your server:

```bash
# Create main directory
mkdir -p /root/app/modula

# Create subdirectories
mkdir -p /root/app/modula/certs     # For Let's Encrypt certificates
mkdir -p /root/app/modula/logs      # For application logs
mkdir -p /root/app/modula/backups   # For database backups

# Ensure proper permissions
chmod 755 /root/app/modula
chmod 700 /root/app/modula/certs    # Secure cert directory
chmod 755 /root/app/modula/logs
chmod 755 /root/app/modula/backups
```

### 4. Create Production Configuration

Copy the production config template to your server and customize it:

```bash
# On your local machine (from project root)
scp deploy/config.production.json root@your-server.com:/root/app/modula/config.json

# Or create it directly on the server
ssh root@your-server.com
nano /root/app/modula/config.json
```

**Production config.json template (see `deploy/config.production.json`):**
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

**Key Configuration Fields to Customize:**

**Environment Settings:**
- **`environment`**: Set to `"production"` (enables HTTPS with Let's Encrypt)
- **`os`**: Operating system (`"linux"`, `"darwin"`, `"windows"`) - auto-detected by default
- **`environment_hosts.production`**: Your production domain (e.g., `"example.com"`)
- **`client_site`**: Primary domain for your CMS (e.g., `"example.com"`)
- **`admin_site`**: Admin interface domain (e.g., `"admin.example.com"`)
- **`port`**: HTTP port (`:80` for standard HTTP)
- **`ssl_port`**: HTTPS port (`:443` for standard HTTPS)
- **`cert_dir`**: Directory for Let's Encrypt certificates (e.g., `"/root/app/modula/certs/"`)

**Database Settings:**
- **`db_driver`**: `"sqlite"`, `"mysql"`, or `"postgres"`
- **`db_url`**: Database file path (SQLite) or connection string (MySQL/PostgreSQL)
- **`db_name`**: Database name

**Security Settings:**
- **`auth_salt`**: Auto-generated on first run, or set manually
- **`cookie_secure`**: Must be `true` for HTTPS
- **`cors_origins`**: Array of allowed origins (use your actual domains)

**Optional - S3 Storage:**
- Configure `bucket_*` fields if using S3-compatible storage for media

**Optional - OAuth:**
- Configure `oauth_*` fields if enabling OAuth authentication

**Optional: Add Health Check Endpoint**

To enable health checks, add this endpoint to your ModulaCMS router (if not already present):

```go
// In your router setup
router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "version": utility.Version,
        "commit": utility.GitCommit,
    })
})
```

### 5. Configure Server Firewall

ModulaCMS has built-in HTTPS with Let's Encrypt, so it needs direct access to ports 80 and 443:

```bash
# UFW example
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP (required for Let's Encrypt challenge)
sudo ufw allow 443/tcp  # HTTPS (ModulaCMS HTTPS server)
sudo ufw enable
```

**Note:** ModulaCMS runs both HTTP and HTTPS servers directly, with automatic Let's Encrypt certificate management. No reverse proxy needed!

### 6. DNS Configuration

Before ModulaCMS can obtain Let's Encrypt certificates, ensure your DNS is configured:

```bash
# Verify your domain points to the server
dig +short your-domain.com
# Should return your server's IP address

# If using admin subdomain
dig +short admin.your-domain.com
# Should also return your server's IP address
```

**Note:** Let's Encrypt requires that your domain is publicly accessible on port 80 for HTTP-01 challenge verification.

**How ModulaCMS HTTPS Works:**

When `environment` is set to anything except `"local"`, ModulaCMS automatically:
1. Uses `golang.org/x/crypto/acme/autocert` for Let's Encrypt integration
2. Whitelists domains from `environment_hosts[environment]`, `client_site`, and `admin_site`
3. Obtains SSL certificates automatically on first HTTPS request
4. Stores certificates in `cert_dir`
5. Renews certificates before expiration (autocert handles this)
6. Serves both HTTP (port 80) and HTTPS (port 443)

**Local Development:**

For local development (`environment: "local"`), ModulaCMS uses self-signed certificates from `cert_dir`:
- Place `localhost.crt` and `localhost.key` in your `cert_dir`
- Or generate them: `openssl req -x509 -newkey rsa:4096 -keyout localhost.key -out localhost.crt -days 365 -nodes`

### 7. Optional: Reverse Proxy (Only if Needed)

If you need a reverse proxy for load balancing, additional security layers, or serving multiple applications, you can place Caddy in front of ModulaCMS. However, **this is not required** for HTTPS since ModulaCMS handles it natively.

See `deploy/Caddyfile` for an example reverse proxy configuration if needed.

## Usage

### Automatic Deployment (Dev)

Push to `dev` or `develop` branch to trigger automatic deployment:

```bash
git checkout dev
git add .
git commit -m "feat: add new feature"
git push origin dev
```

The workflow will:
1. ✅ Run tests
2. 🔨 Build for Linux AMD64
3. 📦 Deploy to server
4. 🔄 Restart service
5. ✔️ Run health checks

### Manual Deployment

If you prefer to deploy manually:

```bash
# From your local machine
just build

# This will:
# - Build the binary
# - rsync to your server (configured in justfile)
```

### Creating Production Releases

To create a production release:

```bash
# Tag a version
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

This will:
1. ✅ Run tests
2. 🔨 Build for all platforms (darwin/linux, amd64/arm64)
3. 📦 Create GitHub release with binaries
4. 📝 Generate release notes

## Monitoring Deployment

### View GitHub Actions logs

1. Go to your repository on GitHub
2. Click "Actions" tab
3. Click on the workflow run
4. View deployment logs

### Check service status on server

```bash
# Check if service is running
sudo systemctl status modulacms

# View logs
sudo journalctl -u modulacms -f

# View last 100 lines
sudo journalctl -u modulacms -n 100

# Check specific time range
sudo journalctl -u modulacms --since "1 hour ago"
```

### Rollback if needed

```bash
# SSH into server
ssh root@your-server.com

# Stop service
sudo systemctl stop modulacms

# Restore backup
cd /root/app/modula
cp modulacms-amd.backup modulacms-amd

# Start service
sudo systemctl start modulacms

# Verify
sudo systemctl status modulacms
```

## Troubleshooting

### Deployment fails with SSH authentication error

**Problem:** `Permission denied (publickey)`

**Solution:**
1. Verify SSH key is correctly added to GitHub Secrets
2. Ensure public key is in `~/.ssh/authorized_keys` on server
3. Check SSH key permissions on server: `chmod 600 ~/.ssh/authorized_keys`

### Service fails to start after deployment

**Problem:** Service shows as failed in systemd

**Solution:**
```bash
# Check detailed error logs
sudo journalctl -u modulacms -n 50 --no-pager

# Common issues:
# 1. Ports already in use (80/443 required for HTTPS)
sudo lsof -i :80
sudo lsof -i :443

# 2. Missing database file
ls -la /root/app/modula/modula.db

# 3. Permission issues
sudo chown -R root:root /root/app/modula
chmod +x /root/app/modula/modulacms-amd
```

### Health check fails

**Problem:** Health check timeout or failure

**Solution:**
1. Ensure ModulaCMS is listening on the expected port
2. Check firewall rules: `sudo ufw status`
3. Verify health endpoint exists (if using HEALTH_CHECK_URL)
4. Check application logs for startup errors

### Build fails with CGO errors

**Problem:** Cross-compilation issues with SQLite

**Solution:**
The workflow installs cross-compilation tools automatically. If issues persist:
1. Check Go version matches (1.24.2)
2. Ensure vendor directory is up to date: `just vendor`
3. Verify `go.mod` has correct CGO settings

## Security Best Practices

1. **Use dedicated deployment user:** Create a user specifically for deployments instead of using `root`
2. **Restrict SSH key:** Use SSH key only for deployment, not general access
3. **Rotate keys regularly:** Update deployment SSH keys every 6 months
4. **Use secrets management:** Never commit credentials to repository
5. **Enable firewall:** Only allow necessary ports
6. **Use HTTPS:** Always use TLS in production
7. **Monitor logs:** Set up log monitoring and alerts

## Advanced Configuration

### Deploy to Multiple Environments

Create separate secrets for staging and production:

```
STAGING_DEPLOY_HOST
STAGING_DEPLOY_USER
PRODUCTION_DEPLOY_HOST
PRODUCTION_DEPLOY_USER
```

Update workflow to handle different environments based on branch.

### Database Migration on Deployment

Add migration step in workflow before restart:

```yaml
- name: Run migrations
  run: |
    ssh $DEPLOY_USER@$DEPLOY_HOST "cd $DEPLOY_PATH && ./modulacms-amd --migrate"
```

### Blue-Green Deployment

For zero-downtime deployments, use two instances and a load balancer:

1. Deploy to inactive instance
2. Run health checks
3. Switch load balancer
4. Update other instance

## Support

For issues with deployment:
1. Check GitHub Actions logs
2. Check server systemd logs: `journalctl -u modulacms`
3. Review TROUBLESHOOTING.md
4. Open an issue at https://github.com/hegner123/modulacms/issues
