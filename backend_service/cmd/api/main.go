package main

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/api/routes/abstractions"
	"data_administration_platform/internal/api/routes/catalogue"
	"data_administration_platform/internal/api/routes/directory"
	"data_administration_platform/internal/api/routes/iam"
	modeltemplates "data_administration_platform/internal/api/routes/model_templates"
	platformstatistics "data_administration_platform/internal/api/routes/platform_statistics"
	"data_administration_platform/internal/api/routes/projects"
	"data_administration_platform/internal/api/routes/storage"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {
	verifyEnvSetup()

	router := chi.NewRouter()

	BASE_PATH := os.Getenv("BASE_PATH")
	if !strings.HasSuffix(BASE_PATH, "/") {
		BASE_PATH = BASE_PATH + "/"
	}

	router.Use(middleware.Heartbeat(BASE_PATH))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{os.Getenv("DOMAIN_URL")},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
	}))

	router.Route(BASE_PATH+lib.API_VERSION, func(r chi.Router) {
		r.Mount("/platformstats", platformstatistics.Router())
		r.Mount("/iam", iam.Router())
		r.Group(func(authedRoutes chi.Router) {
			authedRoutes.Use(lib.AuthenticationMiddleware)
			authedRoutes.Mount("/projects", projects.Router())
			authedRoutes.Mount("/directory", directory.Router())
			authedRoutes.Mount("/storage", storage.Router())
			authedRoutes.Mount("/catalogue", catalogue.Router())
			authedRoutes.Mount("/modeltemplate", modeltemplates.Router())
			authedRoutes.Mount("/abstraction", abstractions.Router())
		})
	})

	log.Printf("Server will be listening on port: %v", os.Getenv("PORT"))
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), router); err != nil {
		intpkglib.Log(intpkglib.LOG_FATAL, "Startup", fmt.Sprintf("Could not start server | reason: %v", err))
	}
}

func verifyEnvSetup() {
	envVariablesRequired := []string{
		"DOMAIN_URL",
		"PORT",
		"PSQL_DBNAME",
		"PSQL_HOST",
		"PSQL_PORT",
		"PSQL_USER",
		"PSQL_PASS",
		"PSQL_SSLMODE",
		"PSQL_SCHEMA",
		"PSQL_DATABASE_DRIVE_NAME",
		"REDIS_HOST",
		"REDIS_PORT",
		"ACCESS_REFRESH_TOKEN",
		"ENCRYPTION_KEY",
		"TMP_DIR",
	}
	envVariablesMissing := []string{}
	for _, evr := range envVariablesRequired {
		if os.Getenv(evr) == "" {
			envVariablesMissing = append(envVariablesMissing, evr)
		}
	}
	if len(envVariablesMissing) > 0 {
		intpkglib.Log(intpkglib.LOG_FATAL, "Startup", fmt.Sprintf("Following env variables not set: %+q", envVariablesMissing))
	}
}
