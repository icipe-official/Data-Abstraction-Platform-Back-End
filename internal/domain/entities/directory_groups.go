package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type DirectoryGroups struct {
	ID            []uuid.UUID `json:"id,omitempty"`
	DisplayName   []string    `json:"display_name,omitempty"`
	Description   []string    `json:"description,omitempty"`
	Data          []any       `json:"data,omitempty"`
	CreatedOn     []time.Time `json:"created_on,omitempty"`
	LastUpdatedOn []time.Time `json:"last_updated_on,omitempty"`
	DeactivatedOn []time.Time `json:"deactivated_on,omitempty"`
}

type directoryGroupsRepository struct {
	RepositoryName string

	ID             string
	DisplayName    string
	Description    string
	Data           string
	CreatedOn      string
	LastUpdatedOn  string
	DeactivatedOn  string
	FullTextSearch string
}

func DirectoryGroupsRepository() directoryGroupsRepository {
	return directoryGroupsRepository{
		RepositoryName: "directory_groups",

		ID:             "id",
		DisplayName:    "display_name",
		Description:    "description",
		Data:           "data",
		CreatedOn:      "created_on",
		LastUpdatedOn:  "last_updated_on",
		DeactivatedOn:  "deactivated_on",
		FullTextSearch: "full_text_search",
	}
}
