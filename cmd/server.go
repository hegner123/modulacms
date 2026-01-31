package main

import (
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func newHTTPServer(addr string, handler http.Handler, tlsConfig *tls.Config) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		TLSConfig:    tlsConfig,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func sanitizeCertDir(configCertDir string) (string, error) {
	if strings.TrimSpace(configCertDir) == "" {
		return "", errors.New("certificate directory path cannot be empty")
	}

	certDir := filepath.Clean(configCertDir)

	absPath, err := filepath.Abs(certDir)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", errors.New("certificate path is not a directory")
	}

	return absPath, nil
}
