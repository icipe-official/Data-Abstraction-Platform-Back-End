package projects

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func Router() *chi.Mux {
	router := chi.NewRouter()

	router.Delete("/roles", func(w http.ResponseWriter, r *http.Request) {
		var DeleteRoles projects

		if err := json.NewDecoder(r.Body).Decode(&DeleteRoles.Roles); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		DeleteRoles.CurrentUser = lib.CtxGetCurrentUser(r)

		if !lib.IsUserAuthorized(false, DeleteRoles.Roles.ProjectID, []string{lib.ROLE_PROJECT_ADMIN}, DeleteRoles.CurrentUser, w) {
			return
		}

		if rolesDeleted, err := DeleteRoles.deleteRoles(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ RolesDeleted int }{RolesDeleted: rolesDeleted}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("%v roles deleted for %v in %v by %v", rolesDeleted, DeleteRoles.Roles.DirectoryID, DeleteRoles.Roles.ProjectID, DeleteRoles.CurrentUser.DirectoryID))
		}
	})

	router.Post("/roles", func(w http.ResponseWriter, r *http.Request) {
		var NewRoles projects

		if err := json.NewDecoder(r.Body).Decode(&NewRoles.Roles); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		NewRoles.CurrentUser = lib.CtxGetCurrentUser(r)

		if !lib.IsUserAuthorized(false, NewRoles.Roles.ProjectID, []string{lib.ROLE_PROJECT_ADMIN}, NewRoles.CurrentUser, w) {
			return
		}

		if noOfRolesCreated, err := NewRoles.createRoles(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ RolesAssigned int64 }{RolesAssigned: noOfRolesCreated}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("%v roles created for %v in project %v by %v", noOfRolesCreated, NewRoles.Roles.DirectoryID, NewRoles.Roles.ProjectID, NewRoles.CurrentUser.DirectoryID))
		}
	})

	router.Delete("/{project_id}", func(w http.ResponseWriter, r *http.Request) {
		var DeleteProject projects

		if projectId, err := lib.GetUUID(chi.URLParam(r, "project_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			DeleteProject.ProjectID = projectId
		}

		DeleteProject.CurrentUser = lib.CtxGetCurrentUser(r)

		if projectsAffected, err := DeleteProject.deleteProject(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ ProjectsAffected int64 }{ProjectsAffected: projectsAffected}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("In delete/deactivate project %v by %v, %v were affected", DeleteProject.ProjectID, DeleteProject.CurrentUser.DirectoryID, projectsAffected))
		}
	})

	router.Put("/{project_id}", func(w http.ResponseWriter, r *http.Request) {
		var UpdateProject projects

		if projectId, err := lib.GetUUID(chi.URLParam(r, "project_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			UpdateProject.ProjectID = projectId
		}

		UpdateProject.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(true, UpdateProject.ProjectID, []string{lib.ROLE_PROJECT_ADMIN}, UpdateProject.CurrentUser, w) {
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&UpdateProject.ProjectUpdate); err != nil || len(UpdateProject.ProjectUpdate.Columns) > 2 {
			lib.SendJsonResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		if err := UpdateProject.updateProject(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct {
				ID            uuid.UUID
				LastUpdatedOn time.Time
			}{ID: UpdateProject.ProjectUpdate.Project.ID, LastUpdatedOn: UpdateProject.ProjectUpdate.Project.LastUpdatedOn}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("Project %v updated by %v", UpdateProject.ProjectUpdate.Project.ID, UpdateProject.CurrentUser.DirectoryID))
		}
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var RetrieveProjects projects
		RetrieveProjects.CurrentUser = lib.CtxGetCurrentUser(r)
		if projectId, err := lib.GetUUID(r.URL.Query().Get("id")); err == nil {
			RetrieveProjects.ProjectID = projectId
			if err := RetrieveProjects.getProjects(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveProjects.RetrieveProject, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get project %v by %v successful", projectId, RetrieveProjects.CurrentUser.DirectoryID))
			}
		} else {
			if sq := r.URL.Query().Get("sq"); sq != "" {
				RetrieveProjects.SearchQuery = sq
			}
			if cogt := r.URL.Query().Get("cogt"); cogt != "" {
				RetrieveProjects.CreatedOnGreaterThan = cogt
			}
			if colt := r.URL.Query().Get("colt"); colt != "" {
				RetrieveProjects.CreatedOnLessThan = colt
			}
			if lugt := r.URL.Query().Get("lugt"); lugt != "" {
				RetrieveProjects.LastUpdatedOnGreaterThan = lugt
			}
			if lult := r.URL.Query().Get("lult"); lult != "" {
				RetrieveProjects.LastUpdatedOnLessThan = lult
			}
			if ia := r.URL.Query().Get("ia"); ia == "true" || ia == "false" {
				RetrieveProjects.IsActive = ia
			}
			if limit, err := strconv.Atoi(r.URL.Query().Get("l")); err == nil {
				RetrieveProjects.Limit = limit
			}
			if offset, err := strconv.Atoi(r.URL.Query().Get("o")); err == nil {
				RetrieveProjects.Offset = offset
			}
			if qs := r.URL.Query().Get("qs"); qs == "true" || qs == "false" {
				RetrieveProjects.QuickSearch = qs
			}
			if err := RetrieveProjects.getProjects(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveProjects.RetrieveProjects, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get projects by %v successful", RetrieveProjects.CurrentUser.DirectoryID))
			}
		}
	})

	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var NewProject projects
		NewProject.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, NewProject.CurrentUser, w) {
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&NewProject.Project); err != nil || (model.Projects{} == NewProject.Project) {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		if err := NewProject.createProject(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ ID string }{NewProject.Project.ID.String()}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("Project %v created by %v", NewProject.Project.ID, NewProject.CurrentUser.DirectoryID))
		}
	})

	return router
}
