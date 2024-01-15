package catalogue

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/table"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"fmt"
	"net/http"

	jet "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

func (n *catalogue) deleteCatalogue() (int64, error) {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	deleteQuery := table.Catalogue.DELETE().WHERE(table.Catalogue.ID.EQ(jet.UUID(n.CatalogueID)).AND(table.Catalogue.DirectoryID.EQ(jet.UUID(n.CurrentUser.DirectoryID))))

	if sqlResults, err := deleteQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete catalogue %v by %v failed | reason: %v", n.CatalogueID, n.CurrentUser.DirectoryID, err))
		return -1, lib.NewError(http.StatusInternalServerError, "Could not delete catalogue")
	} else {
		if deletedRows, err := sqlResults.RowsAffected(); err != nil {
			intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Determining no. of catalogues deleted failed while deleting %v | reason: %v", n.CatalogueID, err))
			return -1, nil
		} else {
			return deletedRows, err
		}
	}
}

func (n *catalogue) getCatalogue() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	var selectQuery jet.SelectStatement

	if n.QuickSearch == "true" {
		selectQuery = jet.SELECT(table.Catalogue.ID, table.Catalogue.Name, table.Catalogue.Description).FROM(table.Catalogue)
	} else {
		selectQuery = jet.SELECT(table.Catalogue.AllColumns, table.Directory.Name, table.Directory.Contacts, table.Projects.Name, table.Projects.Description).
			FROM(table.Catalogue.INNER_JOIN(table.Directory, table.Catalogue.DirectoryID.EQ(table.Directory.ID)).INNER_JOIN(table.Projects, table.Catalogue.ProjectID.EQ(table.Projects.ID)))
	}

	if n.CatalogueID != uuid.Nil {
		selectQuery = selectQuery.WHERE(table.Catalogue.ID.EQ(jet.UUID(n.CatalogueID)))
		n.CatalogueDirectoryProject = catalogueDirectoryProject{}
		if err = selectQuery.Query(db, &n.CatalogueDirectoryProject); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get catalogue %v by %v failed | reason: %v", n.CatalogueID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get catalogue")
		}
	} else {
		var whereCondition jet.BoolExpression
		isWhereClauseSet := false
		if n.SearchQuery != "" {
			whereCondition = jet.BoolExp(lib.GetTextSearchBoolExp(table.Catalogue.CatalogueVector.Name(), n.SearchQuery))
			isWhereClauseSet = true
		}
		if n.CreatedOnGreaterThan != "" {
			cogtCondition := table.Catalogue.CreatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnGreaterThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(cogtCondition)
			} else {
				whereCondition = cogtCondition
				isWhereClauseSet = true
			}
		}
		if n.CreatedOnLessThan != "" {
			coltCondition := table.Catalogue.CreatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnLessThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(coltCondition)
			} else {
				whereCondition = coltCondition
				isWhereClauseSet = true
			}
		}
		if n.LastUpdatedOnGreaterThan != "" {
			lugtCondition := table.Catalogue.LastUpdatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnGreaterThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(lugtCondition)
			} else {
				whereCondition = lugtCondition
				isWhereClauseSet = true
			}
		}
		if n.LastUpdatedOnLessThan != "" {
			lultCondition := table.Catalogue.LastUpdatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnLessThan), jet.String("YYYY-MM-DD")))
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
				iaCondition = table.Catalogue.CanPublicView.IS_TRUE()
			} else {
				iaCondition = table.Catalogue.CanPublicView.IS_FALSE()
			}
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(iaCondition)
			} else {
				whereCondition = iaCondition
				isWhereClauseSet = true
			}
		}
		if n.ProjectID != uuid.Nil {
			pidCondition := table.Catalogue.ProjectID.EQ(jet.UUID(n.ProjectID))
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
			case table.Catalogue.CreatedOn.Name():
				if n.SortByOrder == "asc" {
					selectQuery = selectQuery.ORDER_BY(table.Catalogue.CreatedOn.ASC())
				} else {
					selectQuery = selectQuery.ORDER_BY(table.Catalogue.CreatedOn.DESC())
				}
			case table.Catalogue.LastUpdatedOn.Name():
				if n.SortByOrder == "asc" {
					selectQuery = selectQuery.ORDER_BY(table.Catalogue.LastUpdatedOn.ASC())
				} else {
					selectQuery = selectQuery.ORDER_BY(table.Catalogue.LastUpdatedOn.DESC())
				}
			}
		}
		n.CataloguesDirectoryProject = []catalogueDirectoryProject{}
		if err = selectQuery.Query(db, &n.CataloguesDirectoryProject); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get catalogues by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get catalogues")
		}
	}

	return nil
}

func (n *catalogue) updateCatalogue() error {
	columnsToUpdate := make(jet.ColumnList, 0)
	for _, column := range n.CatalogueUpdate.Columns {
		switch column {
		case table.Catalogue.Name.Name():
			if len(n.CatalogueUpdate.Catalogue.Name) >= 3 {
				columnsToUpdate = append(columnsToUpdate, table.Catalogue.Name)
			}
		case table.Catalogue.Description.Name():
			if len(n.CatalogueUpdate.Catalogue.Description) >= 3 {
				columnsToUpdate = append(columnsToUpdate, table.Catalogue.Description)
			}
		case table.Catalogue.Catalogue.Name():
			columnsToUpdate = append(columnsToUpdate, table.Catalogue.Catalogue)
		case table.Catalogue.CanPublicView.Name():
			columnsToUpdate = append(columnsToUpdate, table.Catalogue.CanPublicView)
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

	updateQuery := table.Catalogue.
		UPDATE(columnsToUpdate).
		MODEL(n.CatalogueUpdate.Catalogue).
		WHERE(
			table.Catalogue.ID.EQ(jet.UUID(n.CatalogueID)).
				AND(table.Catalogue.ProjectID.EQ(jet.UUID(n.ProjectID)).
					AND(table.Catalogue.DirectoryID.EQ(jet.UUID(n.CurrentUser.DirectoryID))),
				)).
		RETURNING(table.Catalogue.ID, table.Catalogue.LastUpdatedOn)

	if err := updateQuery.Query(db, &n.CatalogueUpdate.Catalogue); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Update catalogue %v by %v failed | reason: %v", n.CatalogueID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not update catalogue")
	}

	return nil
}

func (n *catalogue) createCatalogue() error {
	if len(n.Catalogue.Name) < 3 || len(n.Catalogue.Description) < 3 {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	n.Catalogue.DirectoryID = n.CurrentUser.DirectoryID

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	insertQuery := table.Catalogue.
		INSERT(
			table.Catalogue.ProjectID,
			table.Catalogue.DirectoryID,
			table.Catalogue.Name,
			table.Catalogue.Description,
			table.Catalogue.Catalogue,
		).MODEL(n.Catalogue).RETURNING(table.Catalogue.ID)

	if err := insertQuery.Query(db, &n.Catalogue); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Create catalogue by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not create catalogue")
	}

	return nil
}
