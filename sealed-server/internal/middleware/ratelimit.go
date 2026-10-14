package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
	// "github.com/souravlayek/sealed-server/pkg/logger"
)

// RateLimiter struct (per client)
type RateLimiter struct {
	mu        sync.Mutex
	tokens    int
	maxTokens int
	ticker    *time.Ticker
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:    maxTokens,
		maxTokens: maxTokens,
		ticker:    time.NewTicker(refillRate),
	}

	// Background goroutine to refill tokens
	go func() {
		for range rl.ticker.C {
			rl.mu.Lock()
			if rl.tokens < rl.maxTokens {
				rl.tokens++
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

// Stop stops the ticker when not needed
func (rl *RateLimiter) Stop() {
	rl.ticker.Stop()
}

var limit = 10
var refill = 1 * time.Second

// RateLimiterMiddleware applies rate limiting per client IP
func RateLimiterMiddleware(next http.Handler) http.Handler {
	clientLimits := make(map[string]*RateLimiter)
	var mu sync.Mutex

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		// Get or create rate limiter for this IP
		mu.Lock()
		limiter, exists := clientLimits[clientIP]
		if !exists {
			limiter = NewRateLimiter(limit, refill)
			clientLimits[clientIP] = limiter
		}
		mu.Unlock()

		// Check if request is allowed
		if !limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			// logger.Logger.Errorw("Rate limit exceeded", "client_ip", clientIP)
			return
		}

		// Forward request to the next handler
		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the IP address from request
func getClientIP(r *http.Request) string {
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		ip = strings.TrimSpace(ips[0])
	}
	return ip
}
