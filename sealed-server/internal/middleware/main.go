package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func CreateStack(middleware ...Middleware) Middleware {
	length := len(middleware)
	if length == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return func(next http.Handler) http.Handler {
		for i := length - 1; i >= 0; i-- {
			next = middleware[i](next)
		}
		return next
	}
}

var allowedOrigins = map[string]bool{
	"http://localhost:5173":              true,
	"https://sealed.souravlayek.com":     true,
	"https://api-sealed.souravlayek.com": true,
}

// CORSMiddleware handles all preflight OPTIONS requests globally
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin) // Adjust in production
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		// If it's a preflight OPTIONS request, respond immediately with 200 OK
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}
