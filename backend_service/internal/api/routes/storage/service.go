package storage

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

func (n *storage) getStorageTypes() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	selectQuery := table.StorageTypes.SELECT(table.StorageTypes.AllColumns)

	n.StorageTypes = []model.StorageTypes{}
	if err := selectQuery.Query(db, &n.StorageTypes); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get storage types by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not get storage types")
	}

	return nil
}

func (n *storage) getStorages() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	selectQuery := jet.SELECT(table.Storage.AllColumns, table.StorageTypes.AllColumns).
		FROM(table.Storage.INNER_JOIN(table.StorageTypes, table.Storage.StorageTypeID.EQ(table.StorageTypes.ID)))

	if n.StorageID != uuid.Nil {
		selectQuery = selectQuery.WHERE(table.Storage.ID.EQ(jet.UUID(n.StorageID)))
		n.StorageStorageType = storageStorageType{}
		if err = selectQuery.Query(db, &n.StorageStorageType); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get storage %v by %v failed | reason: %v", n.StorageID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get storage")
		}
	} else {
		n.StoragesStorageType = []storageStorageType{}
		if n.Limit > 0 {
			selectQuery = selectQuery.LIMIT(int64(n.Limit))
		}
		if n.Offset > 0 {
			selectQuery = selectQuery.OFFSET(int64(n.Offset))
		}
		if err = selectQuery.Query(db, &n.StoragesStorageType); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get storages by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get storages")
		}
	}

	return nil
}

func (n *storage) getProjectsStorage() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	if !lib.IsUserAuthorized(true, uuid.Nil, []string{}, n.CurrentUser, nil) {
		if n.ProjectID == uuid.Nil {
			return lib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		}
		if !lib.IsUserAuthorized(false, n.ProjectID, []string{}, n.CurrentUser, nil) {
			return lib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		}
	}

	selectQuery := jet.SELECT(table.StorageProjects.AllColumns, table.Storage.Name, table.Storage.StorageTypeID, table.Projects.Name, table.Projects.Description).
		FROM(
			table.StorageProjects.INNER_JOIN(table.Storage, table.StorageProjects.StorageID.EQ(table.Storage.ID)).
				INNER_JOIN(table.Projects, table.StorageProjects.ProjectID.EQ(table.Projects.ID)),
		)

	if n.ProjectID != uuid.Nil {
		selectQuery = selectQuery.WHERE(table.StorageProjects.ProjectID.EQ(jet.UUID(n.ProjectID)))
	}
	if n.Limit > 0 {
		selectQuery = selectQuery.LIMIT(int64(n.Limit))
	}
	if n.Offset > 0 {
		selectQuery = selectQuery.OFFSET(int64(n.Offset))
	}

	n.ProjectsStorage = []projectStorage{}
	if err := selectQuery.Query(db, &n.ProjectsStorage); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get project's storage by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not get project's storage")
	}

	return nil
}

func (n *storage) getFiles() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	var selectQuery jet.SelectStatement
	if n.QuickSearch == "true" {
		selectQuery = table.Files.SELECT(table.Files.ID, table.Files.Tags, table.Files.ContentType)
	} else {

		selectQuery = jet.SELECT(table.Files.AllColumns, table.Storage.Name, table.Storage.StorageTypeID, table.Directory.Name, table.Directory.Contacts, table.Projects.Name, table.Projects.Description).
			FROM(
				table.Files.INNER_JOIN(table.Storage, table.Files.StorageID.EQ(table.Storage.ID)).
					INNER_JOIN(table.Directory, table.Files.DirectoryID.EQ(table.Directory.ID)).
					INNER_JOIN(table.Projects, table.Files.ProjectID.EQ(table.Projects.ID)),
			)
	}

	if n.FileID != uuid.Nil {
		selectQuery = selectQuery.WHERE(table.Files.ID.EQ(jet.UUID(n.FileID)))
		n.FileStorageDirectoryProject = fileStorageDirectoryProject{}
		if err = selectQuery.Query(db, &n.FileStorageDirectoryProject); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get file %v by %v failed | reason: %v", n.FileID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get file")
		}
	} else {
		var whereCondition jet.BoolExpression
		isWhereClauseSet := false
		if n.SearchQuery != "" {
			whereCondition = jet.BoolExp(lib.GetTextSearchBoolExp(table.Files.FileVector.Name(), n.SearchQuery))
			isWhereClauseSet = true
		}
		if n.CreatedOnGreaterThan != "" {
			cogtCondition := table.Files.CreatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnGreaterThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(cogtCondition)
			} else {
				whereCondition = cogtCondition
				isWhereClauseSet = true
			}
		}
		if n.CreatedOnLessThan != "" {
			coltCondition := table.Files.CreatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnLessThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(coltCondition)
			} else {
				whereCondition = coltCondition
				isWhereClauseSet = true
			}
		}
		if n.ProjectID != uuid.Nil {
			pidCondition := table.Files.ProjectID.EQ(jet.UUID(n.ProjectID))
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(pidCondition)
			} else {
				whereCondition = pidCondition
				isWhereClauseSet = true
			}
		}
		if n.FilesWithAbstractions != "" {
			var fwaCondition jet.BoolExpression
			if n.FilesWithAbstractions == "true" {
				fwaCondition = table.Files.ID.IN(table.Abstractions.SELECT(table.Abstractions.FileID))
			} else {
				fwaCondition = table.Files.ID.NOT_IN(table.Abstractions.SELECT(table.Abstractions.FileID))
			}
			if isWhereClauseSet {
				whereCondition = whereCondition.AND(fwaCondition)
			} else {
				whereCondition = fwaCondition
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
		if n.SortyBy == table.Files.CreatedOn.Name() {
			if n.SortByOrder == "asc" {
				selectQuery = selectQuery.ORDER_BY(table.Files.CreatedOn.ASC())
			} else {
				selectQuery = selectQuery.ORDER_BY(table.Files.CreatedOn.DESC())
			}
		}
		n.FilesStorageDirectoryProject = []fileStorageDirectoryProject{}
		if err = selectQuery.Query(db, &n.FilesStorageDirectoryProject); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get files by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get files")
		}
	}

	return nil
}

