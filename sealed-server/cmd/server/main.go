package main

import (
	"expvar"
	"net/http"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/souravlayek/sealed-server/docs"
	"github.com/souravlayek/sealed-server/internal/api/router"
	"github.com/souravlayek/sealed-server/utils"

	"github.com/souravlayek/sealed-server/internal/database"

	"github.com/souravlayek/sealed-server/internal/middleware"
	"github.com/souravlayek/sealed-server/pkg/auth"
	"github.com/souravlayek/sealed-server/pkg/logger"
)

// @title						Sealed Server API
// @version					1.0
// @description				This is a server to manage Sealed resources.
// @termsOfService				http://swagger.io/terms/
//
// @contact.name				API Support
// @contact.url				https://souravlayek.com/contact-me
// @contact.email				souravlayek11@gmail.com
//
// @host						localhost:8080
// @BasePath					/api
// @schemes					http
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
//
// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	loglerInstance := logger.InitLogger()
	defer loglerInstance.Sync()

	if err := utils.LoadEnv(); err != nil {
		logger.Logger.Fatal(err)
	}
	auth.InitJWT()
	conn := database.ConnectDB()
	defer conn.Close()

	apiRouter := http.NewServeMux()
	nestedRouter := router.CreateRoutes()
	genericMiddlewareStack := middleware.CreateStack(
		middleware.Logger,
		middleware.RateLimiterMiddleware,
		middleware.RealIP,
	)

	apiRouter.Handle("/api/", http.StripPrefix("/api", nestedRouter))
	apiRouter.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecContent: docs.SwaggerInfo.ReadDoc(),
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Sealed Server API",
			},
			DarkMode: true,
		})

		if err != nil {
			logger.Logger.Errorf("%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(htmlContent))
	})
	apiRouter.Handle("GET /debug/vars", expvar.Handler())

	server := http.Server{
		Addr:    ":8080",
		Handler: genericMiddlewareStack(apiRouter),
	}
	logger.Logger.Infow("Server is running on port 8080")
	logger.Logger.Fatal(server.ListenAndServe())
}
