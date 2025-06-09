package entities

import (
	"time"
)

type GroupAuthorizationRules struct {
	GroupAuthorizationRulesID []struct {
		ID        []string `sql:"primary_key" json:"id,omitempty"`
		RuleGroup []string `sql:"primary_key" json:"rule_group,omitempty"`
	} `json:"group_authorization_rules_id,omitempty"`
	Description   []string    `json:"description,omitempty"`
	CreatedOn     []time.Time `json:"created_on,omitempty"`
	LastUpdatedOn []time.Time `json:"last_updated_on,omitempty"`
}

type groupAuthorizationRulesRepository struct {
	RepositoryName string

	ID             string
	RuleGroup      string
	Description    string
	CreatedOn      string
	LastUpdatedOn  string
	FullTextSearch string
}

func GroupAuthorizationRulesRepository() groupAuthorizationRulesRepository {
	return groupAuthorizationRulesRepository{
		RepositoryName: "group_authorization_rules",

		ID:             "id",
		RuleGroup:      "rule_group",
		Description:    "description",
		CreatedOn:      "created_on",
		LastUpdatedOn:  "last_updated_on",
		FullTextSearch: "full_text_search",
	}
}
