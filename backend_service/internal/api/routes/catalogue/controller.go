package catalogue

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

	router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		var DeleteCatalogue catalogue

		if catalogueId, err := lib.GetUUID(chi.URLParam(r, "id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			DeleteCatalogue.CatalogueID = catalogueId
		}

		DeleteCatalogue.CurrentUser = lib.CtxGetCurrentUser(r)

		if projectId, err := lib.GetUUID(r.URL.Query().Get("pid")); err == nil {
			DeleteCatalogue.ProjectID = projectId
		}

		if cataloguesAffected, err := DeleteCatalogue.deleteCatalogue(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ CataloguesAffected int64 }{CataloguesAffected: cataloguesAffected}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("In delete catalogue %v by %v, %v were affected", DeleteCatalogue.CatalogueID, DeleteCatalogue.CurrentUser.DirectoryID, cataloguesAffected))
		}
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var RetrieveCatalogue catalogue

		RetrieveCatalogue.CurrentUser = lib.CtxGetCurrentUser(r)
		if catalogueId, err := lib.GetUUID(r.URL.Query().Get("id")); err == nil {
			RetrieveCatalogue.CatalogueID = catalogueId
			if err := RetrieveCatalogue.getCatalogue(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveCatalogue.CatalogueDirectoryProject, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get catalogue %v by %v successful", catalogueId, RetrieveCatalogue.CurrentUser.DirectoryID))
			}
		} else {
			if sq := r.URL.Query().Get("sq"); sq != "" {
				RetrieveCatalogue.SearchQuery = sq
			}
			if cogt := r.URL.Query().Get("cogt"); cogt != "" {
				RetrieveCatalogue.CreatedOnGreaterThan = cogt
			}
			if colt := r.URL.Query().Get("colt"); colt != "" {
				RetrieveCatalogue.CreatedOnLessThan = colt
			}
			if lugt := r.URL.Query().Get("lugt"); lugt != "" {
				RetrieveCatalogue.LastUpdatedOnGreaterThan = lugt
			}
			if lult := r.URL.Query().Get("lult"); lult != "" {
				RetrieveCatalogue.LastUpdatedOnLessThan = lult
			}
			if projectId, err := lib.GetUUID(r.URL.Query().Get("pid")); err != nil {
				RetrieveCatalogue.ProjectID = projectId
			}
			if ia := r.URL.Query().Get("cpv"); ia == "true" || ia == "false" {
				RetrieveCatalogue.CanPublicView = ia
			}
			if limit, err := strconv.Atoi(r.URL.Query().Get("l")); err == nil {
				RetrieveCatalogue.Limit = limit
			}
			if offset, err := strconv.Atoi(r.URL.Query().Get("o")); err == nil {
				RetrieveCatalogue.Offset = offset
			}
			if qs := r.URL.Query().Get("qs"); qs == "true" || qs == "false" {
				RetrieveCatalogue.QuickSearch = qs
			}
			if sb := r.URL.Query().Get("sb"); sb != "" {
				RetrieveCatalogue.SortyBy = sb
			}
			if sbo := r.URL.Query().Get("sbo"); sbo != "" {
				RetrieveCatalogue.SortByOrder = sbo
			}
			if err := RetrieveCatalogue.getCatalogue(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveCatalogue.CataloguesDirectoryProject, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get catalogues by %v successful", RetrieveCatalogue.CurrentUser.DirectoryID))
			}
		}
	})

	router.Put("/{project_id}/{catalogue_id}", func(w http.ResponseWriter, r *http.Request) {
		var UpdateCatalogue catalogue

		if err := json.NewDecoder(r.Body).Decode(&UpdateCatalogue.CatalogueUpdate); err != nil || len(UpdateCatalogue.CatalogueUpdate.Columns) > 4 {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		if projectId, err := lib.GetUUID(chi.URLParam(r, "project_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			UpdateCatalogue.ProjectID = projectId
		}

		if catalogueId, err := lib.GetUUID(chi.URLParam(r, "catalogue_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			UpdateCatalogue.CatalogueID = catalogueId
		}

		UpdateCatalogue.CurrentUser = lib.CtxGetCurrentUser(r)
		if err := UpdateCatalogue.updateCatalogue(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct {
				ID            uuid.UUID
				LastUpdatedOn time.Time
			}{ID: UpdateCatalogue.CatalogueUpdate.Catalogue.ID, LastUpdatedOn: UpdateCatalogue.CatalogueUpdate.Catalogue.LastUpdatedOn}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("Catalogue %v updated by %v", UpdateCatalogue.CatalogueUpdate.Catalogue.ID, UpdateCatalogue.CurrentUser.DirectoryID))
		}
	})

	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var NewCatalogue catalogue

		if err := json.NewDecoder(r.Body).Decode(&NewCatalogue.Catalogue); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		NewCatalogue.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(false, NewCatalogue.Catalogue.ProjectID, []string{lib.ROLE_CATALOGUE_CREATOR}, NewCatalogue.CurrentUser, w) {
			return
		}

		if err := NewCatalogue.createCatalogue(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ ID uuid.UUID }{ID: NewCatalogue.Catalogue.ID}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("Catalogue %v created by %v", NewCatalogue.Catalogue.ID, NewCatalogue.CurrentUser.DirectoryID))
		}
	})

	return router
}
