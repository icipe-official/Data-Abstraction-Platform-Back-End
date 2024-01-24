package projects

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
	RetrieveProject          RetrieveProject
	RetrieveProjects         []RetrieveProject
	DirectoryProjectRoles    model.DirectoryProjectsRoles
	Roles                    struct {
		DirectoryID  uuid.UUID
		ProjectID    uuid.UUID
		ProjectRoles []string
	}
}

type RetrieveProject struct {
	ID                     uuid.UUID `sql:"primary_key"`
	Name                   string
	Description            string
	CreatedOn              time.Time
	LastUpdatedOn          time.Time
	IsActive               bool
	OwnerDirectoryID       uuid.UUID
	OwnerDirectoryName     string
	OwnerDirectoryContacts pq.StringArray
	Storage                []Storage
}

type Storage struct {
	StorageID   uuid.UUID `sql:"primary_key"`
	StorageName string
}
