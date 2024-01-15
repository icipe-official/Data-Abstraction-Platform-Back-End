package modeltemplates

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"

	"github.com/google/uuid"
)

const currentSection = "ModelTemplates"

type modeltemplates struct {
	SearchQuery              string    `json:"-"`
	CreatedOnGreaterThan     string    `json:"-"`
	CreatedOnLessThan        string    `json:"-"`
	ModelTemplateID          uuid.UUID `json:"-"`
	ProjectID                uuid.UUID `json:"-"`
	LastUpdatedOnGreaterThan string    `json:"-"`
	LastUpdatedOnLessThan    string    `json:"-"`
	CanPublicView            string    `json:"-"`
	Limit                    int       `json:"-"`
	Offset                   int       `json:"-"`
	QuickSearch              string    `json:"-"`
	SortyBy                  string    `json:"-"`
	SortByOrder              string    `json:"-"`
	ModelTemplateUpdate      struct {
		ModelTemplate model.ModelTemplates
		Columns       []string
	}
	CurrentUser                    lib.User
	ModelTemplate                  model.ModelTemplates
	ModelTemplatesDirectoryProject []modelTemplateDirectoryProject
	ModelTemplateDirectoryProject  modelTemplateDirectoryProject
}

type modelTemplateDirectoryProject struct {
	model.ModelTemplates
	Directory model.Directory
	Project   model.Projects
}
