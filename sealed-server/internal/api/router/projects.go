package router

import (
	"net/http"

	"github.com/souravlayek/sealed-server/internal/api/handlers"
)

func CreateProjectsRoutes() *http.ServeMux {
	router := NewCustomRouter()

	router.HandleFunc("POST /{$}", handlers.CreateProject)
	router.HandleFunc("GET /{$}", handlers.GetAllProjects)
	router.HandleFunc("GET /{slug}", handlers.GetProjectDetails)
	router.HandleFunc("DELETE /{id}", handlers.DeleteProject)
	router.HandleFunc("PUT /{id}", handlers.UpdateProjectName)
	router.HandleFunc("POST /{id}/invite-member", handlers.InviteProjectMember)
	router.HandleFunc("POST /{id}/secret", handlers.CreateProjectSecret)
	router.HandleFunc("PUT /{id}/secret/{secretId}", handlers.UpdateProjectSecret)
	router.HandleFunc("DELETE /{id}/secret/{environment}/{secretKey}", handlers.DeleteProjectSecret)
	router.HandleFunc("PUT /{id}/environment", handlers.AddEnvironmentToProject)
	router.HandleFunc("DELETE /{id}/environment/{environment}", handlers.DeleteEnvironment)

	return &router.ServeMux
}
