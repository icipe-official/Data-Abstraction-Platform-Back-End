package storage

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

	router.Delete("/file", func(w http.ResponseWriter, r *http.Request) {
		var DeleteFile storage

		if err := json.NewDecoder(r.Body).Decode(&DeleteFile.File); err != nil || (model.Files{} == DeleteFile.File) {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		DeleteFile.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(false, DeleteFile.File.ProjectID, []string{lib.ROLE_PROJECT_ADMIN, lib.ROLE_FILE_CREATOR}, DeleteFile.CurrentUser, w) {
			return
		}

		if err := DeleteFile.getFile(); err != nil {
			lib.SendErrorResponse(err, w)
			return
		}

		if !lib.IsUserAuthorized(false, DeleteFile.File.ProjectID, []string{lib.ROLE_PROJECT_ADMIN}, DeleteFile.CurrentUser, nil) {
			if DeleteFile.FileStorage.DirectoryID != DeleteFile.CurrentUser.DirectoryID {
				lib.SendErrorResponse(lib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden)), w)
				return
			}
		}

		filesAffected, err := DeleteFile.deleteFileInfo()
		if err != nil {
			lib.SendErrorResponse(err, w)
		}
		if err := DeleteFile.deleteFile(); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete file %v from project %v by %v failed | reason: %v", DeleteFile.File.ID, DeleteFile.File.ProjectID, DeleteFile.CurrentUser.DirectoryID, err))
			if err := DeleteFile.recreateFileInfo(); err != nil {
				lib.SendErrorResponse(err, w)
				return
			}
			lib.SendErrorResponse(lib.NewError(http.StatusInternalServerError, "Could not delete file"), w)
			return
		} else {
			lib.SendJsonResponse(struct{ FilesAffected int64 }{FilesAffected: filesAffected}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("In delete file %v by %v, %v were affected", DeleteFile.File.ID, DeleteFile.CurrentUser.DirectoryID, filesAffected))
		}

	})

	router.Get("/file", func(w http.ResponseWriter, r *http.Request) {
		var RetrieveFiles storage
		RetrieveFiles.CurrentUser = lib.CtxGetCurrentUser(r)
		if fileId, err := lib.GetUUID(r.URL.Query().Get("id")); err == nil {
			RetrieveFiles.FileID = fileId
			if err := RetrieveFiles.getFiles(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveFiles.FileStorageDirectoryProject, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get file %v by %v successful", fileId, RetrieveFiles.CurrentUser.DirectoryID))
			}
		} else {
			if sq := r.URL.Query().Get("sq"); sq != "" {
				RetrieveFiles.SearchQuery = sq
			}
			if cogt := r.URL.Query().Get("cogt"); cogt != "" {
				RetrieveFiles.CreatedOnGreaterThan = cogt
			}
			if colt := r.URL.Query().Get("colt"); colt != "" {
				RetrieveFiles.CreatedOnLessThan = colt
			}
			if pid, err := lib.GetUUID(r.URL.Query().Get("pid")); err == nil {
				RetrieveFiles.ProjectID = pid
			}
			if limit, err := strconv.Atoi(r.URL.Query().Get("l")); err == nil {
				RetrieveFiles.Limit = limit
			}
			if offset, err := strconv.Atoi(r.URL.Query().Get("o")); err == nil {
				RetrieveFiles.Offset = offset
			}
			if qs := r.URL.Query().Get("qs"); qs == "true" || qs == "false" {
				RetrieveFiles.QuickSearch = qs
			}
			if sb := r.URL.Query().Get("sb"); sb != "" {
				RetrieveFiles.SortyBy = sb
			}
			if sbo := r.URL.Query().Get("sbo"); sbo != "" {
				RetrieveFiles.SortByOrder = sbo
			}
			if fwa := r.URL.Query().Get("fwa"); fwa == "true" || fwa == "false" {
				RetrieveFiles.FilesWithAbstractions = fwa
			}
			if err := RetrieveFiles.getFiles(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveFiles.FilesStorageDirectoryProject, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get files by %v successful", RetrieveFiles.CurrentUser.DirectoryID))
			}
		}
	})

	router.Get("/file/{project_id}/{file_id}", func(w http.ResponseWriter, r *http.Request) {
		var RetrievedFile storage

		RetrievedFile.File = model.Files{}
		if projectId, err := lib.GetUUID(chi.URLParam(r, "project_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			RetrievedFile.File.ProjectID = projectId
		}

		if fileId, err := lib.GetUUID(chi.URLParam(r, "file_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			RetrievedFile.File.ID = fileId
		}

		if err := RetrievedFile.getFile(); err != nil {
			lib.SendErrorResponse(err, w)
			return
		}

		RetrievedFile.CurrentUser = lib.CtxGetCurrentUser(r)
		if err := RetrievedFile.sendFile(w); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Send file %v from project %v by %v failed | reason: %v", RetrievedFile.File.ID, RetrievedFile.File.ProjectID, RetrievedFile.CurrentUser.DirectoryID, err))
			lib.SendErrorResponse(lib.NewError(http.StatusInternalServerError, "Could not send file"), w)
		} else {
			intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Download file %v by %v successful", RetrievedFile.File.ID, RetrievedFile.CurrentUser.DirectoryID))
		}
	})

	router.Post("/file", func(w http.ResponseWriter, r *http.Request) {
		var NewFile storage

		if err := r.ParseMultipartForm(50); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		file, handler, err := r.FormFile("FileUpload")
		if err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}
		defer file.Close()

		NewFile.StorageProject = model.StorageProjects{}
		NewFile.File = model.Files{}
		NewFile.CurrentUser = lib.CtxGetCurrentUser(r)
		if projectId, err := lib.GetUUID(r.FormValue("ProjectID")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			if !lib.IsUserAuthorized(false, projectId, []string{lib.ROLE_FILE_CREATOR}, NewFile.CurrentUser, w) {
				return
			}
			NewFile.StorageProject.ProjectID = projectId
			NewFile.File.ProjectID = projectId
		}
		if storageId, err := lib.GetUUID(r.FormValue("StorageID")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			NewFile.StorageProject.StorageID = storageId
			NewFile.File.StorageID = storageId
		}
		NewFile.File.DirectoryID = NewFile.CurrentUser.DirectoryID
		NewFile.File.Tags = r.FormValue("Tags")
		NewFile.File.ContentType = handler.Header.Get("Content-Type")

		if err := NewFile.createFile(); err != nil {
			lib.SendErrorResponse(err, w)
			return
		}

		if err := NewFile.storeFile(&file); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct {
				ID        uuid.UUID
				CreatedOn time.Time
			}{ID: NewFile.File.ID, CreatedOn: NewFile.File.CreatedOn}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("File %v created in project %v by %v", NewFile.File.ID, NewFile.StorageProject.ProjectID, NewFile.CurrentUser.DirectoryID))
		}
	})

	router.Delete("/project", func(w http.ResponseWriter, r *http.Request) {
		var DeleteStorageProject storage

		if err := json.NewDecoder(r.Body).Decode(&DeleteStorageProject.AddRemoveStorageProject); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		DeleteStorageProject.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, DeleteStorageProject.CurrentUser, w) {
			return
		}

		if storageProjectsAffected, err := DeleteStorageProject.deleteStorageProject(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ StorageProjectsAffected int64 }{StorageProjectsAffected: storageProjectsAffected}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("In delete storage for project %v by %v, %v were affected", DeleteStorageProject.AddRemoveStorageProject.ProjectID, DeleteStorageProject.CurrentUser.DirectoryID, storageProjectsAffected))
		}
	})

	router.Get("/project", func(w http.ResponseWriter, r *http.Request) {
		var RetrieveProjectsStorage storage

		RetrieveProjectsStorage.CurrentUser = lib.CtxGetCurrentUser(r)

		if cogt := r.URL.Query().Get("cogt"); cogt != "" {
			RetrieveProjectsStorage.CreatedOnGreaterThan = cogt
		}
		if colt := r.URL.Query().Get("colt"); colt != "" {
			RetrieveProjectsStorage.CreatedOnLessThan = colt
		}
		if pid, err := lib.GetUUID(r.URL.Query().Get("pid")); err == nil {
			RetrieveProjectsStorage.ProjectID = pid
		}
		if limit, err := strconv.Atoi(r.URL.Query().Get("l")); err == nil {
			RetrieveProjectsStorage.Limit = limit
		}
		if offset, err := strconv.Atoi(r.URL.Query().Get("o")); err == nil {
			RetrieveProjectsStorage.Offset = offset
		}

		if err := RetrieveProjectsStorage.getProjectsStorage(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(RetrieveProjectsStorage.ProjectsStorage, w)
			intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get project's storage by %v successful", RetrieveProjectsStorage.CurrentUser.DirectoryID))
		}
	})

	router.Post("/project", func(w http.ResponseWriter, r *http.Request) {
		var NewStorageProject storage
		if err := json.NewDecoder(r.Body).Decode(&NewStorageProject.AddRemoveStorageProject); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		NewStorageProject.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, NewStorageProject.CurrentUser, w) {
			return
		}

		if err := NewStorageProject.createStorageProject(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ CreatedOn time.Time }{CreatedOn: time.Now()}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("Storage added to project %v by %v", NewStorageProject.AddRemoveStorageProject.ProjectID, NewStorageProject.CurrentUser.DirectoryID))
		}
	})

	router.Get("/type", func(w http.ResponseWriter, r *http.Request) {
		var RetrieveStorageTypes storage

		RetrieveStorageTypes.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, RetrieveStorageTypes.CurrentUser, w) {
			return
		}

		if err := RetrieveStorageTypes.getStorageTypes(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(RetrieveStorageTypes.StorageTypes, w)
			intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get storage types by %v successful", RetrieveStorageTypes.CurrentUser.DirectoryID))
		}
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var RetrieveStorage storage

		RetrieveStorage.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, RetrieveStorage.CurrentUser, w) {
			return
		}

		if storageId, err := lib.GetUUID(r.URL.Query().Get("id")); err == nil {
			RetrieveStorage.StorageID = storageId
			if err := RetrieveStorage.getStorages(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveStorage.RetrieveStorage, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get storage %v by %v successful", storageId, RetrieveStorage.CurrentUser.DirectoryID))
			}
		} else {
			if limit, err := strconv.Atoi(r.URL.Query().Get("l")); err == nil {
				RetrieveStorage.Limit = limit
			}
			if offset, err := strconv.Atoi(r.URL.Query().Get("o")); err == nil {
				RetrieveStorage.Offset = offset
			}
			if cogt := r.URL.Query().Get("cogt"); cogt != "" {
				RetrieveStorage.CreatedOnGreaterThan = cogt
			}
			if colt := r.URL.Query().Get("colt"); colt != "" {
				RetrieveStorage.CreatedOnLessThan = colt
			}
			if logt := r.URL.Query().Get("logt"); logt != "" {
				RetrieveStorage.LastUpdatedOnOnGreaterThan = logt
			}
			if lolt := r.URL.Query().Get("lolt"); lolt != "" {
				RetrieveStorage.LastUpdatedOnLessThan = lolt
			}
			if sb := r.URL.Query().Get("sb"); sb != "" {
				RetrieveStorage.SortyBy = sb
			}
			if sbo := r.URL.Query().Get("sbo"); sbo != "" {
				RetrieveStorage.SortByOrder = sbo
			}
			if err := RetrieveStorage.getStorages(); err != nil {
				lib.SendErrorResponse(err, w)
			} else {
				lib.SendJsonResponse(RetrieveStorage.RetrieveStorages, w)
				intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get storages by %v successful", RetrieveStorage.CurrentUser.DirectoryID))
			}
		}
	})

	router.Delete("/{storage_id}", func(w http.ResponseWriter, r *http.Request) {
		var DeleteStorage storage

		if storageID, err := lib.GetUUID(chi.URLParam(r, "storage_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			DeleteStorage.StorageID = storageID
		}

		DeleteStorage.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, DeleteStorage.CurrentUser, w) {
			return
		}

		if storageAffected, err := DeleteStorage.deleteStorage(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ StorageAffected int64 }{StorageAffected: storageAffected}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("In delete/deactivate storage %v by %v, %v were affected", DeleteStorage.StorageID, DeleteStorage.CurrentUser.DirectoryID, storageAffected))
		}
	})

	router.Put("/{storage_id}", func(w http.ResponseWriter, r *http.Request) {
		var UpdateStorage storage

		if storageID, err := lib.GetUUID(chi.URLParam(r, "storage_id")); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			UpdateStorage.StorageID = storageID
		}

		if err := json.NewDecoder(r.Body).Decode(&UpdateStorage.StorageUpdate); err != nil || len(UpdateStorage.StorageUpdate.Columns) > 4 {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		UpdateStorage.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, UpdateStorage.CurrentUser, w) {
			return
		}

		if err := UpdateStorage.updateStorage(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct {
				ID            uuid.UUID
				LastUpdatedOn time.Time
			}{ID: UpdateStorage.StorageUpdate.Storage.ID, LastUpdatedOn: UpdateStorage.StorageUpdate.Storage.LastUpdatedOn}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("Storage %v updated by %v", UpdateStorage.StorageUpdate.Storage.ID, UpdateStorage.CurrentUser.DirectoryID))
		}
	})

	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var NewStorage storage
		if err := json.NewDecoder(r.Body).Decode(&NewStorage.Storage); err != nil || (model.Storage{} == NewStorage.Storage) {
			lib.SendErrorResponse(lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		NewStorage.CurrentUser = lib.CtxGetCurrentUser(r)
		if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, NewStorage.CurrentUser, w) {
			return
		}

		if err := NewStorage.createStorage(); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			lib.SendJsonResponse(struct{ ID uuid.UUID }{ID: NewStorage.Storage.ID}, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("Storage %v created by %v", NewStorage.Storage.ID, NewStorage.CurrentUser.DirectoryID))
		}
	})

	return router
}
