package router

import (
	"net/http"

	"github.com/souravlayek/sealed-server/internal/api/handlers"
	"github.com/souravlayek/sealed-server/internal/middleware"
)

type MyCustomRouter struct {
	http.ServeMux
}

func NewCustomRouter() *MyCustomRouter {
	return &MyCustomRouter{}
}

func (mr *MyCustomRouter) HandlerFuncWithMiddleware(pattern string, mdw []middleware.Middleware, handlerFunc http.HandlerFunc) {
	stackedMiddleWare := middleware.CreateStack(mdw...)
	mr.Handle(pattern, stackedMiddleWare(handlerFunc))
}

func CreateRoutes() *MyCustomRouter {
	router := NewCustomRouter()
	v1Router := NewCustomRouter()
	router.Handle("/v1/", http.StripPrefix("/v1", &v1Router.ServeMux))
	// Add your routes
	v1Router.Handle("/auth/", http.StripPrefix("/auth", CreateAuthRoutes()))
	protectedRoutes := middleware.CreateStack(middleware.RequireAuth)
	v1Router.HandlerFuncWithMiddleware("GET /projects/{projectSlug}/{environment}/secrets", []middleware.Middleware{middleware.ValidateDeviceRequest}, handlers.GetAllProjectSecretsForEnv)

	v1Router.Handle("/projects/", http.StripPrefix("/projects", protectedRoutes(CreateProjectsRoutes())))
	v1Router.Handle("/device/", http.StripPrefix("/device", CreateDeviceRoutes()))

	return router
}
