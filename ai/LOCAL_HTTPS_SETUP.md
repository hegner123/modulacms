# Running ModulaCMS Locally with HTTPS

This guide explains how to run ModulaCMS with HTTPS on your local machine for development.

## Quick Setup

### 1. Generate Self-Signed Certificates

#### Option A: Using ModulaCMS Built-in Generator (Easiest)

```bash
# Generate certificates automatically
./modulacms-x86 --gen-certs

# This will create:
# - certs/localhost.crt
# - certs/localhost.key
```

#### Option B: Using OpenSSL

```bash
# Create certs directory if it doesn't exist
mkdir -p certs

# Generate self-signed certificate (valid for 365 days)
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout certs/localhost.key \
  -out certs/localhost.crt \
  -days 365 \
  -subj "/CN=localhost"

# Verify certificates were created
ls -la certs/
```

### 2. Configure ModulaCMS for Local HTTPS

Update your `config.json`:

```json
{
  "environment": "local",
  "os": "darwin",
  "port": ":8080",
  "ssl_port": ":4443",
  "cert_dir": "./certs/",
  "client_site": "localhost",
  "admin_site": "localhost",
  "db_driver": "sqlite",
  "db_url": "./modula.db",
  "cookie_secure": false,
  "cors_origins": ["https://localhost:4443"]
}
```

**Key settings:**
- **`environment`**: Set to `"local"` (enables HTTPS with self-signed certs)
- **`cert_dir`**: Directory containing `localhost.crt` and `localhost.key`
- **`ssl_port`**: HTTPS port (e.g., `:4443`)
- **`cookie_secure`**: Keep as `false` for local development

### 3. Run ModulaCMS

```bash
just dev
./modulacms-x86
```

You should see:
```
Server is running at https://localhost:4443
Server is running at http://localhost:8080
```

### 4. Trust the Self-Signed Certificate

Your browser will show a security warning because the certificate is self-signed. You have two options:

#### Option A: Accept Browser Warning (Quick)
- Visit https://localhost:4443
- Click "Advanced" → "Proceed to localhost (unsafe)"
- This is fine for local development

#### Option B: Trust the Certificate (Better)

**macOS:**
```bash
# Add certificate to keychain
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain certs/localhost.crt

# Verify
security find-certificate -c localhost -a | grep "localhost"
```

**Linux:**
```bash
# Copy certificate to system certificates
sudo cp certs/localhost.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

**Windows:**
```powershell
# Import certificate (Run as Administrator)
Import-Certificate -FilePath certs/localhost.crt -CertStoreLocation Cert:\LocalMachine\Root
```

## HTTP-Only Mode (No HTTPS)

If you want to disable HTTPS entirely (not recommended):

```json
{
  "environment": "http-only",
  "os": "darwin",
  "port": ":8080",
  "ssl_port": ":4443",
  "cert_dir": "./certs/",
  "client_site": "localhost",
  "admin_site": "localhost"
}
```

When `environment` is set to `"http-only"`, ModulaCMS will only run the HTTP server.

## Advanced: Custom Domain for Local Development

To use a custom domain locally (e.g., `https://modulacms.local`):

### 1. Add to `/etc/hosts`

```bash
# Edit hosts file
sudo nano /etc/hosts

# Add this line
127.0.0.1   modulacms.local
```

### 2. Generate Certificate for Custom Domain

```bash
# Generate certificate with custom domain
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout certs/modulacms.local.key \
  -out certs/modulacms.local.crt \
  -days 365 \
  -subj "/CN=modulacms.local" \
  -addext "subjectAltName=DNS:modulacms.local"

# Rename to localhost.crt/key (what ModulaCMS expects)
mv certs/modulacms.local.crt certs/localhost.crt
mv certs/modulacms.local.key certs/localhost.key
```

### 3. Update Config

