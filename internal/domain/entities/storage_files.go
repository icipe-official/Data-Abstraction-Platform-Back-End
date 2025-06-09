package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type StorageFiles struct {
	ID                  []uuid.UUID `sql:"primary_key" json:"id,omitempty"`
	DirectoryGroupsID   []uuid.UUID `json:"directory_groups_id,omitempty"`
	DirectoryID         []uuid.UUID `json:"directory_id,omitempty"`
	StorageFileMimeType []string    `json:"storage_file_mime_type,omitempty"`
	OriginalName        []string    `json:"original_name,omitempty"`
	Tags                []string    `json:"tags,omitempty"`
	EditAuthorized      []bool      `json:"edit_authorized,omitempty"`
	EditUnauthorized    []bool      `json:"edit_unauthorized,omitempty"`
	ViewAuthorized      []bool      `json:"view_authorized,omitempty"`
	ViewUnauthorized    []bool      `json:"view_unauthorized,omitempty"`
	SizeInBytes         []int64     `json:"size_in_bytes,omitempty"`
	CreatedOn           []time.Time `json:"created_on,omitempty"`
	LastUpdatedOn       []time.Time `json:"last_updated_on,omitempty"`
	DeactivatedOn       []time.Time `json:"deactivated_on,omitempty"`
}

type storageFilesRepository struct {
	RepositoryName string

	ID                  string
	DirectoryGroupsID   string
	DirectoryID         string
	StorageFileMimeType string
	OriginalName        string
	Tags                string
	EditAuthorized      string
	EditUnauthorized    string
	ViewAuthorized      string
	ViewUnauthorized    string
	SizeInBytes         string
	CreatedOn           string
	LastUpdatedOn       string
	DeactivatedOn       string
	FullTextSearch      string
}

func StorageFilesRepository() storageFilesRepository {
	return storageFilesRepository{
		RepositoryName: "storage_files",

		ID:                  "id",
		DirectoryGroupsID:   "directory_groups_id",
		DirectoryID:         "directory_id",
		StorageFileMimeType: "storage_file_mime_type",
		OriginalName:        "original_name",
		Tags:                "tags",
		EditAuthorized:      "edit_authorized",
		EditUnauthorized:    "edit_unauthorized",
		ViewAuthorized:      "view_authorized",
		ViewUnauthorized:    "view_unauthorized",
		SizeInBytes:         "size_in_bytes",
		CreatedOn:           "created_on",
		LastUpdatedOn:       "last_updated_on",
		DeactivatedOn:       "deactivated_on",
		FullTextSearch:      "full_text_search",
	}
}
