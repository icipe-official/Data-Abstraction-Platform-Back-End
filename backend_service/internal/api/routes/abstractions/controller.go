package abstractions

import (
	"data_administration_platform/internal/api/lib"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func Router() *chi.Mux {
	router := chi.NewRouter()

	router.Post("/review/{project_id}", func(w http.ResponseWriter, r *http.Request) {
		var ReviewAbstraction abstractions

		if projectId, err := lib.GetUUID(chi.URLParam(r, "project_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			ReviewAbstraction.ProjectID = projectId
		}

		ReviewAbstraction.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(false, ReviewAbstraction.ProjectID, []string{lib.ROLE_REVIEWER}, ReviewAbstraction.CurrentUser, w) {
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&ReviewAbstraction.AbstractionReview); err != nil || (abstractionReview{} == ReviewAbstraction.AbstractionReview) {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		if err := ReviewAbstraction.reviewAbstraction(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct {
				ID            uuid.UUID
				LastUpdatedOn time.Time
			}{ID: ReviewAbstraction.AbstractionReview.DirectoryID, LastUpdatedOn: ReviewAbstraction.AbstractionReview.LastUpdatedOn}, w)
			intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Abtraction %v reviewed by %v", ReviewAbstraction.AbstractionReview.AbstractionID, ReviewAbstraction.AbstractionReview.DirectoryID))
		}
	})

	router.Post("/filefromdata/{project_id}", func(w http.ResponseWriter, r *http.Request) {
		var FileFromAbstractionData abstractions

		if projectId, err := lib.GetUUID(chi.URLParam(r, "project_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			FileFromAbstractionData.ProjectID = projectId
		}
		if modelTemplateId, err := lib.GetUUID(r.URL.Query().Get("mtid")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			FileFromAbstractionData.ModelTemplateID = modelTemplateId
		}

		if err := json.NewDecoder(r.Body).Decode(&FileFromAbstractionData.FileFromAbstractionData); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			if err := json.Unmarshal([]byte(FileFromAbstractionData.FileFromAbstractionData.ModelTemplate), &FileFromAbstractionData.FileFromData.ModelTemplate); err != nil {
				lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
				return
			}
		}

		FileFromAbstractionData.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(false, FileFromAbstractionData.ProjectID, []string{}, FileFromAbstractionData.CurrentUser, w) {
			return
		}

		if cogt := r.URL.Query().Get("cogt"); cogt != "" {
			FileFromAbstractionData.CreatedOnGreaterThan = cogt
		}
		if colt := r.URL.Query().Get("colt"); colt != "" {
			FileFromAbstractionData.CreatedOnLessThan = colt
		}
		if lugt := r.URL.Query().Get("lugt"); lugt != "" {
			FileFromAbstractionData.LastUpdatedOnGreaterThan = lugt
		}
		if lult := r.URL.Query().Get("lult"); lult != "" {
			FileFromAbstractionData.LastUpdatedOnLessThan = lult
		}
		if iv := r.URL.Query().Get("iv"); iv == "true" || iv == "false" {
			FileFromAbstractionData.IsVerified = iv
		}
		if did, err := lib.GetUUID(r.URL.Query().Get("did")); err == nil {
			FileFromAbstractionData.DirectoryID = did
		}
		if sb := r.URL.Query().Get("sb"); sb != "" {
			FileFromAbstractionData.SortyBy = sb
		}
		if sbo := r.URL.Query().Get("sbo"); sbo != "" {
			FileFromAbstractionData.SortByOrder = sbo
		}
		if ft := r.URL.Query().Get("ft"); ft == lib.GEN_FILE_EXCEL || ft == lib.GEN_FILE_CSV {
			FileFromAbstractionData.FileFromData.FileType = ft
		} else {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, "File type not specified"), w)
			return
		}
		if iv := r.URL.Query().Get("ss"); iv == "true" || iv == "false" {
			FileFromAbstractionData.FileFromData.SingleSheet = iv
		}

		if err := FileFromAbstractionData.genFileFromAbstractions(); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Generate file from data failed | reason: %v", err))
			lib.SendErrorResponse(lib.NewError(http.StatusInternalServerError, "Could not generate file from data"), w)
		} else if FileFromAbstractionData.FileFromData.FilePath != "" {
			abstractionFile, err := os.Open(FileFromAbstractionData.FileFromData.FilePath)
			if err != nil {
				intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Open generated file %v by %v failed | reason: %v", FileFromAbstractionData.FileFromData.FilePath, FileFromAbstractionData.CurrentUser.DirectoryID, err))
				lib.SendErrorResponse(lib.NewError(http.StatusInternalServerError, "Could not open generated file from data"), w)
			}
			defer abstractionFile.Close()
			lib.SendFile(w, &FileFromAbstractionData.FileFromData.FileContentType, abstractionFile)
			intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, "File generation from abstracted data successful")
		} else {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
		}
		if FileFromAbstractionData.FileFromData.FilePath != "" {
			if err := os.Remove(FileFromAbstractionData.FileFromData.FilePath); err != nil {
				intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Delete file at %v failed | reason: %v", FileFromAbstractionData.FileFromData.FilePath, err))
			}
		}
	})

	router.Get("/{project_id}", func(w http.ResponseWriter, r *http.Request) {
		var RetrieveAbstractions abstractions

		if projectId, err := lib.GetUUID(chi.URLParam(r, "project_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			RetrieveAbstractions.ProjectID = projectId
		}

		RetrieveAbstractions.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(false, RetrieveAbstractions.ProjectID, []string{}, RetrieveAbstractions.CurrentUser, w) {
			return
		}

		if abstractionId, err := lib.GetUUID(r.URL.Query().Get("id")); err == nil {
			RetrieveAbstractions.AbstractionID = abstractionId
			if err := RetrieveAbstractions.getAbstractions(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveAbstractions.RetrieveAbstraction, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get abstraction %v by %v successful", abstractionId, RetrieveAbstractions.CurrentUser.DirectoryID))
			}
		} else if modelTemplateId, err := lib.GetUUID(r.URL.Query().Get("mtid")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			RetrieveAbstractions.ModelTemplateID = modelTemplateId
			if cogt := r.URL.Query().Get("cogt"); cogt != "" {
				RetrieveAbstractions.CreatedOnGreaterThan = cogt
			}
			if colt := r.URL.Query().Get("colt"); colt != "" {
				RetrieveAbstractions.CreatedOnLessThan = colt
			}
			if lugt := r.URL.Query().Get("lugt"); lugt != "" {
				RetrieveAbstractions.LastUpdatedOnGreaterThan = lugt
			}
			if lult := r.URL.Query().Get("lult"); lult != "" {
				RetrieveAbstractions.LastUpdatedOnLessThan = lult
			}
			if iv := r.URL.Query().Get("iv"); iv == "true" || iv == "false" {
				RetrieveAbstractions.IsVerified = iv
			}
			if did, err := lib.GetUUID(r.URL.Query().Get("did")); err == nil {
				RetrieveAbstractions.DirectoryID = did
			}
			if limit, err := strconv.Atoi(r.URL.Query().Get("l")); err == nil {
				RetrieveAbstractions.Limit = limit
			}
			if offset, err := strconv.Atoi(r.URL.Query().Get("o")); err == nil {
				RetrieveAbstractions.Offset = offset
			}
			if sb := r.URL.Query().Get("sb"); sb != "" {
				RetrieveAbstractions.SortyBy = sb
			}
			if sbo := r.URL.Query().Get("sbo"); sbo != "" {
				RetrieveAbstractions.SortByOrder = sbo
			}
			if err := RetrieveAbstractions.getAbstractions(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveAbstractions.RetrieveAbstractions, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get abstractions by %v successful", RetrieveAbstractions.CurrentUser.DirectoryID))
			}
		}
	})

	router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		var DeleteAbstraction abstractions

		if abstractionId, err := lib.GetUUID(chi.URLParam(r, "id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			DeleteAbstraction.AbstractionID = abstractionId
		}

		if projectId, err := lib.GetUUID(r.URL.Query().Get("pid")); err == nil {
			DeleteAbstraction.ProjectID = projectId
		}

		DeleteAbstraction.CurrentUser = lib.CtxGetCurrentUser(r)

		if abstractionsAffected, err := DeleteAbstraction.deleteAbstraction(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ AbstractionsAffected int64 }{AbstractionsAffected: abstractionsAffected}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("In delete abstraction %v by %v, %v were affected", DeleteAbstraction.AbstractionID, DeleteAbstraction.CurrentUser.DirectoryID, abstractionsAffected))
		}
	})

	router.Put("/{project_id}/{abstraction_id}", func(w http.ResponseWriter, r *http.Request) {
		var UpdateAbstraction abstractions

		if err := json.NewDecoder(r.Body).Decode(&UpdateAbstraction.AbstractionUpdate); err != nil || len(UpdateAbstraction.AbstractionUpdate.Columns) > 2 {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		if projectId, err := lib.GetUUID(chi.URLParam(r, "project_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			UpdateAbstraction.ProjectID = projectId
		}

		if abstractionId, err := lib.GetUUID(chi.URLParam(r, "abstraction_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			UpdateAbstraction.AbstractionID = abstractionId
		}

		UpdateAbstraction.CurrentUser = lib.CtxGetCurrentUser(r)
		if err := UpdateAbstraction.updateAbstraction(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct {
				ID            uuid.UUID
				LastUpdatedOn time.Time
			}{ID: UpdateAbstraction.AbstractionUpdate.Abstraction.ID, LastUpdatedOn: UpdateAbstraction.AbstractionUpdate.Abstraction.LastUpdatedOn}, w)
			intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Abstraction %v updated by %v", UpdateAbstraction.AbstractionUpdate.Abstraction.ID, UpdateAbstraction.CurrentUser.DirectoryID))
		}
	})

	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var CreateAbstractions abstractions

		if err := json.NewDecoder(r.Body).Decode(&CreateAbstractions.AbstractionsCreation); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		CreateAbstractions.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(false, CreateAbstractions.AbstractionsCreation.ProjectID, []string{lib.ROLE_ABSTRACTIONS_ADMIN}, CreateAbstractions.CurrentUser, w) {
			return
		}

		if abstractionsCreated, err := CreateAbstractions.createAbstractions(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct {
				AbstractionsCreated int64
			}{AbstractionsCreated: abstractionsCreated}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("%v abstractions created by %v", abstractionsCreated, CreateAbstractions.CurrentUser.DirectoryID))
		}
	})

	return router
}