```json
{
  "environment": "local",
  "os": "darwin",
  "client_site": "modulacms.local",
  "admin_site": "admin.modulacms.local",
  "port": ":80",
  "ssl_port": ":443",
  "cert_dir": "./certs/"
}
```

**Note:** Running on ports 80/443 requires sudo:
```bash
sudo ./modulacms-x86
```

## Using mkcert (Easiest Option)

[mkcert](https://github.com/FiloSottile/mkcert) automatically creates locally-trusted certificates:

### 1. Install mkcert

```bash
# macOS
brew install mkcert
brew install nss # for Firefox

# Linux
sudo apt install libnss3-tools
wget https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-linux-amd64
chmod +x mkcert-v1.4.4-linux-amd64
sudo mv mkcert-v1.4.4-linux-amd64 /usr/local/bin/mkcert

# Windows (with Chocolatey)
choco install mkcert
```

### 2. Setup Local CA

```bash
# Install local CA
mkcert -install
```

### 3. Generate Certificates

```bash
# Create certs directory
mkdir -p certs

# Generate certificates for localhost
mkcert -key-file certs/localhost.key -cert-file certs/localhost.crt localhost 127.0.0.1 ::1

# Or for custom domain
mkcert -key-file certs/localhost.key -cert-file certs/localhost.crt modulacms.local "*.modulacms.local"
```

### 4. Run ModulaCMS

```bash
just dev
./modulacms-x86
```

**No browser warnings!** The certificate is automatically trusted.

## Environment Options Summary

| Environment | HTTPS | Certificate | Use Case |
|------------|-------|-------------|----------|
| `local` | ✅ Yes | Self-signed (cert_dir) | Local development with HTTPS |
| `http-only` | ❌ No | N/A | HTTP-only development |
| `development` | ✅ Yes | Let's Encrypt (autocert) | Dev server with real domain |
| `staging` | ✅ Yes | Let's Encrypt (autocert) | Staging server |
| `production` | ✅ Yes | Let's Encrypt (autocert) | Production server |

## Testing HTTPS Locally

```bash
# Test HTTP
curl http://localhost:8080/

# Test HTTPS (skip cert verification)
curl -k https://localhost:4443/

# Test with certificate
curl --cacert certs/localhost.crt https://localhost:4443/
```

## Troubleshooting

### Certificate not found error

```
Error: failed to load certificate
```

**Solution:** Ensure certificates exist in `cert_dir`:
```bash
ls -la ./certs/
# Should show: localhost.crt and localhost.key
```

### Port already in use

```
Error: bind: address already in use
```

**Solution:** Check what's using the port:
```bash
# Check port 4443
lsof -i :4443

# Kill the process
kill -9 <PID>
```

### Browser shows NET::ERR_CERT_INVALID

This is normal for self-signed certificates. Either:
1. Click "Advanced" → "Proceed"
2. Use mkcert for trusted certificates
3. Manually trust the certificate (see "Trust the Self-Signed Certificate" above)

### HTTPS not starting

Check the logs for certificate errors:
```bash
./modulacms-x86 2>&1 | grep -i cert
```

Ensure:
- `cert_dir` exists and is readable
- `localhost.crt` and `localhost.key` are in `cert_dir`
- Certificate files have correct permissions: `chmod 644 certs/localhost.crt` and `chmod 600 certs/localhost.key`

## Why Use HTTPS Locally?

1. **Match Production Environment** - Test with the same protocol as production
2. **Test Security Features** - Cookies with `Secure` flag, CSP headers, etc.
3. **Service Workers** - Required for PWAs and service workers
4. **Modern APIs** - Geolocation, camera, microphone require HTTPS
5. **OAuth/SSO Testing** - Many OAuth providers require HTTPS redirect URLs

## Next Steps

- For production deployment, see [DEPLOYMENT.md](DEPLOYMENT.md)
- For OAuth setup, see `ai/domain/AUTH_AND_OAUTH.md`
- For configuration options, see [internal/config/config.go](internal/config/config.go)