func (n *storage) deleteStorageProject() (int64, error) {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var whereClause jet.BoolExpression
	whereClauseSet := false
	if n.StorageProject.ProjectID != uuid.Nil {
		whereClause = table.StorageProjects.ProjectID.EQ(jet.UUID(n.StorageProject.ProjectID))
		whereClauseSet = true
	}
	if n.StorageProject.StorageID != uuid.Nil {
		condition := table.StorageProjects.StorageID.EQ(jet.UUID(n.StorageProject.StorageID))
		if whereClauseSet {
			whereClause = whereClause.AND(condition)
		} else {
			whereClause = condition
		}
	}

	deleteQuery := table.StorageProjects.DELETE().WHERE(whereClause)
	if sqlResults, err := deleteQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete storage project, project %v and/or stroage %v by %v, failed | reason: %v", n.StorageProject.ProjectID, n.StorageProject.StorageID, n.CurrentUser.DirectoryID, err))
		return -1, lib.NewError(http.StatusInternalServerError, "Could not delete storage project")
	} else {
		if deletedRows, err := sqlResults.RowsAffected(); err != nil {
			intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Determining no. of storage projects deleted failed while deleting project %v and/or storage %v | reason: %v", n.StorageProject.ProjectID, n.StorageProject.StorageID, err))
			return -1, nil
		} else {
			return deletedRows, err
		}
	}
}

func (n *storage) deleteFileInfo() (int64, error) {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	deleteQuery := table.Files.DELETE().WHERE(table.Files.ID.EQ(jet.UUID(n.File.ID)))

	if sqlResults, err := deleteQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete file %v by %v failed | reason: %v", n.File.ID, n.CurrentUser.DirectoryID, err))
		return -1, lib.NewError(http.StatusInternalServerError, "Could not delete file info")
	} else {
		if deletedRows, err := sqlResults.RowsAffected(); err != nil {
			intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Determining no. of files deleted failed while deleting %v | reason: %v", n.File.ID, err))
			return -1, nil
		} else {
			return deletedRows, err
		}
	}
}

func (n *storage) getFile() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	selectQuery := jet.SELECT(
		table.Files.DirectoryID,
		table.Files.ContentType,
		table.Storage.StorageTypeID,
		table.Storage.Storage,
	).
		FROM(table.Files.INNER_JOIN(table.Storage, table.Files.StorageID.EQ(table.Storage.ID))).
		WHERE(table.Files.ID.EQ(jet.UUID(n.File.ID)).AND(table.Files.ProjectID.EQ(jet.UUID(n.File.ProjectID))))

	n.FileStorage = fileStorage{}
	if err := selectQuery.Query(db, &n.FileStorage); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get info for file %v by %v failed | reason: %v", n.File.ID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not get file info")
	}

	return nil
}

func (n *storage) createFile() error {
	if len(n.File.Tags) < 1 {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	selectQuery := jet.SELECT(
		table.StorageProjects.CreatedOn,
		table.Storage.StorageTypeID,
		table.Storage.Storage,
	).FROM(
		table.StorageProjects.INNER_JOIN(table.Storage, table.StorageProjects.StorageID.EQ(table.Storage.ID).AND(table.Storage.IsActive.IS_TRUE())),
	).WHERE(
		table.StorageProjects.StorageID.EQ(jet.UUID(n.StorageProject.StorageID)).AND(table.StorageProjects.ProjectID.EQ(jet.UUID(n.StorageProject.ProjectID))))

	n.ProjectStorage = storageProject{}

	if err := selectQuery.Query(db, &n.ProjectStorage); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get storage %v for project %v by %v failed | reason: %v", n.StorageProject.StorageID, n.StorageProject.ProjectID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not get storage for project")
	}

	if n.ProjectStorage.CreatedOn.IsZero() {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	insertQuery := table.Files.
		INSERT(
			table.Files.StorageID,
			table.Files.ProjectID,
			table.Files.DirectoryID,
			table.Files.Tags,
			table.Files.ContentType,
		).MODEL(n.File).RETURNING(table.Files.ID, table.Files.CreatedOn)

	if err := insertQuery.Query(db, &n.File); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Create file entry in storage %v for project %v by %v failed | reason: %v", n.StorageProject.StorageID, n.StorageProject.ProjectID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not create file")
	}

	return nil
}

