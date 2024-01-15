package storage

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/google/uuid"
)

const currentSection = "Storage"

type storage struct {
	StorageTypes                 []model.StorageTypes
	StoragesStorageType          []storageStorageType
	StorageStorageType           storageStorageType
	ProjectsStorage              []projectStorage
	SearchQuery                  string    `json:"-"`
	ProjectID                    uuid.UUID `json:"-"`
	CreatedOnGreaterThan         string    `json:"-"`
	CreatedOnLessThan            string    `json:"-"`
	Limit                        int       `json:"-"`
	Offset                       int       `json:"-"`
	FileID                       uuid.UUID `json:"-"`
	QuickSearch                  string    `json:"-"`
	SortyBy                      string    `json:"-"`
	SortByOrder                  string    `json:"-"`
	FilesWithAbstractions        string    `json:"-"`
	FilesStorageDirectoryProject []fileStorageDirectoryProject
	FileStorageDirectoryProject  fileStorageDirectoryProject
	StorageUpdate                struct {
		Storage model.Storage
		Columns []string
	}
	StorageID      uuid.UUID
	FileStorage    fileStorage
	CurrentUser    lib.User
	File           model.Files
	StorageProject model.StorageProjects
	Storage        model.Storage
	ProjectStorage storageProject
}

type storageStorageType struct {
	model.Storage
	StorageType model.StorageTypes
}

type projectStorage struct {
	model.StorageProjects
	Storage model.Storage
	Project model.Projects
}

type fileStorageDirectoryProject struct {
	model.Files
	Storage   model.Storage
	Directory model.Directory
	Project   model.Projects
}

type storageProject struct {
	model.StorageProjects
	Storage model.Storage
}

type fileStorage struct {
	model.Files
	Storage model.Storage
}

type mountedStorage struct {
	Path string `json:"path"`
}

func (n *storage) deleteFile() error {
	switch n.FileStorage.Storage.StorageTypeID {
	case lib.STORAGE_AZURE_BLOB_MOUNTED, lib.STORAGE_LOCAL:
		var storageProperties mountedStorage
		if err := json.Unmarshal([]byte(n.FileStorage.Storage.Storage), &storageProperties); err != nil {
			return fmt.Errorf("get storage properties failed | reason: %v", err)
		}
		if err := os.Remove(fmt.Sprintf("%v/%v/%v", storageProperties.Path, n.File.ProjectID, n.File.ID.String())); err != nil {
			return fmt.Errorf("could not delete file | reason: %v", err)
		}
	default:
		return fmt.Errorf("storage %v is invalid/unsupported", n.ProjectStorage.Storage.StorageTypeID)
	}

	return nil
}

func (n *storage) sendFile(w http.ResponseWriter) error {
	switch n.FileStorage.Storage.StorageTypeID {
	case lib.STORAGE_AZURE_BLOB_MOUNTED, lib.STORAGE_LOCAL:
		var storageProperties mountedStorage
		if err := json.Unmarshal([]byte(n.FileStorage.Storage.Storage), &storageProperties); err != nil {
			return fmt.Errorf("get storage properties failed | reason: %v", err)
		}
		var file *os.File
		var tmpFileId uuid.UUID
		if n.FileStorage.Storage.StorageTypeID == lib.STORAGE_LOCAL {
			if openFile, err := os.Open(fmt.Sprintf("%v/%v/%v", storageProperties.Path, n.File.ProjectID, n.File.ID.String())); err != nil {
				return fmt.Errorf("open file failed | reason: %v", err)
			} else {
				file = openFile
			}
		} else {
			srcFile, err := os.Open(fmt.Sprintf("%v/%v/%v", storageProperties.Path, n.File.ProjectID, n.File.ID.String()))
			if err != nil {
				return fmt.Errorf("open src file failed | reason: %v", err)
			}
			tmpFileId = uuid.New()
			dstFile, err := os.Create(fmt.Sprintf("%v/%v", lib.TMP_DIR, tmpFileId.String()))
			if err != nil {
				srcFile.Close()
				return fmt.Errorf("create dst tmp file failed | reason: %v", err)
			}
			if _, err = io.Copy(dstFile, srcFile); err != nil {
				srcFile.Close()
				dstFile.Close()
				return fmt.Errorf("copy src file contents to dst file failed | reason: %v", err)
			}
			if openFile, err := os.Open(fmt.Sprintf("%v/%v", lib.TMP_DIR, tmpFileId.String())); err != nil {
				return fmt.Errorf("open new tmp file failed | reason: %v", err)
			} else {
				file = openFile
			}
		}
		defer file.Close()
		lib.SendFile(w, &n.File.ContentType, file)
		if tmpFileId != uuid.Nil {
			if err := os.Remove(fmt.Sprintf("%v/%v", lib.TMP_DIR, tmpFileId.String())); err != nil {
				intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Delete tmp file %v failed | reason: %v", tmpFileId, err))
			}
		}
	default:
		return fmt.Errorf("storage %v is invalid/unsupported", n.ProjectStorage.Storage.StorageTypeID)
	}

	return nil
}

func (n *storage) storeFile(file *multipart.File) error {
	switch n.ProjectStorage.Storage.StorageTypeID {
	case lib.STORAGE_AZURE_BLOB_MOUNTED, lib.STORAGE_LOCAL:
		var storageProperties mountedStorage
		if err := json.Unmarshal([]byte(n.ProjectStorage.Storage.Storage), &storageProperties); err != nil {
			return fmt.Errorf("get storage properties failed | reason: %v", err)
		}
		out, err := os.Create(fmt.Sprintf("%v/%v/%v", storageProperties.Path, n.StorageProject.ProjectID, n.File.ID.String()))
		if err != nil {
			return fmt.Errorf("create file %v in %v failed | reason: %v", n.File.ID, n.StorageProject.ProjectID, err)
		}
		defer out.Close()

		if _, err = io.Copy(out, *file); err != nil {
			return fmt.Errorf("copy contents to file %v in %v failed | reason: %v", n.File.ID, n.StorageProject.ProjectID, err)
		}
	default:
		return fmt.Errorf("storage %v is invalid/unsupported", n.ProjectStorage.Storage.StorageTypeID)
	}
	return nil
}

func (n *storage) createFolder(folder string) error {
	switch n.Storage.StorageTypeID {
	case lib.STORAGE_AZURE_BLOB_MOUNTED, lib.STORAGE_LOCAL:
		var storageProperties mountedStorage
		if err := json.Unmarshal([]byte(n.Storage.Storage), &storageProperties); err != nil {
			return fmt.Errorf("get storage properties failed | reason: %v", err)
		}
		if _, err := os.Stat(fmt.Sprintf("%v/%v", storageProperties.Path, folder)); os.IsNotExist(err) {
			if err = os.Mkdir(fmt.Sprintf("%v/%v", storageProperties.Path, folder), os.ModeDir|0755); err != nil {
				return fmt.Errorf("create folder %v failed | reason: %v", folder, err)
			}
		}
	default:
		return fmt.Errorf("storage %v is invalid/unsupported", n.Storage.ID)
	}
	return nil
}
