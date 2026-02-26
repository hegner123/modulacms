package utility

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

// ============================================================
// GenerateSelfSignedCert
// ============================================================

func TestGenerateSelfSignedCert_CreatesFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	certDir := filepath.Join(dir, "certs")

	err := GenerateSelfSignedCert(certDir, "localhost")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert returned error: %v", err)
	}

	// Verify cert file exists
	certPath := filepath.Join(certDir, "localhost.crt")
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Fatal("certificate file not created")
	}

	// Verify key file exists
	keyPath := filepath.Join(certDir, "localhost.key")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatal("key file not created")
	}
}

func TestGenerateSelfSignedCert_ValidCertificate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	certDir := filepath.Join(dir, "certs")
	domain := "example.local"

	err := GenerateSelfSignedCert(certDir, domain)
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert returned error: %v", err)
	}

	// Read and parse the certificate
	certPEM, err := os.ReadFile(filepath.Join(certDir, "localhost.crt"))
	if err != nil {
		t.Fatalf("failed to read certificate: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("failed to decode PEM block from certificate")
	}
	if block.Type != "CERTIFICATE" {
		t.Errorf("PEM block type = %q, want %q", block.Type, "CERTIFICATE")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	// Verify domain is in DNSNames
	foundDomain := false
	for _, dns := range cert.DNSNames {
		if dns == domain {
			foundDomain = true
			break
		}
	}
	if !foundDomain {
		t.Errorf("certificate DNSNames %v does not contain %q", cert.DNSNames, domain)
	}

	// Verify localhost is also in DNSNames
	foundLocalhost := false
	for _, dns := range cert.DNSNames {
		if dns == "localhost" {
			foundLocalhost = true
			break
		}
	}
	if !foundLocalhost {
		t.Error("certificate DNSNames does not contain 'localhost'")
	}

	// Verify subject
	if cert.Subject.CommonName != domain {
		t.Errorf("subject CN = %q, want %q", cert.Subject.CommonName, domain)
	}

	// Verify organization
	if len(cert.Subject.Organization) == 0 || cert.Subject.Organization[0] != "ModulaCMS Local Development" {
		t.Errorf("subject Org = %v, want [ModulaCMS Local Development]", cert.Subject.Organization)
	}

	// Verify IP addresses include 127.0.0.1
	foundIPv4 := false
	for _, ip := range cert.IPAddresses {
		if ip.String() == "127.0.0.1" {
			foundIPv4 = true
			break
		}
	}
	if !foundIPv4 {
		t.Error("certificate IP addresses do not include 127.0.0.1")
	}
}

func TestGenerateSelfSignedCert_ValidKey(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	certDir := filepath.Join(dir, "certs")

	err := GenerateSelfSignedCert(certDir, "localhost")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert returned error: %v", err)
	}

	// Read and parse the key
	keyPEM, err := os.ReadFile(filepath.Join(certDir, "localhost.key"))
	if err != nil {
		t.Fatalf("failed to read key: %v", err)
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		t.Fatal("failed to decode PEM block from key")
	}
	if block.Type != "RSA PRIVATE KEY" {
		t.Errorf("PEM block type = %q, want %q", block.Type, "RSA PRIVATE KEY")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	// 4096-bit RSA key should have 512 bytes
	if key.Size() != 512 {
		t.Errorf("key size = %d bytes, want 512 (4096 bits)", key.Size())
	}
}

func TestGenerateSelfSignedCert_KeyFilePermissions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	certDir := filepath.Join(dir, "certs")

	err := GenerateSelfSignedCert(certDir, "localhost")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert returned error: %v", err)
	}

	keyPath := filepath.Join(certDir, "localhost.key")
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("failed to stat key file: %v", err)
	}

	// Key file should have 0600 permissions
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("key file permissions = %04o, want 0600", perm)
	}
}

func TestGenerateSelfSignedCert_CreatesDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Nested directory that does not exist yet
	certDir := filepath.Join(dir, "a", "b", "c", "certs")

	err := GenerateSelfSignedCert(certDir, "localhost")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert returned error: %v", err)
	}

	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		t.Fatal("certificate directory not created")
	}
}

func TestGenerateSelfSignedCert_ExistingDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Should succeed even if directory already exists
	err := GenerateSelfSignedCert(dir, "localhost")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert returned error on existing dir: %v", err)
	}
}

// Note: TrustCertificate is not tested because it prompts for user input (fmt.Scanln)
// and executes privileged system commands (sudo). To test it, the function would need
// to accept an io.Reader for input and a command executor interface.
