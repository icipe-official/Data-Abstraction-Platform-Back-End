package projects

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"

	"github.com/google/uuid"
)

const currentSection = "Projects"

type projects struct {
	ProjectUpdate struct {
		Project model.Projects
		Columns []string
	}
	SearchQuery              string    `json:"-"`
	CreatedOnGreaterThan     string    `json:"-"`
	CreatedOnLessThan        string    `json:"-"`
	ProjectID                uuid.UUID `json:"-"`
	LastUpdatedOnGreaterThan string    `json:"-"`
	LastUpdatedOnLessThan    string    `json:"-"`
	IsActive                 string    `json:"-"`
	Limit                    int       `json:"-"`
	Offset                   int       `json:"-"`
	QuickSearch              string    `json:"-"`
	CurrentUser              lib.User
	Project                  model.Projects
	ProjectDirectory         projectsDirectory
	ProjectsDirectory        []projectsDirectory
	DirectoryProjectRoles    model.DirectoryProjectsRoles
	Roles                    struct {
		DirectoryID  uuid.UUID
		ProjectID    uuid.UUID
		ProjectRoles []string
	}
}

type projectsDirectory struct {
	model.Projects
	Directory model.Directory
}
