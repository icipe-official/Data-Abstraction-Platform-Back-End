package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type IamGroupAuthorizations struct {
	ID                        []uuid.UUID `json:"id,omitempty"`
	IamCredentialsID          []uuid.UUID `json:"iam_credentials_id,omitempty"`
	GroupRuleAuthorizationsID []uuid.UUID `json:"group_rule_authorizations_id,omitempty"`
	CreatedOn                 []time.Time `json:"created_on,omitempty"`
	DeactivatedOn             []time.Time `json:"deactivated_on,omitempty"`
}

type iamGroupAuthorizationsRepository struct {
	RepositoryName string

	ID                        string
	IamCredentialsID          string
	GroupRuleAuthorizationsID string
	CreatedOn                 string
	DeactivatedOn             string
}

func IamGroupAuthorizationsRepository() iamGroupAuthorizationsRepository {
	return iamGroupAuthorizationsRepository{
		RepositoryName: "iam_group_authorizations",

		ID:                        "id",
		IamCredentialsID:          "iam_credentials_id",
		GroupRuleAuthorizationsID: "group_rule_authorizations_id",
		CreatedOn:                 "created_on",
		DeactivatedOn:             "deactivated_on",
	}
}
