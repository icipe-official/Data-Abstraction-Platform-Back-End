//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var AbstractionReviews = newAbstractionReviewsTable("public", "abstraction_reviews", "")

type abstractionReviewsTable struct {
	postgres.Table

	// Columns
	AbstractionID postgres.ColumnString
	DirectoryID   postgres.ColumnString
	Review        postgres.ColumnBool
	CreatedOn     postgres.ColumnTimestampz
	LastUpdatedOn postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type AbstractionReviewsTable struct {
	abstractionReviewsTable

	EXCLUDED abstractionReviewsTable
}

// AS creates new AbstractionReviewsTable with assigned alias
func (a AbstractionReviewsTable) AS(alias string) *AbstractionReviewsTable {
	return newAbstractionReviewsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new AbstractionReviewsTable with assigned schema name
func (a AbstractionReviewsTable) FromSchema(schemaName string) *AbstractionReviewsTable {
	return newAbstractionReviewsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new AbstractionReviewsTable with assigned table prefix
func (a AbstractionReviewsTable) WithPrefix(prefix string) *AbstractionReviewsTable {
	return newAbstractionReviewsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new AbstractionReviewsTable with assigned table suffix
func (a AbstractionReviewsTable) WithSuffix(suffix string) *AbstractionReviewsTable {
	return newAbstractionReviewsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newAbstractionReviewsTable(schemaName, tableName, alias string) *AbstractionReviewsTable {
	return &AbstractionReviewsTable{
		abstractionReviewsTable: newAbstractionReviewsTableImpl(schemaName, tableName, alias),
		EXCLUDED:                newAbstractionReviewsTableImpl("", "excluded", ""),
	}
}

func newAbstractionReviewsTableImpl(schemaName, tableName, alias string) abstractionReviewsTable {
	var (
		AbstractionIDColumn = postgres.StringColumn("abstraction_id")
		DirectoryIDColumn   = postgres.StringColumn("directory_id")
		ReviewColumn        = postgres.BoolColumn("review")
		CreatedOnColumn     = postgres.TimestampzColumn("created_on")
		LastUpdatedOnColumn = postgres.TimestampzColumn("last_updated_on")
		allColumns          = postgres.ColumnList{AbstractionIDColumn, DirectoryIDColumn, ReviewColumn, CreatedOnColumn, LastUpdatedOnColumn}
		mutableColumns      = postgres.ColumnList{ReviewColumn, CreatedOnColumn, LastUpdatedOnColumn}
	)

	return abstractionReviewsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		AbstractionID: AbstractionIDColumn,
		DirectoryID:   DirectoryIDColumn,
		Review:        ReviewColumn,
		CreatedOn:     CreatedOnColumn,
		LastUpdatedOn: LastUpdatedOnColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}