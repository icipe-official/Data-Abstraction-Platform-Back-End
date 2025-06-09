package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type MetadataModelsDirectoryGroups struct {
	DirectoryGroupsID []uuid.UUID `json:"directory_groups_id,omitempty"`
	MetadataModelsID  []uuid.UUID `json:"metadata_models_id,omitempty"`
	CreatedOn         []time.Time `json:"created_on,omitempty"`
	LastUpdatedOn     []time.Time `json:"last_updated_on,omitempty"`
}

type metadataModelsDirectoryGroupsRepository struct {
	RepositoryName string

	DirectoryGroupsID string
	MetadataModelsID  string
	CreatedOn         string
	LastUpdatedOn     string
}

func MetadataModelsDirectoryGroupsRepository() metadataModelsDirectoryGroupsRepository {
	return metadataModelsDirectoryGroupsRepository{
		RepositoryName: "metadata_models_directory_groups",

		DirectoryGroupsID: "directory_groups_id",
		MetadataModelsID:  "metadata_models_id",
		CreatedOn:         "created_on",
		LastUpdatedOn:     "last_updated_on",
	}
}
