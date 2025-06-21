package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type StorageFilesTemporary struct {
	ID                  []uuid.UUID `sql:"primary_key" json:"id,omitempty"`
	StorageFileMimeType []string    `json:"storage_file_mime_type,omitempty"`
	OriginalName        []string    `json:"original_name,omitempty"`
	Tags                []string    `json:"tags,omitempty"`
	SizeInBytes         []int64     `json:"size_in_bytes,omitempty"`
	CreatedOn           []time.Time `json:"created_on,omitempty"`
}

type storageFilesTemporaryRepository struct {
	RepositoryName string

	ID                  string
	StorageFileMimeType string
	OriginalName        string
	Tags                string
	SizeInBytes         string
	CreatedOn           string
}

func StorageFilesTemporaryRepository() storageFilesTemporaryRepository {
	return storageFilesTemporaryRepository{
		RepositoryName: "storage_files_temporary",

		ID:                  "id",
		StorageFileMimeType: "storage_file_mime_type",
		OriginalName:        "original_name",
		Tags:                "tags",
		SizeInBytes:         "size_in_bytes",
		CreatedOn:           "created_on",
	}
}

type StorageFilesTemporaryDelete struct {
	Success int64
	Failed  []error
}
