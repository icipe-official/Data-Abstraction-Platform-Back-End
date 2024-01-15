package modeltemplates

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/table"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"fmt"
	"net/http"

	jet "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

func (n *modeltemplates) deleteModelTemplate() (int64, error) {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	deleteQuery := table.ModelTemplates.DELETE().WHERE(table.ModelTemplates.ID.EQ(jet.UUID(n.ModelTemplateID)).AND(table.ModelTemplates.DirectoryID.EQ(jet.UUID(n.CurrentUser.DirectoryID))))

	if sqlResults, err := deleteQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete modeltemplate %v by %v failed | reason: %v", n.ModelTemplateID, n.CurrentUser.DirectoryID, err))
		return -1, lib.NewError(http.StatusInternalServerError, "Could not delete modeltemplate")
	} else {
		if deletedRows, err := sqlResults.RowsAffected(); err != nil {
			intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Determining no. of modeltemplates deleted failed while deleting %v | reason: %v", n.ModelTemplateID, err))
			return -1, nil
		} else {
			return deletedRows, err
		}
	}
}

func (n *modeltemplates) getModelTemplates() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	var selectQuery jet.SelectStatement
	if n.QuickSearch == "true" {
		selectQuery = jet.SELECT(table.ModelTemplates.ID, table.ModelTemplates.TemplateName, table.ModelTemplates.DataName, table.ModelTemplates.Description).
			FROM(table.ModelTemplates)
	} else {
		selectQuery = jet.SELECT(table.ModelTemplates.AllColumns, table.Directory.Name, table.Directory.Contacts, table.Projects.Name, table.Projects.Description).
			FROM(table.ModelTemplates.INNER_JOIN(table.Directory, table.ModelTemplates.DirectoryID.EQ(table.Directory.ID)).INNER_JOIN(table.Projects, table.ModelTemplates.ProjectID.EQ(table.Projects.ID)))
	}

	if n.ModelTemplateID != uuid.Nil {
		selectQuery = selectQuery.WHERE(table.ModelTemplates.ID.EQ(jet.UUID(n.ModelTemplateID)))
		n.ModelTemplateDirectoryProject = modelTemplateDirectoryProject{}
		if err = selectQuery.Query(db, &n.ModelTemplateDirectoryProject); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get modeltemplate %v by %v failed | reason: %v", n.ModelTemplateID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get modeltemplate")
		}
	} else {
		var whereCondition jet.BoolExpression
		isWhereClauseSet := false
		if n.SearchQuery != "" {
			whereCondition = jet.BoolExp(lib.GetTextSearchBoolExp(table.ModelTemplates.ModelTemplateVector.Name(), n.SearchQuery))
			isWhereClauseSet = true
		}
		if n.CreatedOnGreaterThan != "" {
			cogtCondition := table.ModelTemplates.CreatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnGreaterThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(cogtCondition)
			} else {
				whereCondition = cogtCondition
				isWhereClauseSet = true
			}
		}
		if n.CreatedOnLessThan != "" {
			coltCondition := table.ModelTemplates.CreatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnLessThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(coltCondition)
			} else {
				whereCondition = coltCondition
				isWhereClauseSet = true
			}
		}
		if n.LastUpdatedOnGreaterThan != "" {
			lugtCondition := table.ModelTemplates.LastUpdatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnGreaterThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(lugtCondition)
			} else {
				whereCondition = lugtCondition
				isWhereClauseSet = true
			}
		}
		if n.LastUpdatedOnLessThan != "" {
			lultCondition := table.ModelTemplates.LastUpdatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnLessThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(lultCondition)
			} else {
				whereCondition = lultCondition
				isWhereClauseSet = true
			}
		}
		if n.CanPublicView != "" {
			var iaCondition jet.BoolExpression
			if n.CanPublicView == "true" {
				iaCondition = table.ModelTemplates.CanPublicView.IS_TRUE()
			} else {
				iaCondition = table.ModelTemplates.CanPublicView.IS_FALSE()
			}
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(iaCondition)
			} else {
				whereCondition = iaCondition
				isWhereClauseSet = true
			}
		}
		if n.ProjectID != uuid.Nil {
			pidCondition := table.ModelTemplates.ProjectID.EQ(jet.UUID(n.ProjectID))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(pidCondition)
			} else {
				whereCondition = pidCondition
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
		if n.SortyBy != "" {
			switch n.SortyBy {
			case table.ModelTemplates.CreatedOn.Name():
				if n.SortByOrder == "asc" {
					selectQuery = selectQuery.ORDER_BY(table.ModelTemplates.CreatedOn.ASC())
				} else {
					selectQuery = selectQuery.ORDER_BY(table.ModelTemplates.CreatedOn.DESC())
				}
			case table.ModelTemplates.LastUpdatedOn.Name():
				if n.SortByOrder == "asc" {
					selectQuery = selectQuery.ORDER_BY(table.ModelTemplates.LastUpdatedOn.ASC())
				} else {
					selectQuery = selectQuery.ORDER_BY(table.ModelTemplates.LastUpdatedOn.DESC())
				}
			}
		}
		n.ModelTemplatesDirectoryProject = []modelTemplateDirectoryProject{}
		if err = selectQuery.Query(db, &n.ModelTemplatesDirectoryProject); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get modeltemplates by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get modeltemplates")
		}
	}

	return nil
}

