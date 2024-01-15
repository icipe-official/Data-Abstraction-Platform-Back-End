//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"time"
)

type Directory struct {
	ID              uuid.UUID `sql:"primary_key"`
	Name            string
	Contacts        pq.StringArray
	CreatedOn       time.Time
	LastUpdatedOn   time.Time
	DirectoryVector string `json:"-"`
}
