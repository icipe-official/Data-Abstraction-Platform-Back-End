package initdatabase

import (
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"data_administration_platform/internal/pkg/data_administration_platform/public/table"
	"data_administration_platform/internal/pkg/lib"
	"database/sql"
	"fmt"
)

var projectRoles = []*model.ProjectRoles{
	{
		ID:          "project_admin",
		Description: "Manage users in a project",
	},
	{
		ID:          "model_templates_creator",
		Description: "Can create model templates",
	},
	{
		ID:          "catalogue_creator",
		Description: "Can create catalogues and manage their own catalogues",
	},
	{
		ID:          "file_creator",
		Description: "Can upload files",
	},
	{
		ID:          "abstractions_admin",
		Description: "Can create abstractions, assign, and reassign abstractions",
	},
	{
		ID:          "editor",
		Description: "Can edit abstractions or respond to surveys",
	},
	{
		ID:          "reviewer",
		Description: "Can review abstractions or surveys",
	},
	{
		ID:          "explorer",
		Description: "Can consume data in a project",
	},
}

func InsertProjectRoles(db *sql.DB) (int64, error) {
	insertQuery := table.ProjectRoles.INSERT(table.ProjectRoles.AllColumns).MODELS(projectRoles).ON_CONFLICT(table.ProjectRoles.ID).DO_NOTHING()
	sqlResults, err := insertQuery.Exec(db)
	if err != nil {
		return 0, fmt.Errorf("could not insert project roles | reason: %v", err)
	}
	rowsAffected, err := sqlResults.RowsAffected()
	if err != nil {
		lib.Log(lib.LOG_WARNING, "Project Roles", fmt.Sprintf("Could not determine number of rows inserted for project roles | reason: %v", err))
	}
	return rowsAffected, nil
}
