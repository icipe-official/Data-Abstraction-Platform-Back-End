package catalogue

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"

	"github.com/google/uuid"
)

const currentSection = "Catalogue"

type catalogue struct {
	SearchQuery              string    `json:"-"`
	CreatedOnGreaterThan     string    `json:"-"`
	CreatedOnLessThan        string    `json:"-"`
	CatalogueID              uuid.UUID `json:"-"`
	ProjectID                uuid.UUID `json:"-"`
	LastUpdatedOnGreaterThan string    `json:"-"`
	LastUpdatedOnLessThan    string    `json:"-"`
	CanPublicView            string    `json:"-"`
	Limit                    int       `json:"-"`
	Offset                   int       `json:"-"`
	QuickSearch              string    `json:"-"`
	SortyBy                  string    `json:"-"`
	SortByOrder              string    `json:"-"`
	CatalogueUpdate          struct {
		Catalogue model.Catalogue
		Columns   []string
	}
	CurrentUser                lib.User
	Catalogue                  model.Catalogue
	CataloguesDirectoryProject []catalogueDirectoryProject
	CatalogueDirectoryProject  catalogueDirectoryProject
}

type catalogueDirectoryProject struct {
	model.Catalogue
	Directory model.Directory
	Project   model.Projects
}
