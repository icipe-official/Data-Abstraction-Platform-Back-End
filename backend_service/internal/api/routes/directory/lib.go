package directory

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const currentSection = "Directory"

type directory struct {
	DirectoryUpdate struct {
		Directory    model.Directory
		IsSystemUser string
		Columns      []string
	}
	SearchQuery              string    `json:"-"`
	CreatedOnGreaterThan     string    `json:"-"`
	CreatedOnLessThan        string    `json:"-"`
	DirectoryID              uuid.UUID `json:"-"`
	ProjectRole              string    `json:"-"`
	ProjectID                uuid.UUID `json:"-"`
	LastUpdatedOnGreaterThan string    `json:"-"`
	LastUpdatedOnLessThan    string    `json:"-"`
	Limit                    int       `json:"-"`
	Offset                   int       `json:"-"`
	QuickSearch              string    `json:"-"`
	SortyBy                  string    `json:"-"`
	SortByOrder              string    `json:"-"`
	CurrentUser              lib.User
	Directory                model.Directory
	DirectoryCreate          directoryCreate `json:"-"`
	RetrieveUser             RetrieveUser
	RetrieveUsers            []RetrieveUser
}

type RetrieveUser struct {
	ID                  uuid.UUID `sql:"primary_key"`
	Name                string
	Contacts            pq.StringArray
	CreatedOn           time.Time
	LastUpdatedOn       time.Time
	SystemUserCreatedOn time.Time
	IamEmail            *string
	IamCreatedOn        time.Time
	IamLastUpdatedOn    time.Time
	ProjectsRoles       []Project
}

type Project struct {
	ProjectID          uuid.UUID `sql:"primary_key"`
	ProjectName        string
	ProjectDescription string
	ProjectCreatedOn   string
	ProjectRoles       []ProjectRoles
}

type ProjectRoles struct {
	ProjectRoleID        string `sql:"primary_key"`
	ProjectRoleCreatedOn time.Time
}

type directoryCreate struct {
	model.Directory
	Email        string
	IsSystemUser bool
	ProjectID    uuid.UUID
}
