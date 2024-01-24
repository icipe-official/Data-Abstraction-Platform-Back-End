package platformstatistics

import (
	"data_administration_platform/internal/api/lib"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() *chi.Mux {
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Use(lib.AuthenticationMiddleware)

		r.Get("/projects", func(w http.ResponseWriter, r *http.Request) {
			var RetrieveProjectsStats platformstats

			RetrieveProjectsStats.CurrentUser = lib.CtxGetCurrentUser(r)
			if projectId, err := lib.GetUUID(r.URL.Query().Get("id")); err == nil {
				RetrieveProjectsStats.ProjectID = projectId
				if err := RetrieveProjectsStats.getProjectStats(); err != nil {
					lib.SendErrorResponse(err, w)
				} else {
					lib.SendJsonResponse(RetrieveProjectsStats.ProjectStats, w)
					intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get stats for project %v by %v successful", projectId, RetrieveProjectsStats.CurrentUser.DirectoryID))
				}
			} else {
				if err := RetrieveProjectsStats.getProjectStats(); err != nil {
					lib.SendErrorResponse(err, w)
				} else {
					lib.SendJsonResponse(RetrieveProjectsStats.ProjectsStats, w)
					intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get stats for projects by %v successful", RetrieveProjectsStats.CurrentUser.DirectoryID))
				}
			}
		})
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var RetrievePlatformStats platformstats

		if err := RetrievePlatformStats.getPlatfromStats(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(RetrievePlatformStats.PlatformStats, w)
			intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, "Get platform stats successful")
		}
	})

	return router
}
