package directory

import (
	"data_administration_platform/internal/api/lib"
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

	router.Delete("/{directory_id}", func(w http.ResponseWriter, r *http.Request) {
		var UserDelete directory
		UserDelete.CurrentUser = lib.CtxGetCurrentUser(r)

		if directoryId, err := lib.GetUUID(chi.URLParam(r, "directory_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			UserDelete.DirectoryID = directoryId
		}

		if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, UserDelete.CurrentUser, w) {
			return
		}

		if err := UserDelete.deleteUser(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ UsersAffected int64 }{UsersAffected: 1}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("User %v deleted/deactivated by %v", UserDelete.DirectoryID, UserDelete.CurrentUser.DirectoryID))
		}
	})

	router.Put("/{directory_id}", func(w http.ResponseWriter, r *http.Request) {
		var UserUpdate directory
		UserUpdate.CurrentUser = lib.CtxGetCurrentUser(r)

		if directoryId, err := lib.GetUUID(chi.URLParam(r, "directory_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			UserUpdate.DirectoryID = directoryId
		}

		if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, UserUpdate.CurrentUser, nil) {
			if projectId, err := lib.GetUUID(r.URL.Query().Get("pid")); err != nil {
				lib.SendErrorResponse(lib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden)), w)
				return
			} else {
				UserUpdate.ProjectID = projectId
			}
		}

		_ = json.NewDecoder(r.Body).Decode(&UserUpdate.DirectoryUpdate)

		if err := UserUpdate.updateUser(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct {
				ID            uuid.UUID
				LastUpdatedOn time.Time
			}{ID: UserUpdate.DirectoryID, LastUpdatedOn: time.Now()}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("User %v updated by %v", UserUpdate.DirectoryID, UserUpdate.CurrentUser.DirectoryID))
		}
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var RetrieveUsers directory
		RetrieveUsers.CurrentUser = lib.CtxGetCurrentUser(r)
		if directoryId, err := lib.GetUUID(r.URL.Query().Get("id")); err == nil {
			RetrieveUsers.DirectoryID = directoryId
			if err := RetrieveUsers.getUsers(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveUsers.RetrieveUser, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get user %v by %v successful", directoryId, RetrieveUsers.CurrentUser.DirectoryID))
			}
		} else {
			if sq := r.URL.Query().Get("sq"); sq != "" {
				RetrieveUsers.SearchQuery = sq
			}
			if cogt := r.URL.Query().Get("cogt"); cogt != "" {
				RetrieveUsers.CreatedOnGreaterThan = cogt
			}
			if colt := r.URL.Query().Get("colt"); colt != "" {
				RetrieveUsers.CreatedOnLessThan = colt
			}
			if lugt := r.URL.Query().Get("lugt"); lugt != "" {
				RetrieveUsers.LastUpdatedOnGreaterThan = lugt
			}
			if lult := r.URL.Query().Get("lult"); lult != "" {
				RetrieveUsers.LastUpdatedOnLessThan = lult
			}
			if pid, err := lib.GetUUID(r.URL.Query().Get("pid")); err == nil {
				RetrieveUsers.ProjectID = pid
			}
			if prole := r.URL.Query().Get("prole"); prole != "" {
				RetrieveUsers.ProjectRole = prole
			}
			if limit, err := strconv.Atoi(r.URL.Query().Get("l")); err == nil {
				RetrieveUsers.Limit = limit
			}
			if offset, err := strconv.Atoi(r.URL.Query().Get("o")); err == nil {
				RetrieveUsers.Offset = offset
			}
			if qs := r.URL.Query().Get("qs"); qs == "true" || qs == "false" {
				RetrieveUsers.QuickSearch = qs
			}
			if sb := r.URL.Query().Get("sb"); sb != "" {
				RetrieveUsers.SortyBy = sb
			}
			if sbo := r.URL.Query().Get("sbo"); sbo != "" {
				RetrieveUsers.SortByOrder = sbo
			}
			if err := RetrieveUsers.getUsers(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveUsers.RetrieveUsers, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get users by %v successful", RetrieveUsers.CurrentUser.DirectoryID))
			}
		}
	})

	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var NewUser directory

		if err := json.NewDecoder(r.Body).Decode(&NewUser.DirectoryCreate); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		NewUser.CurrentUser = lib.CtxGetCurrentUser(r)

		if NewUser.DirectoryCreate.ProjectID == uuid.Nil || NewUser.DirectoryCreate.IsSystemUser {
			if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, NewUser.CurrentUser, w) {
				return
			}
		} else {
			if !lib.IsUserAuthorized(false, NewUser.DirectoryCreate.ProjectID, []string{lib.ROLE_PROJECT_ADMIN}, NewUser.CurrentUser, w) {
				return
			}
		}

		if err := NewUser.createUser(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ ID string }{NewUser.Directory.ID.String()}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("User %v created by %v", NewUser.Directory.ID, NewUser.CurrentUser.DirectoryID))
		}
	})

	return router
}
