package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type MetadataModels struct {
	ID                []uuid.UUID `json:"id,omitempty"`
	DirectoryID       []uuid.UUID `json:"directory_id,omitempty"`
	DirectoryGroupsID []uuid.UUID `json:"directory_groups_id,omitempty"`
	Name              []string    `json:"name,omitempty"`
	Description       []string    `json:"description,omitempty"`
	EditAuthorized    []bool      `json:"edit_authorized,omitempty"`
	EditUnauthorized  []bool      `json:"edit_unauthorized,omitempty"`
	ViewAuthorized    []bool      `json:"view_authorized,omitempty"`
	ViewUnauthorized  []bool      `json:"view_unauthorized,omitempty"`
	Tags              []string    `json:"tags,omitempty"`
	Data              []any       `json:"data,omitempty"`
	CreatedOn         []time.Time `json:"created_on,omitempty"`
	LastUpdatedOn     []time.Time `json:"last_updated_on,omitempty"`
	DeactivatedOn     []time.Time `json:"deactivated_on,omitempty"`
}

type metadataModelsRepository struct {
	RepositoryName string

	ID                string
	DirectoryGroupsID string
	DirectoryID       string
	Name              string
	Description       string
	EditAuthorized    string
	EditUnauthorized  string
	ViewAuthorized    string
	ViewUnauthorized  string
	Tags              string
	Data              string
	CreatedOn         string
	LastUpdatedOn     string
	DeactivatedOn     string
	FullTextSearch    string
}

func MetadataModelsRepository() metadataModelsRepository {
	return metadataModelsRepository{
		RepositoryName: "metadata_models",

		ID:                "id",
		DirectoryGroupsID: "directory_groups_id",
		DirectoryID:       "directory_id",
		Name:              "name",
		Description:       "description",
		EditAuthorized:    "edit_authorized",
		EditUnauthorized:  "edit_unauthorized",
		ViewAuthorized:    "view_authorized",
		ViewUnauthorized:  "view_unauthorized",
		Tags:              "tags",
		Data:              "data",
		CreatedOn:         "created_on",
		LastUpdatedOn:     "last_updated_on",
		DeactivatedOn:     "deactivated_on",
		FullTextSearch:    "full_text_search",
	}
}
