package entities

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
)

type IamAuthenticationHeaders struct {
	AccessTokenHeader  string        `json:"access_token_header"`
	RefreshTokenHeader string        `json:"refresh_token_header"`
	CookieHttpOnly     bool          `json:"cookie_http_only"`
	CookieSameSite     http.SameSite `json:"cookie_same_site"`
	CookieSecure       bool          `json:"cookie_secure"`
	CookieDomain       string        `json:"cookie_domain"`
}

type IamSession struct {
	IamCredential *IamCredentials `json:"iam_credential,omitempty"`
	Directory     *Directory      `json:"directory,omitempty"`
}

type IamOpenIDEndpoints struct {
	LoginEndpoint             string `json:"login_endpoint,omitempty"`
	RegistrationEndpoint      string `json:"registration_endpoint,omitempty"`
	AccountManagementEndpoint string `json:"account_management_endpoint,omitempty"`
}

type IamGroupAuthorizationRule struct {
	ID        string `json:"id"`
	RuleGroup string `json:"rule_group"`
}

// formed by joining fields in iamGroupAuthorizationsRepository and groupAuthorizationRulesRepository.
type IamAuthorizationRule struct {
	ID                          uuid.UUID `json:"id,omitempty"`
	DirectoryGroupID            uuid.UUID `json:"directory_group_id,omitempty"`
	GroupAuthorizationRuleID    string    `json:"group_authorization_rule_id,omitempty"`
	GroupAuthorizationRuleGroup string    `json:"group_authorization_rule_group,omitempty"`
}

type IamAuthorizationRules map[string]*IamAuthorizationRule

const (
	AUTH_RULE_GROUP_GROUP_RULE_AUTHORIZATIONS        string = "group_rule_authorizations"
	AUTH_RULE_GROUP_IAM_GROUP_AUTHORIZATIONS         string = "iam_group_authorizations"
	AUTH_RULE_GROUP_DIRECTORY_GROUPS                 string = "directory_groups"
	AUTH_RULE_GROUP_DIRECTORY_GROUPS_TYPES           string = "directory_groups_types"
	AUTH_RULE_GROUP_DIRECTORY                        string = "directory"
	AUTH_RULE_GROUP_IAM_CREDENTIALS                  string = "iam_credentials"
	AUTH_RULE_GROUP_METADATA_MODELS                  string = "metadata_models"
	AUTH_RULE_GROUP_METADATA_MODELS_DIRECTORY        string = "metadata_models_directory"
	AUTH_RULE_GROUP_METADATA_MODELS_DIRECTORY_GROUPS string = "metadata_models_directory_groups"
	AUTH_RULE_GROUP_STORAGE_FILES                    string = "storage_files"
	AUTH_RULE_GROUP_ABSTRACTIONS_DIRECTORY_GROUPS    string = "abstractions_directory_groups"
	AUTH_RULE_GROUP_ABSTRACTIONS                     string = "abstractions"
	AUTH_RULE_GROUP_ABSTRACTIONS_REVIEWS             string = "abstractions_reviews"

	AUTH_RULE_ASSIGN_PREFIX string = "assign_"

	AUTH_RULE_CREATE                 string = "create"
	AUTH_RULE_ASSIGN_CREATE          string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_CREATE
	AUTH_RULE_CREATE_OTHERS          string = "create_others"
	AUTH_RULE_ASSIGN_CREATE_OTHERS   string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_CREATE_OTHERS
	AUTH_RULE_RETRIEVE_SELF          string = "retrieve_self"
	AUTH_RULE_ASSIGN_RETRIEVE_SELF   string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_RETRIEVE_SELF
	AUTH_RULE_RETRIEVE               string = "retrieve"
	AUTH_RULE_ASSIGN_RETRIEVE        string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_RETRIEVE
	AUTH_RULE_RETRIEVE_OTHERS        string = "retrieve_others"
	AUTH_RULE_ASSIGN_RETRIEVE_OTHERS string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_RETRIEVE_OTHERS
	AUTH_RULE_UPDATE                 string = "update"
	AUTH_RULE_ASSIGN_UPDATES         string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_UPDATE
	AUTH_RULE_UPDATE_SELF            string = "update_self"
	AUTH_RULE_ASSIGN_UPDATE_SELF     string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_UPDATE_SELF
	AUTH_RULE_UPDATE_OTHERS          string = "update_others"
	AUTH_RULE_ASSIGN_UPDATE_OTHERS   string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_UPDATE_OTHERS
	AUTH_RULE_DELETE                 string = "delete"
	AUTH_RULE_ASSIGN_DELETE          string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_DELETE
	AUTH_RULE_DELETE_SELF            string = "delete_self"
	AUTH_RULE_ASSIGN_DELETE_SELF     string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_DELETE_SELF
	AUTH_RULE_DELETE_OTHERS          string = "delete_others"
	AUTH_RULE_ASSIGN_DELETE_OTHERS   string = AUTH_RULE_ASSIGN_PREFIX + AUTH_RULE_DELETE_OTHERS
	AUTH_RULE_ASSIGN_ALL             string = AUTH_RULE_ASSIGN_PREFIX + "all"
	AUTH_RULE_UPDATE_DIRECTORY       string = AUTH_RULE_UPDATE + "_" + AUTH_RULE_GROUP_DIRECTORY
)
