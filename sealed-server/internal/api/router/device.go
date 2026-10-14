package router

import (
	"net/http"

	"github.com/souravlayek/sealed-server/internal/api/handlers"
	"github.com/souravlayek/sealed-server/internal/middleware"
)

func CreateDeviceRoutes() *http.ServeMux {
	router := NewCustomRouter()
	router.HandlerFuncWithMiddleware("GET /{$}", []middleware.Middleware{middleware.RequireAuth}, handlers.GetAllDevices)
	router.HandleFunc("POST /code", handlers.CreateDeviceCode)
	router.HandleFunc("POST /token", handlers.ValidateToken)
	router.HandlerFuncWithMiddleware("POST /validate", []middleware.Middleware{middleware.RequireAuth}, handlers.ValidateAuthentication)
	return &router.ServeMux
}
