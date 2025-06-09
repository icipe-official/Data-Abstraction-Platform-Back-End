package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type IamCredentials struct {
	ID             []uuid.UUID `json:"id,omitempty"`
	DirectoryID    []uuid.UUID `json:"directory_id,omitempty"`
	OpenidUserInfo []struct {
		OpenidSub               []string `json:"openid_sub,omitempty"`
		OpenidPreferredUsername []string `json:"openid_preferred_username,omitempty"`
		OpenidEmail             []string `json:"openid_email,omitempty"`
		OpenidEmailVerified     []bool   `json:"openid_email_verified,omitempty"`
		OpenidGivenName         []string `json:"openid_given_name,omitempty"`
		OpenidFamilyName        []string `json:"openid_family_name,omitempty"`
	} `json:"openid_user_info,omitempty"`
	CreatedOn     []time.Time `json:"created_on,omitempty"`
	LastUpdatedOn []time.Time `json:"last_updated_on,omitempty"`
	DeactivatedOn []time.Time `json:"deactivated_on,omitempty"`
}

type iamCredentialsRepository struct {
	RepositoryName string

	ID                      string
	DirectoryID             string
	OpenidSub               string
	OpenidPreferredUsername string
	OpenidEmail             string
	OpenidEmailVerified     string
	OpenidGivenName         string
	OpenidFamilyName        string
	CreatedOn               string
	LastUpdatedOn           string
	DeactivatedOn           string
	FullTextSearch          string
}

func IamCredentialsRepository() iamCredentialsRepository {
	return iamCredentialsRepository{
		RepositoryName: "iam_credentials",

		ID:                      "id",
		DirectoryID:             "directory_id",
		OpenidSub:               "openid_sub",
		OpenidPreferredUsername: "openid_preferred_username",
		OpenidEmail:             "openid_email",
		OpenidEmailVerified:     "openid_email_verified",
		OpenidGivenName:         "openid_given_name",
		OpenidFamilyName:        "openid_family_name",
		CreatedOn:               "created_on",
		LastUpdatedOn:           "last_updated_on",
		DeactivatedOn:           "deactivated_on",
		FullTextSearch:          "full_text_search",
	}
}
