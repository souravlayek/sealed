package router

import (
	"net/http"

	"github.com/souravlayek/sealed-server/internal/api/handlers"
)

func CreateProjectsRoutes() *http.ServeMux {
	router := NewCustomRouter()

	router.HandleFunc("POST /{$}", handlers.CreateProject)
	router.HandleFunc("GET /{$}", handlers.GetAllProjects)
	router.HandleFunc("GET /{id}", handlers.GetProjectDetails)
	router.HandleFunc("POST /{id}/invite-member", handlers.InviteProjectMember)
	router.HandleFunc("POST /{id}/secret", handlers.CreateProjectSecret)
	router.HandleFunc("PUT /{id}/environment", handlers.AddEnvironmentToProject)

	return &router.ServeMux
}
