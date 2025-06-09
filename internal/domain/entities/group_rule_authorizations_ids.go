package entities

import (
	"github.com/gofrs/uuid/v5"
)

type GroupRuleAuthorizationsIDs struct {
	ID                                   []uuid.UUID `json:"id,omitempty"`
	CreationIamGroupAuthorizationsID     []uuid.UUID `json:"creation_iam_group_authorizations_id,omitempty"`
	DeactivationIamGroupAuthorizationsID []uuid.UUID `json:"deactivation_iam_group_authorizations_id,omitempty"`
}

type groupRuleAuthorizationsIDsRepository struct {
	RepositoryName string

	ID                                   string
	CreationIamGroupAuthorizationsID     string
	DeactivationIamGroupAuthorizationsID string
}

func GroupRuleAuthorizationsIDsRepository() groupRuleAuthorizationsIDsRepository {
	return groupRuleAuthorizationsIDsRepository{
		RepositoryName: "group_rule_authorizations_ids",

		ID:                                   "id",
		CreationIamGroupAuthorizationsID:     "creation_iam_group_authorizations_id",
		DeactivationIamGroupAuthorizationsID: "deactivation_iam_group_authorizations_id",
	}
}
