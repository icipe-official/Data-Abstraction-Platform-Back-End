//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"github.com/google/uuid"
	"time"
)

type Abstractions struct {
	ID              uuid.UUID `sql:"primary_key"`
	ModelTemplateID uuid.UUID
	FileID          uuid.UUID
	DirectoryID     uuid.UUID
	ProjectID       uuid.UUID
	Tags            *string
	Abstraction     string
	IsVerified      bool
	CreatedOn       time.Time
	LastUpdatedOn   time.Time
}