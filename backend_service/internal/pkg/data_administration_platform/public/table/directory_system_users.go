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

var DirectorySystemUsers = newDirectorySystemUsersTable("public", "directory_system_users", "")

type directorySystemUsersTable struct {
	postgres.Table

	// Columns
	DirectoryID postgres.ColumnString
	CreatedOn   postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type DirectorySystemUsersTable struct {
	directorySystemUsersTable

	EXCLUDED directorySystemUsersTable
}

// AS creates new DirectorySystemUsersTable with assigned alias
func (a DirectorySystemUsersTable) AS(alias string) *DirectorySystemUsersTable {
	return newDirectorySystemUsersTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new DirectorySystemUsersTable with assigned schema name
func (a DirectorySystemUsersTable) FromSchema(schemaName string) *DirectorySystemUsersTable {
	return newDirectorySystemUsersTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new DirectorySystemUsersTable with assigned table prefix
func (a DirectorySystemUsersTable) WithPrefix(prefix string) *DirectorySystemUsersTable {
	return newDirectorySystemUsersTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new DirectorySystemUsersTable with assigned table suffix
func (a DirectorySystemUsersTable) WithSuffix(suffix string) *DirectorySystemUsersTable {
	return newDirectorySystemUsersTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newDirectorySystemUsersTable(schemaName, tableName, alias string) *DirectorySystemUsersTable {
	return &DirectorySystemUsersTable{
		directorySystemUsersTable: newDirectorySystemUsersTableImpl(schemaName, tableName, alias),
		EXCLUDED:                  newDirectorySystemUsersTableImpl("", "excluded", ""),
	}
}

func newDirectorySystemUsersTableImpl(schemaName, tableName, alias string) directorySystemUsersTable {
	var (
		DirectoryIDColumn = postgres.StringColumn("directory_id")
		CreatedOnColumn   = postgres.TimestampzColumn("created_on")
		allColumns        = postgres.ColumnList{DirectoryIDColumn, CreatedOnColumn}
		mutableColumns    = postgres.ColumnList{CreatedOnColumn}
	)

	return directorySystemUsersTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		DirectoryID: DirectoryIDColumn,
		CreatedOn:   CreatedOnColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
