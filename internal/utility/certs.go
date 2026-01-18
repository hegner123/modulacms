package utility

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// GenerateSelfSignedCert generates a self-signed certificate for local development
func GenerateSelfSignedCert(certDir string, domain string) error {
	// Ensure cert directory exists
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Set up certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Valid for 1 year

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   domain,
			Organization: []string{"ModulaCMS Local Development"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain, "localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	// Create self-signed certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate to file
	certPath := filepath.Join(certDir, "localhost.crt")
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	DefaultLogger.Info("Generated certificate", certPath)

	// Write private key to file
	keyPath := filepath.Join(certDir, "localhost.key")
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyFile.Close()

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Set secure permissions on key file
	if err := os.Chmod(keyPath, 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %w", err)
	}

	DefaultLogger.Info("Generated private key", keyPath)

	return nil
}

// TrustCertificate provides OS-specific instructions or attempts to trust the certificate
func TrustCertificate(certPath string) error {
	osType := runtime.GOOS

	switch osType {
	case "darwin":
		return trustCertMacOS(certPath)
	case "linux":
		return trustCertLinux(certPath)
	case "windows":
		return trustCertWindows(certPath)
	default:
		DefaultLogger.Warn("Automatic certificate trust not supported on", nil, osType)
		DefaultLogger.Info("Please manually trust the certificate:", certPath)
		return nil
	}
}

func trustCertMacOS(certPath string) error {
	DefaultLogger.Info("")
	DefaultLogger.Info("═══════════════════════════════════════════════════════")
	DefaultLogger.Info("To trust this certificate on macOS, run:")
	DefaultLogger.Info("")
	DefaultLogger.Info("  sudo security add-trusted-cert -d -r trustRoot \\")
	DefaultLogger.Info("    -k /Library/Keychains/System.keychain " + certPath)
	DefaultLogger.Info("")
	DefaultLogger.Info("═══════════════════════════════════════════════════════")
	DefaultLogger.Info("")

	// Ask if user wants to trust now
	fmt.Print("Would you like to trust this certificate now? (requires sudo) [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" {
		DefaultLogger.Info("Running certificate trust command...")
		cmd := exec.Command("sudo", "security", "add-trusted-cert", "-d", "-r", "trustRoot",
			"-k", "/Library/Keychains/System.keychain", certPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to trust certificate: %w", err)
		}

		DefaultLogger.Info("✓ Certificate trusted successfully!")
		DefaultLogger.Info("Restart your browser to see the changes")
	} else {
		DefaultLogger.Info("Skipped. You can trust it later with the command above.")
	}

	return nil
}

func trustCertLinux(certPath string) error {
	DefaultLogger.Info("")
	DefaultLogger.Info("═══════════════════════════════════════════════════════")
	DefaultLogger.Info("To trust this certificate on Linux, run:")
	DefaultLogger.Info("")
	DefaultLogger.Info("  sudo cp " + certPath + " /usr/local/share/ca-certificates/modulacms.crt")
	DefaultLogger.Info("  sudo update-ca-certificates")
	DefaultLogger.Info("")
	DefaultLogger.Info("═══════════════════════════════════════════════════════")
	DefaultLogger.Info("")

	// Ask if user wants to trust now
	fmt.Print("Would you like to trust this certificate now? (requires sudo) [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" {
		DefaultLogger.Info("Running certificate trust commands...")

		// Copy certificate
		copyCmd := exec.Command("sudo", "cp", certPath, "/usr/local/share/ca-certificates/modulacms.crt")
		copyCmd.Stdout = os.Stdout
		copyCmd.Stderr = os.Stderr
		copyCmd.Stdin = os.Stdin

		if err := copyCmd.Run(); err != nil {
			return fmt.Errorf("failed to copy certificate: %w", err)
		}

		// Update certificates
		updateCmd := exec.Command("sudo", "update-ca-certificates")
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr

		if err := updateCmd.Run(); err != nil {
			return fmt.Errorf("failed to update certificates: %w", err)
		}

		DefaultLogger.Info("✓ Certificate trusted successfully!")
		DefaultLogger.Info("Restart your browser to see the changes")
	} else {
		DefaultLogger.Info("Skipped. You can trust it later with the commands above.")
	}

	return nil
}

func trustCertWindows(certPath string) error {
	DefaultLogger.Info("")
	DefaultLogger.Info("═══════════════════════════════════════════════════════")
	DefaultLogger.Info("To trust this certificate on Windows, run PowerShell as Administrator:")
	DefaultLogger.Info("")
	DefaultLogger.Info("  Import-Certificate -FilePath \"" + certPath + "\" \\")
	DefaultLogger.Info("    -CertStoreLocation Cert:\\LocalMachine\\Root")
	DefaultLogger.Info("")
	DefaultLogger.Info("Or using certutil:")
	DefaultLogger.Info("  certutil -addstore Root \"" + certPath + "\"")
	DefaultLogger.Info("")
	DefaultLogger.Info("═══════════════════════════════════════════════════════")
	DefaultLogger.Info("")

	// Ask if user wants to trust now
	fmt.Print("Would you like to trust this certificate now? (requires admin) [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" {
		DefaultLogger.Info("Running certificate trust command...")
		cmd := exec.Command("certutil", "-addstore", "Root", certPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			DefaultLogger.Warn("certutil failed, trying PowerShell...", err)

			psCmd := exec.Command("powershell", "-Command",
				fmt.Sprintf("Import-Certificate -FilePath '%s' -CertStoreLocation Cert:\\LocalMachine\\Root", certPath))
			psCmd.Stdout = os.Stdout
			psCmd.Stderr = os.Stderr

			if err := psCmd.Run(); err != nil {
				return fmt.Errorf("failed to trust certificate: %w", err)
			}
		}

		DefaultLogger.Info("✓ Certificate trusted successfully!")
		DefaultLogger.Info("Restart your browser to see the changes")
	} else {
		DefaultLogger.Info("Skipped. You can trust it later with the commands above.")
	}

	return nil
}
