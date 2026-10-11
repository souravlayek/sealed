package router

import (
	"net/http"

	"github.com/souravlayek/sealed-server/internal/api/handlers"
	"github.com/souravlayek/sealed-server/internal/middleware"
)

func CreateAuthRoutes() *http.ServeMux {
	authRouter := NewCustomRouter()
	authRouter.HandleFunc("POST /register", handlers.RegisterUser)
	authRouter.HandleFunc("POST /signin", handlers.SigninUser)
	authRouter.HandleFunc("POST /revalidate-token", handlers.GetAccessToken)
	authRouter.HandlerFuncWithMiddleware("GET /user", []middleware.Middleware{middleware.RequireAuth}, handlers.GetUser)

	return &authRouter.ServeMux
}
