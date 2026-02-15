package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/time/rate"
)

// ipLimiterEntry pairs a token-bucket limiter with its last-seen timestamp
// so the cleanup goroutine can evict stale entries.
type ipLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter implements per-IP rate limiting using the token bucket algorithm.
// It tracks individual limiters for each IP address and enforces configurable
// rate limits to prevent abuse of authentication endpoints.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiterEntry
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
	done     chan struct{} // closed to stop cleanup goroutine
}

// NewRateLimiter creates a new rate limiter with the specified rate and burst size.
// The rate parameter controls how many requests per second are allowed.
// The burst parameter controls how many requests can be made in a short burst.
// Example: NewRateLimiter(0.16667, 10) allows 10 requests per minute with burst of 10.
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiterEntry),
		rate:     r,
		burst:    b,
		cleanup:  time.Minute * 10, // Cleanup unused limiters every 10 minutes
		done:     make(chan struct{}),
	}

	// Start background cleanup goroutine
	go rl.cleanupLimiters()

	return rl
}

// getLimiter retrieves or creates a rate limiter for a specific IP address.
// Each IP gets its own independent rate limiter with a tracked last-seen time.
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[ip]
	if !exists {
		entry = &ipLimiterEntry{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.limiters[ip] = entry
		return entry.limiter
	}

	entry.lastSeen = time.Now()
	return entry.limiter
}

// Middleware returns an HTTP middleware handler that enforces rate limits.
// If a client exceeds the rate limit, it receives a 429 Too Many Requests response.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			utility.DefaultLogger.Fwarn("Rate limit exceeded for IP", nil, ip)
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getIP extracts the client IP address from the request.
// It checks X-Forwarded-For and X-Real-IP headers for proxied requests,
// falling back to RemoteAddr if headers are not present.
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxied requests)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// cleanupLimiters periodically removes unused rate limiters to prevent memory leaks.
// Limiters whose lastSeen timestamp is older than the cleanup interval are evicted.
func (rl *RateLimiter) cleanupLimiters() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-rl.done:
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-rl.cleanup)
			rl.mu.Lock()
			for ip, entry := range rl.limiters {
				if entry.lastSeen.Before(cutoff) {
					delete(rl.limiters, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Close stops the background cleanup goroutine. It is safe to call once.
func (rl *RateLimiter) Close() {
	close(rl.done)
}

// Size returns the number of active rate limiters.
// This is primarily useful for testing and monitoring.
func (rl *RateLimiter) Size() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.limiters)
}

// Clear removes all rate limiters.
// This should only be used in testing.
func (rl *RateLimiter) Clear() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limiters = make(map[string]*ipLimiterEntry)
}
