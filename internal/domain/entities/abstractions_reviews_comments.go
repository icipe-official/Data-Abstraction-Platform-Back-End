package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type AbstractionsReviewsComments struct {
	ID             []uuid.UUID `json:"id,omitempty"`
	AbstractionsID []uuid.UUID `json:"abstractions_id,omitempty"`
	DirectoryID    []uuid.UUID `json:"directory_id,omitempty"`
	Comment        []string    `json:"comment,omitempty"`
	CreatedOn      []time.Time `json:"created_on,omitempty"`
}

type abstractionsReviewsCommentsRepository struct {
	RepositoryName string

	ID             string
	AbstractionsID string
	DirectoryID    string
	Comment        string
	CreatedOn      string
}

func AbstractionsReviewsCommentsRepository() abstractionsReviewsCommentsRepository {
	return abstractionsReviewsCommentsRepository{
		RepositoryName: "abstractions_reviews_comments",

		ID:             "id",
		AbstractionsID: "abstractions_id",
		DirectoryID:    "directory_id",
		Comment:        "comment",
		CreatedOn:      "created_on",
	}
}
