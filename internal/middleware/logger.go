package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (r *ResponseRecorder) WriteHeader(status int) {
	r.StatusCode = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	r.Body = b
	return r.ResponseWriter.Write(b)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rc := &ResponseRecorder{
			ResponseWriter: w,
			StatusCode:     http.StatusOK}
		next.ServeHTTP(rc, r)
		duration := time.Since(startTime)
		logger := log.Info()
		if rc.StatusCode != http.StatusOK {
			logger = log.Error().Bytes("body", rc.Body)
		}

		logger.Str("protocol", "http").
			Str("method", r.Method).
			Str("path", r.RequestURI).
			Int("status_code", rc.StatusCode).
			Str("status_text", http.StatusText(rc.StatusCode)).
			Dur("duration", duration).
			Str("ip", r.RemoteAddr).
			Msg("received a HTTP request")
	})
}
