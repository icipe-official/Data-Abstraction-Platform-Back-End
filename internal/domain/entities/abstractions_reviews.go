package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type AbstractionsReviews struct {
	AbstractionsID []uuid.UUID `json:"abstractions_id,omitempty"`
	DirectoryID    []uuid.UUID `json:"directory_id,omitempty"`
	ReviewPass     []bool      `json:"review_pass,omitempty"`
	CreatedOn      []time.Time `json:"created_on,omitempty"`
	LastUpdatedOn  []time.Time `json:"last_updated_on,omitempty"`
}

type abstractionsReviewsRepository struct {
	RepositoryName string

	AbstractionsID string
	DirectoryID    string
	ReviewPass     string
	CreatedOn      string
	LastUpdatedOn  string
}

func AbstractionsReviewsRepository() abstractionsReviewsRepository {
	return abstractionsReviewsRepository{
		RepositoryName: "abstractions_reviews",

		AbstractionsID: "abstractions_id",
		DirectoryID:    "directory_id",
		ReviewPass:     "review_pass",
		CreatedOn:      "created_on",
		LastUpdatedOn:  "last_updated_on",
	}
}
