package main

import (
	"data_administration_platform/internal/pkg/lib"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-jet/jet/v2/generator/metadata"
	"github.com/go-jet/jet/v2/generator/postgres"
	"github.com/go-jet/jet/v2/generator/template"
	jet "github.com/go-jet/jet/v2/postgres"
	pq "github.com/lib/pq"
)

func main() {
	const currentSection = "Gen Database Code"
	dbCodeDirectory := os.Getenv("TABLES_MODELS_DIRECTORY")
	if dbCodeDirectory == "" {
		lib.Log(lib.LOG_FATAL, currentSection, "Invalid directory to put in database code")
	}

	port, err := strconv.Atoi(os.Getenv("PSQL_PORT"))
	if err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, fmt.Sprintf("Cannot set postgres port | reason: %v", err))
	}

	postgresConnection := postgres.DBConnection{
		Host:       os.Getenv("PSQL_HOST"),
		Port:       port,
		User:       os.Getenv("PSQL_USER"),
		Password:   os.Getenv("PSQL_PASS"),
		DBName:     os.Getenv("PSQL_DBNAME"),
		SchemaName: os.Getenv("PSQL_SCHEMA"),
		SslMode:    os.Getenv("PSQL_SSLMODE"),
	}

	if err = postgres.Generate(
		dbCodeDirectory,
		postgresConnection,
		template.Default(jet.Dialect).
			UseSchema(func(schemaMetaData metadata.Schema) template.Schema {
				return template.DefaultSchema(schemaMetaData).
					UseSQLBuilder(template.DefaultSQLBuilder().
						UseTable(func(table metadata.Table) template.TableSQLBuilder {
							if table.Name == "schema_migrations" {
								return template.TableSQLBuilder{
									Skip: true,
								}
							}
							return template.DefaultTableSQLBuilder(table)
						}),
					).UseModel(template.DefaultModel().
					UseTable(func(table metadata.Table) template.TableModel {
						var tableModel template.TableModel
						switch table.Name {
						case "schema_migrations":
							tableModel = template.TableModel{
								Skip: true,
							}
						default:
							tableModel = template.DefaultTableModel(table).
								UseField(func(columnMetaData metadata.Column) template.TableModelField {
									tableModelField := template.DefaultTableModelField(columnMetaData)
									if strings.Contains(columnMetaData.Name, "vector") || columnMetaData.Name == "password" || columnMetaData.Name == "ticket_number" || columnMetaData.Name == "pin" {
										tableModelField = tableModelField.UseTags(`json:"-"`)
									}
									if table.Name == "catalogue" && columnMetaData.Name == "catalogue" || table.Name == "directory" && columnMetaData.Name == "contacts" {
										tableModelField.Type = template.NewType(pq.StringArray{})
									}
									return tableModelField
								})
						}
						return tableModel
					}),
				)
			}),
	); err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, fmt.Sprintf("Failed to generate models and tables | reason: %v", err))
	}
}
