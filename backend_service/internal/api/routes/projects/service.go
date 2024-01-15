package projects

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"data_administration_platform/internal/pkg/data_administration_platform/public/table"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"fmt"
	"net/http"

	jet "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

func (n *projects) deleteRoles() (int, error) {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return -1, err
	}
	defer db.Close()

	rolesDeleted := 0
	for _, role := range n.Roles.ProjectRoles {
		deleteQuery := table.DirectoryProjectsRoles.DELETE().
			WHERE(
				table.DirectoryProjectsRoles.DirectoryID.EQ(jet.UUID(n.Roles.DirectoryID)).
					AND(table.DirectoryProjectsRoles.ProjectID.EQ(jet.UUID(n.Roles.ProjectID))).
					AND(table.DirectoryProjectsRoles.ProjectRoleID.EQ(jet.String(role))),
			)
		if _, err := deleteQuery.Exec(db); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete role %v for %v in %v by %v failed | reason: %v", role, n.Roles.DirectoryID, n.Roles.ProjectID, n.CurrentUser.DirectoryID, err))
		} else {
			rolesDeleted += 1
		}
	}

	return rolesDeleted, nil
}

func (n *projects) deleteProject() (int64, error) {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	deleteQuery := table.Projects.DELETE().WHERE(table.Projects.ID.EQ(jet.UUID(n.ProjectID)).AND(table.Projects.DirectoryID.EQ(jet.UUID(n.CurrentUser.DirectoryID))))

	if sqlResults, err := deleteQuery.Exec(db); err != nil {
		updateQuery := table.Projects.
			UPDATE(table.Projects.IsActive).
			MODEL(model.Projects{IsActive: false}).
			WHERE(table.Projects.ID.EQ(jet.UUID(n.ProjectID)).AND(table.Projects.DirectoryID.EQ(jet.UUID(n.CurrentUser.DirectoryID))))
		if sqlResults, err := updateQuery.Exec(db); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Deactivate project %v by %v failed | reason: %v", n.ProjectID, n.CurrentUser.DirectoryID, err))
			return -1, lib.NewError(http.StatusInternalServerError, "Could not deactivate project")
		} else {
			if deletedRows, err := sqlResults.RowsAffected(); err != nil {
				intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Determining no. of projects deleted failed while deleting %v | reason: %v", n.ProjectID, err))
				return -1, nil
			} else {
				return deletedRows, err
			}
		}
	} else {
		if deletedRows, err := sqlResults.RowsAffected(); err != nil {
			intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Determining no. of projects deleted failed while deleting %v | reason: %v", n.ProjectID, err))
			return -1, nil
		} else {
			return deletedRows, err
		}
	}
}

func (n *projects) updateProject() error {
	columnsToUpdate := make(jet.ColumnList, 0)
	for _, column := range n.ProjectUpdate.Columns {
		switch column {
		case table.Projects.Name.Name():
			if len(n.ProjectUpdate.Project.Name) >= 3 {
				columnsToUpdate = append(columnsToUpdate, table.Projects.Name)
			}
		case table.Projects.Description.Name():
			if len(n.ProjectUpdate.Project.Description) >= 3 {
				columnsToUpdate = append(columnsToUpdate, table.Projects.Description)
			}
		}
	}

	if len(columnsToUpdate) < 1 {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	updateQuery := table.Projects.
		UPDATE(columnsToUpdate).
		MODEL(n.ProjectUpdate.Project).
		WHERE(table.Projects.ID.EQ(jet.UUID(n.ProjectID))).
		RETURNING(table.Projects.ID, table.Projects.LastUpdatedOn)

	if err = updateQuery.Query(db, &n.ProjectUpdate.Project); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Update project %v by %v failed | reason: %v", n.ProjectID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not update project")
	}

	return nil
}

func (n *projects) createRoles() (int64, error) {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return -1, err
	}
	defer db.Close()

	NewDirectoryProjectRoles := []model.DirectoryProjectsRoles{}
	for _, role := range n.Roles.ProjectRoles {
		NewDirectoryProjectRoles = append(NewDirectoryProjectRoles, model.DirectoryProjectsRoles{DirectoryID: n.Roles.DirectoryID, ProjectID: n.Roles.ProjectID, ProjectRoleID: role})
	}

	insertQuery := table.DirectoryProjectsRoles.
		INSERT(table.DirectoryProjectsRoles.ProjectID, table.DirectoryProjectsRoles.DirectoryID, table.DirectoryProjectsRoles.ProjectRoleID).
		MODELS(NewDirectoryProjectRoles)

	if sqlResults, err := insertQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Add role(s) for %v in project %v by %v failed | reason: %v", n.Roles.DirectoryID, n.Roles.ProjectID, n.CurrentUser.DirectoryID, err))
		return -1, lib.NewError(http.StatusInternalServerError, "Could not add new roles")
	} else {
		if rowsAffected, err := sqlResults.RowsAffected(); err != nil {
			intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Could not determine no of rows inserted when adding project roles | reason: %v", err))
			return -1, nil
		} else {
			return rowsAffected, nil
		}
	}
}