func (n *modeltemplates) updateModelTemplate() error {
	columnsToUpdate := make(jet.ColumnList, 0)
	for _, column := range n.ModelTemplateUpdate.Columns {
		switch column {
		case table.ModelTemplates.TemplateName.Name():
			if len(n.ModelTemplateUpdate.ModelTemplate.TemplateName) >= 3 {
				columnsToUpdate = append(columnsToUpdate, table.ModelTemplates.TemplateName)
			}
		case table.ModelTemplates.Description.Name():
			if len(n.ModelTemplateUpdate.ModelTemplate.Description) >= 3 {
				columnsToUpdate = append(columnsToUpdate, table.ModelTemplates.Description)
			}
		case table.ModelTemplates.DataName.Name():
			if len(n.ModelTemplateUpdate.ModelTemplate.DataName) >= 3 {
				columnsToUpdate = append(columnsToUpdate, table.ModelTemplates.DataName)
			}
		case table.ModelTemplates.ModelTemplate.Name():
			columnsToUpdate = append(columnsToUpdate, table.ModelTemplates.ModelTemplate)
		case table.ModelTemplates.CanPublicView.Name():
			columnsToUpdate = append(columnsToUpdate, table.ModelTemplates.CanPublicView)
		case table.ModelTemplates.VerificationQuorum.Name():
			columnsToUpdate = append(columnsToUpdate, table.ModelTemplates.VerificationQuorum)
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

	updateQuery := table.ModelTemplates.
		UPDATE(columnsToUpdate).
		MODEL(n.ModelTemplateUpdate.ModelTemplate).
		WHERE(
			table.ModelTemplates.ID.EQ(jet.UUID(n.ModelTemplateID)).
				AND(table.ModelTemplates.ProjectID.EQ(jet.UUID(n.ProjectID)).
					AND(table.ModelTemplates.DirectoryID.EQ(jet.UUID(n.CurrentUser.DirectoryID))),
				)).
		RETURNING(table.ModelTemplates.ID, table.ModelTemplates.LastUpdatedOn)

	if err := updateQuery.Query(db, &n.ModelTemplateUpdate.ModelTemplate); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Update modeltemplate %v by %v failed | reason: %v", n.ModelTemplateID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not update modeltemplate")
	}

	return nil
}

func (n *modeltemplates) createModelTemplate() error {
	if len(n.ModelTemplate.TemplateName) < 3 || len(n.ModelTemplate.Description) < 3 || len(n.ModelTemplate.DataName) < 3 {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	n.ModelTemplate.DirectoryID = n.CurrentUser.DirectoryID

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	insertQuery := table.ModelTemplates.
		INSERT(
			table.ModelTemplates.ProjectID,
			table.ModelTemplates.DirectoryID,
			table.ModelTemplates.TemplateName,
			table.ModelTemplates.DataName,
			table.ModelTemplates.Description,
			table.ModelTemplates.ModelTemplate,
		).MODEL(n.ModelTemplate).RETURNING(table.ModelTemplates.ID)

	if err := insertQuery.Query(db, &n.ModelTemplate); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Create modeltemplate by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not create modeltempate")
	}

	return nil
}
