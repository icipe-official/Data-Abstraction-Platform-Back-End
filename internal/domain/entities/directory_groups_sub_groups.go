package entities

import (
	"github.com/gofrs/uuid/v5"
)

type DirectoryGroupsSubGroups struct {
	ParentGroupID []uuid.UUID `json:"parent_group_id,omitempty"`
	SubGroupID    []uuid.UUID `json:"sub_group_id,omitempty"`
}

type directoryGroupsSubGroupsRepository struct {
	RepositoryName string

	ParentGroupID string
	SubGroupID    string
}

func DirectoryGroupsSubGroupsRepository() directoryGroupsSubGroupsRepository {
	return directoryGroupsSubGroupsRepository{
		RepositoryName: "directory_groups_sub_groups",

		ParentGroupID: "parent_group_id",
		SubGroupID:    "sub_group_id",
	}
}
