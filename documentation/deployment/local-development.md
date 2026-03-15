# Local Development with HTTPS

Set up HTTPS for local development to test secure cookies, OAuth flows, and other TLS-dependent features.

## Quick Start

### 1. Generate Self-Signed Certificates

**Option A: Built-in generator (recommended)**

```bash
./modulacms-x86 --gen-certs
```

This creates `certs/localhost.crt` and `certs/localhost.key`.

**Option B: OpenSSL**

```bash
mkdir -p certs
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout certs/localhost.key \
  -out certs/localhost.crt \
  -days 365 \
  -subj "/CN=localhost"
```

### 2. Configure for Local HTTPS

Set these fields in your `modula.config.json`:

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

Key settings:
- `environment` set to `"local"` enables HTTPS using self-signed certificates from `cert_dir`
- `ssl_port` determines the HTTPS port (`:4443` avoids needing root privileges)
- `cookie_secure` should be `false` for local development to avoid cookie issues with self-signed certs

### 3. Build and Run

```bash
just dev
./modulacms-x86
```

Output:

```
Server is running at https://localhost:4443
Server is running at http://localhost:8080
```

### 4. Handle the Browser Warning

Your browser will show a security warning for self-signed certificates. You have two options:

**Accept the warning (quick):** Visit `https://localhost:4443`, click "Advanced", then "Proceed to localhost (unsafe)".

**Trust the certificate (persistent):**

macOS:

```bash
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain certs/localhost.crt
```

Linux:

```bash
sudo cp certs/localhost.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

### 5. Test the Connection

```bash
# Test HTTP
curl http://localhost:8080/

# Test HTTPS (skip certificate verification)
curl -k https://localhost:4443/

# Test HTTPS (with certificate)
curl --cacert certs/localhost.crt https://localhost:4443/
```

## Using mkcert (No Browser Warnings)

[mkcert](https://github.com/FiloSottile/mkcert) creates locally-trusted certificates that browsers accept without warnings.

Install mkcert:

```bash
# macOS
brew install mkcert
brew install nss  # for Firefox support

# Linux
sudo apt install libnss3-tools
# Download mkcert binary from https://github.com/FiloSottile/mkcert/releases
```

Set up the local CA and generate certificates:

```bash
mkcert -install
mkdir -p certs
mkcert -key-file certs/localhost.key -cert-file certs/localhost.crt localhost 127.0.0.1 ::1
```

Run ModulaCMS normally. Browsers will trust the certificate without any warnings.

## HTTP-Only Mode

To disable HTTPS entirely and run only the HTTP server, set `environment` to `"http-only"`:

```json
{
  "environment": "http-only",
  "port": ":8080",
  "client_site": "localhost",
  "admin_site": "localhost"
}
```

This is useful when you do not need to test TLS-dependent features.

## Custom Local Domain

To use a domain like `modulacms.local` instead of `localhost`:

### 1. Add a hosts entry

```bash
sudo nano /etc/hosts
# Add:
127.0.0.1   modulacms.local
```

### 2. Generate a certificate for the custom domain

```bash
# With OpenSSL
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout certs/localhost.key \
  -out certs/localhost.crt \
  -days 365 \
  -subj "/CN=modulacms.local" \
  -addext "subjectAltName=DNS:modulacms.local"

# Or with mkcert
mkcert -key-file certs/localhost.key -cert-file certs/localhost.crt modulacms.local "*.modulacms.local"
```

> **Good to know**: ModulaCMS expects the certificate files to be named `localhost.crt` and `localhost.key` in the `cert_dir` directory regardless of the domain.

### 3. Update modula.config.json

```json
{
  "environment": "local",
  "client_site": "modulacms.local",
  "admin_site": "admin.modulacms.local",
  "port": ":80",
  "ssl_port": ":443",
  "cert_dir": "./certs/"
}
```

Running on ports 80 and 443 requires elevated privileges:

```bash
sudo ./modulacms-x86
```

## Environment Modes

| Environment | HTTPS | Certificate Source | Use Case |
|-------------|-------|--------------------|----------|
| `local` | Yes | Self-signed from `cert_dir` | Local development with HTTPS |
| `http-only` | No | N/A | HTTP-only development |
| `development` | Yes | Let's Encrypt (autocert) | Dev server with real domain |
| `staging` | Yes | Let's Encrypt (autocert) | Staging server |
| `production` | Yes | Let's Encrypt (autocert) | Production server |

## Troubleshooting

### Certificate not found

```
Error: failed to load certificate
```

Verify the certificate files exist in your `cert_dir`:

```bash
ls -la ./certs/
# Should contain: localhost.crt and localhost.key
```

### Port already in use

```
Error: bind: address already in use
```

Find and stop the process using the port:

```bash
lsof -i :4443
kill <PID>
```

### HTTPS server not starting

Check for certificate-related errors in the output:

```bash
./modulacms-x86 2>&1 | grep -i cert
```

Verify file permissions: `chmod 644 certs/localhost.crt` and `chmod 600 certs/localhost.key`.