func (n *projects) getProjects() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	var selectQuery jet.SelectStatement
	if n.QuickSearch == "true" {
		selectQuery = jet.SELECT(table.Projects.ID, table.Projects.Name, table.Projects.Description).
			FROM(table.Projects)
	} else {
		selectQuery = jet.SELECT(table.Projects.AllColumns, table.Directory.Name, table.Directory.Contacts).
			FROM(table.Projects.INNER_JOIN(table.Directory, table.Projects.DirectoryID.EQ(table.Directory.ID)))
	}
	if n.ProjectID != uuid.Nil {
		selectQuery = selectQuery.WHERE(table.Projects.ID.EQ(jet.UUID(n.ProjectID)))
		n.ProjectDirectory = projectsDirectory{}
		if err = selectQuery.Query(db, &n.ProjectDirectory); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get project %v by %v failed | reason: %v", n.ProjectID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get project")
		}
	} else {
		var whereCondition jet.BoolExpression
		isWhereClauseSet := false
		if n.SearchQuery != "" {
			whereCondition = jet.BoolExp(lib.GetTextSearchBoolExp(table.Projects.ProjectsVector.Name(), n.SearchQuery))
			isWhereClauseSet = true
		}
		if n.CreatedOnGreaterThan != "" {
			cogtCondition := table.Projects.CreatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnGreaterThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(cogtCondition)
			} else {
				whereCondition = cogtCondition
				isWhereClauseSet = true
			}
		}
		if n.CreatedOnLessThan != "" {
			coltCondition := table.Projects.CreatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnLessThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(coltCondition)
			} else {
				whereCondition = coltCondition
				isWhereClauseSet = true
			}
		}
		if n.LastUpdatedOnGreaterThan != "" {
			lugtCondition := table.Projects.LastUpdatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnGreaterThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(lugtCondition)
			} else {
				whereCondition = lugtCondition
				isWhereClauseSet = true
			}
		}
		if n.LastUpdatedOnLessThan != "" {
			lultCondition := table.Projects.LastUpdatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnLessThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(lultCondition)
			} else {
				whereCondition = lultCondition
				isWhereClauseSet = true
			}
		}
		if n.IsActive != "" {
			var iaCondition jet.BoolExpression
			if n.IsActive == "true" {
				iaCondition = table.Projects.IsActive.IS_TRUE()
			} else {
				iaCondition = table.Projects.IsActive.IS_FALSE()
			}
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(iaCondition)
			} else {
				whereCondition = iaCondition
				isWhereClauseSet = true
			}
		}
		if isWhereClauseSet {
			selectQuery = selectQuery.WHERE(whereCondition)
		}
		if n.Limit > 0 {
			selectQuery = selectQuery.LIMIT(int64(n.Limit))
		}
		if n.Offset > 0 {
			selectQuery = selectQuery.OFFSET(int64(n.Offset))
		}
		n.ProjectsDirectory = []projectsDirectory{}
		if err = selectQuery.Query(db, &n.ProjectsDirectory); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get projects by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get projects")
		}
	}

	return nil
}

func (n *projects) createProject() error {
	if len(n.Project.Name) < 3 || len(n.Project.Description) < 3 || n.Project.DirectoryID == uuid.Nil {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	insertQuery := table.Projects.
		INSERT(table.Projects.DirectoryID, table.Projects.Name, table.Projects.Description).
		MODEL(n.Project).
		RETURNING(table.Projects.ID)

	n.DirectoryProjectRoles = model.DirectoryProjectsRoles{
		DirectoryID:   n.Project.DirectoryID,
		ProjectRoleID: lib.ROLE_PROJECT_ADMIN,
	}
	if err = insertQuery.Query(db, &n.Project); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Creating project by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not create new project")
	}

	n.DirectoryProjectRoles.ProjectID = n.Project.ID
	insertQuery = table.DirectoryProjectsRoles.
		INSERT(table.DirectoryProjectsRoles.DirectoryID, table.DirectoryProjectsRoles.ProjectID, table.DirectoryProjectsRoles.ProjectRoleID).
		MODEL(n.DirectoryProjectRoles).
		RETURNING(table.DirectoryProjectsRoles.CreatedOn)
	if _, err = insertQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Assign user %v project_admin role by %v failed | reason: %v", n.DirectoryProjectRoles.DirectoryID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not assign project_admin role to project owner")
	}

	return nil
}
