package main

import (
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// handlerSwap allows hot-swapping the HTTP handler at runtime.
// Used to swap from a placeholder (DB not initialized) to the real handler
// after TUI DB init without restarting the HTTP server.
type handlerSwap struct {
	mu sync.RWMutex
	h  http.Handler
}

func (s *handlerSwap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	h := s.h
	s.mu.RUnlock()
	h.ServeHTTP(w, r)
}

func (s *handlerSwap) set(h http.Handler) {
	s.mu.Lock()
	s.h = h
	s.mu.Unlock()
}

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
