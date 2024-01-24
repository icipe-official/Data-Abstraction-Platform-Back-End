//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package view

import (
	"github.com/go-jet/jet/v2/postgres"
)

var PlatformStatistics = newPlatformStatisticsTable("public", "platform_statistics", "")

type platformStatisticsTable struct {
	postgres.Table

	// Columns
	NoOfProjects       postgres.ColumnInteger
	NoOfModelTemplates postgres.ColumnInteger
	NoOfCatalogues     postgres.ColumnInteger
	NoOfAbstractions   postgres.ColumnInteger

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type PlatformStatisticsTable struct {
	platformStatisticsTable

	EXCLUDED platformStatisticsTable
}

// AS creates new PlatformStatisticsTable with assigned alias
func (a PlatformStatisticsTable) AS(alias string) *PlatformStatisticsTable {
	return newPlatformStatisticsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new PlatformStatisticsTable with assigned schema name
func (a PlatformStatisticsTable) FromSchema(schemaName string) *PlatformStatisticsTable {
	return newPlatformStatisticsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new PlatformStatisticsTable with assigned table prefix
func (a PlatformStatisticsTable) WithPrefix(prefix string) *PlatformStatisticsTable {
	return newPlatformStatisticsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new PlatformStatisticsTable with assigned table suffix
func (a PlatformStatisticsTable) WithSuffix(suffix string) *PlatformStatisticsTable {
	return newPlatformStatisticsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newPlatformStatisticsTable(schemaName, tableName, alias string) *PlatformStatisticsTable {
	return &PlatformStatisticsTable{
		platformStatisticsTable: newPlatformStatisticsTableImpl(schemaName, tableName, alias),
		EXCLUDED:                newPlatformStatisticsTableImpl("", "excluded", ""),
	}
}

func newPlatformStatisticsTableImpl(schemaName, tableName, alias string) platformStatisticsTable {
	var (
		NoOfProjectsColumn       = postgres.IntegerColumn("no_of_projects")
		NoOfModelTemplatesColumn = postgres.IntegerColumn("no_of_model_templates")
		NoOfCataloguesColumn     = postgres.IntegerColumn("no_of_catalogues")
		NoOfAbstractionsColumn   = postgres.IntegerColumn("no_of_abstractions")
		allColumns               = postgres.ColumnList{NoOfProjectsColumn, NoOfModelTemplatesColumn, NoOfCataloguesColumn, NoOfAbstractionsColumn}
		mutableColumns           = postgres.ColumnList{NoOfProjectsColumn, NoOfModelTemplatesColumn, NoOfCataloguesColumn, NoOfAbstractionsColumn}
	)

	return platformStatisticsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		NoOfProjects:       NoOfProjectsColumn,
		NoOfModelTemplates: NoOfModelTemplatesColumn,
		NoOfCatalogues:     NoOfCataloguesColumn,
		NoOfAbstractions:   NoOfAbstractionsColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}