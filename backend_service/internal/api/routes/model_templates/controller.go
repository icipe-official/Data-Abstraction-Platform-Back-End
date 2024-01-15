package modeltemplates

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

	router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		var DeleteModelTemplate modeltemplates

		if modelTemplateId, err := lib.GetUUID(chi.URLParam(r, "id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			DeleteModelTemplate.ModelTemplateID = modelTemplateId
		}

		DeleteModelTemplate.CurrentUser = lib.CtxGetCurrentUser(r)

		if modelTemplatesAffected, err := DeleteModelTemplate.deleteModelTemplate(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ ModelTemplatesAffected int64 }{ModelTemplatesAffected: modelTemplatesAffected}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("In delete modeltemplate %v by %v, %v were affected", DeleteModelTemplate.ModelTemplateID, DeleteModelTemplate.CurrentUser.DirectoryID, modelTemplatesAffected))
		}
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var RetrieveModelTemplates modeltemplates

		RetrieveModelTemplates.CurrentUser = lib.CtxGetCurrentUser(r)
		if templateId, err := lib.GetUUID(r.URL.Query().Get("id")); err == nil {
			RetrieveModelTemplates.ModelTemplateID = templateId
			if err := RetrieveModelTemplates.getModelTemplates(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveModelTemplates.ModelTemplateDirectoryProject, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get modeltemplate %v by %v successful", templateId, RetrieveModelTemplates.CurrentUser.DirectoryID))
			}
		} else {
			if sq := r.URL.Query().Get("sq"); sq != "" {
				RetrieveModelTemplates.SearchQuery = sq
			}
			if cogt := r.URL.Query().Get("cogt"); cogt != "" {
				RetrieveModelTemplates.CreatedOnGreaterThan = cogt
			}
			if colt := r.URL.Query().Get("colt"); colt != "" {
				RetrieveModelTemplates.CreatedOnLessThan = colt
			}
			if lugt := r.URL.Query().Get("lugt"); lugt != "" {
				RetrieveModelTemplates.LastUpdatedOnGreaterThan = lugt
			}
			if lult := r.URL.Query().Get("lult"); lult != "" {
				RetrieveModelTemplates.LastUpdatedOnLessThan = lult
			}
			if projectId, err := lib.GetUUID(r.URL.Query().Get("pid")); err != nil {
				RetrieveModelTemplates.ProjectID = projectId
			}
			if ia := r.URL.Query().Get("cpv"); ia == "true" || ia == "false" {
				RetrieveModelTemplates.CanPublicView = ia
			}
			if limit, err := strconv.Atoi(r.URL.Query().Get("l")); err == nil {
				RetrieveModelTemplates.Limit = limit
			}
			if offset, err := strconv.Atoi(r.URL.Query().Get("o")); err == nil {
				RetrieveModelTemplates.Offset = offset
			}
			if qs := r.URL.Query().Get("qs"); qs == "true" || qs == "false" {
				RetrieveModelTemplates.QuickSearch = qs
			}
			if sb := r.URL.Query().Get("sb"); sb != "" {
				RetrieveModelTemplates.SortyBy = sb
			}
			if sbo := r.URL.Query().Get("sbo"); sbo != "" {
				RetrieveModelTemplates.SortByOrder = sbo
			}
			if err := RetrieveModelTemplates.getModelTemplates(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveModelTemplates.ModelTemplatesDirectoryProject, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get modeltemplates by %v successful", RetrieveModelTemplates.CurrentUser.DirectoryID))
			}
		}
	})

	router.Put("/{project_id}/{template_id}", func(w http.ResponseWriter, r *http.Request) {
		var UpdateModelTemplate modeltemplates

		if err := json.NewDecoder(r.Body).Decode(&UpdateModelTemplate.ModelTemplateUpdate); err != nil || len(UpdateModelTemplate.ModelTemplateUpdate.Columns) > 6 {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		if projectId, err := lib.GetUUID(chi.URLParam(r, "project_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			UpdateModelTemplate.ProjectID = projectId
		}

		if templateId, err := lib.GetUUID(chi.URLParam(r, "template_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			UpdateModelTemplate.ModelTemplateID = templateId
		}

		UpdateModelTemplate.CurrentUser = lib.CtxGetCurrentUser(r)
		if err := UpdateModelTemplate.updateModelTemplate(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct {
				ID            uuid.UUID
				LastUpdatedOn time.Time
			}{ID: UpdateModelTemplate.ModelTemplateUpdate.ModelTemplate.ID, LastUpdatedOn: UpdateModelTemplate.ModelTemplateUpdate.ModelTemplate.LastUpdatedOn}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("Modeltemplate %v updated by %v", UpdateModelTemplate.ModelTemplateUpdate.ModelTemplate.ID, UpdateModelTemplate.CurrentUser.DirectoryID))
		}
	})

	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var NewModelTemplate modeltemplates

		if err := json.NewDecoder(r.Body).Decode(&NewModelTemplate.ModelTemplate); err != nil || (model.ModelTemplates{} == NewModelTemplate.ModelTemplate) {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		NewModelTemplate.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(false, NewModelTemplate.ModelTemplate.ProjectID, []string{lib.ROLE_MODEL_TEMPLATES_CREATOR}, NewModelTemplate.CurrentUser, w) {
			return
		}

		if err := NewModelTemplate.createModelTemplate(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ ID uuid.UUID }{ID: NewModelTemplate.ModelTemplate.ID}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("ModelTemplate %v created by %v", NewModelTemplate.ModelTemplate.ID, NewModelTemplate.CurrentUser.DirectoryID))
		}
	})

	return router
}
