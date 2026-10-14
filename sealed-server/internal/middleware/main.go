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
