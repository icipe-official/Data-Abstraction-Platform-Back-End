package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Abstractions struct {
	ID                            []uuid.UUID `json:"id,omitempty"`
	AbstractionsDirectoryGroupsID []uuid.UUID `json:"abstractions_directory_groups_id,omitempty"`
	DirectoryID                   []uuid.UUID `json:"directory_id,omitempty"`
	StorageFilesID                []uuid.UUID `json:"storage_files_id,omitempty"`
	Completed                     []bool      `json:"completed,omitempty"`
	ReviewPass                    []bool      `json:"review_pass,omitempty"`
	Tags                          []string    `json:"tags,omitempty"`
	Data                          []any       `json:"data,omitempty"`
	CreatedOn                     []time.Time `json:"created_on,omitempty"`
	LastUpdatedOn                 []time.Time `json:"last_updated_on,omitempty"`
	DeactivatedOn                 []time.Time `json:"deactivated_on,omitempty"`
}

type abstractionsRepository struct {
	RepositoryName string

	ID                            string
	AbstractionsDirectoryGroupsID string
	DirectoryID                   string
	StorageFilesID                string
	Completed                     string
	ReviewPass                    string
	Tags                          string
	Data                          string
	CreatedOn                     string
	LastUpdatedOn                 string
	DeactivatedOn                 string
	FullTextSearch                string
}

func AbstractionsRepository() abstractionsRepository {
	return abstractionsRepository{
		RepositoryName: "abstractions",

		ID:                            "id",
		AbstractionsDirectoryGroupsID: "abstractions_directory_groups_id",
		DirectoryID:                   "directory_id",
		StorageFilesID:                "storage_files_id",
		Data:                          "data",
		Tags:                          "tags",
		Completed:                     "completed",
		ReviewPass:                    "review_pass",
		CreatedOn:                     "created_on",
		LastUpdatedOn:                 "last_updated_on",
		DeactivatedOn:                 "deactivated_on",
		FullTextSearch:                "full_text_search",
	}
}

type AbstractionsUpdateDirectory struct {
	DirectoryID                []uuid.UUID `json:"directory_id,omitempty"`
	DirectoryGroupID           []uuid.UUID `json:"directory_group_id,omitempty"`
	AbstractionsID             []uuid.UUID `json:"abstractions_id,omitempty"`
	StorageFilesFullTextSearch []string    `json:"storage_files_full_text_search,omitempty"`
	Completed                  *bool       `json:"completed,omitempty"`
	ReviewPass                 *bool       `json:"review_pass,omitempty"`
	NewDirectoryID             *uuid.UUID  `json:"new_directory_id,omitempty"`
}

type AbstractionsExportColumns struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

type AbstractionsExportData struct {
	Columns          []AbstractionsExportColumns `json:"columns,omitempty"`
	MetadataModelsID uuid.UUID                   `json:"metadata_models_id,omitempty"`
	MetadataModel    any                         `json:"metadata_model,omitempty"`
}