func (n *storage) createStorageProject() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	insertQuery := table.StorageProjects.
		INSERT(table.StorageProjects.ProjectID, table.StorageProjects.StorageID).
		MODEL(n.StorageProject).
		RETURNING(table.StorageProjects.ProjectID, table.StorageProjects.StorageID, table.StorageProjects.CreatedOn)

	if err := insertQuery.Query(db, &n.StorageProject); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Add storage %v to project %v by %v failed | reason: %v", n.StorageProject.StorageID, n.StorageProject.ProjectID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not add storage to project")
	}

	n.Storage = model.Storage{}
	selectQuery := table.Storage.SELECT(table.Storage.AllColumns).WHERE(table.Storage.ID.EQ(jet.UUID(n.StorageProject.StorageID)))
	if err := selectQuery.Query(db, &n.Storage); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get storage %v by %v for folder creation failed | reason: %v", n.StorageProject.StorageID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not create project folder")
	}

	if err := n.createFolder(n.StorageProject.ProjectID.String()); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Create project folder %v in storage %v by %v failed | reason: %v", n.StorageProject.ProjectID, n.StorageProject.StorageID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not create project folder")
	}

	return nil
}

func (n *storage) deleteStorage() (int64, error) {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return -1, lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	deleteQuery := table.Storage.DELETE().WHERE(table.Storage.ID.EQ(jet.UUID(n.StorageID)))

	if sqlResults, err := deleteQuery.Exec(db); err != nil {
		updateQuery := table.Storage.
			UPDATE(table.Storage.IsActive).
			MODEL(model.Storage{IsActive: false}).
			WHERE(table.Storage.ID.EQ(jet.UUID(n.StorageID)))
		if sqlResults, err := updateQuery.Exec(db); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Deactivate storage %v by %v failed | reason: %v", n.StorageID, n.CurrentUser.DirectoryID, err))
			return -1, lib.NewError(http.StatusInternalServerError, "Could not deactivate storage")
		} else {
			if deletedRows, err := sqlResults.RowsAffected(); err != nil {
				intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Determining no. of storage deleted failed while deleting %v | reason: %v", n.StorageID, err))
				return -1, nil
			} else {
				return deletedRows, err
			}
		}
	} else {
		if deletedRows, err := sqlResults.RowsAffected(); err != nil {
			intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Determining no. of storage deleted failed while deleting %v | reason: %v", n.StorageID, err))
			return -1, nil
		} else {
			return deletedRows, err
		}
	}
}

func (n *storage) updateStorage() error {
	columnsToUpdate := make(jet.ColumnList, 0)
	for _, column := range n.StorageUpdate.Columns {
		switch column {
		case table.Storage.StorageTypeID.Name():
			columnsToUpdate = append(columnsToUpdate, table.Storage.StorageTypeID)
		case table.Storage.Name.Name():
			if len(n.StorageUpdate.Storage.Name) >= 3 {
				columnsToUpdate = append(columnsToUpdate, table.Storage.Name)
			}
		case table.Storage.Storage.Name():
			columnsToUpdate = append(columnsToUpdate, table.Storage.Storage)
		case table.Storage.IsActive.Name():
			columnsToUpdate = append(columnsToUpdate, table.Storage.IsActive)
		}
	}

	if len(columnsToUpdate) < 1 {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	updateQuery := table.Storage.
		UPDATE(columnsToUpdate).
		MODEL(n.StorageUpdate.Storage).
		WHERE(table.Storage.ID.EQ(jet.UUID(n.StorageID))).
		RETURNING(table.Storage.ID, table.Storage.LastUpdatedOn)

	if err := updateQuery.Query(db, &n.StorageUpdate.Storage); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Update storage %v failed | reason: %v", n.StorageID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not update storage")
	}

	return nil
}

func (n *storage) createStorage() error {
	if len(n.Storage.Name) < 3 {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	insertQuery := table.Storage.
		INSERT(table.Storage.StorageTypeID, table.Storage.Name, table.Storage.Storage).
		MODEL(n.Storage).
		RETURNING(table.Storage.ID)

	if err := insertQuery.Query(db, &n.Storage); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Create storage by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not create storage")
	}

	return nil
}
