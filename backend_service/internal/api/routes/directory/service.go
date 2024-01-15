package directory

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

func (n *directory) deleteUser() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	deleteQuery := table.DirectoryIam.DELETE().WHERE(table.DirectoryIam.DirectoryID.EQ(jet.UUID(n.DirectoryID)))

	if _, err := deleteQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete user %v credentials by %v failed | reason: %v", n.DirectoryID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not delete user")
	}

	deleteQuery = table.DirectorySystemUsers.DELETE().WHERE(table.DirectorySystemUsers.DirectoryID.EQ(jet.UUID(n.DirectoryID)))

	if _, err := deleteQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete system user %v by %v failed | reason: %v", n.DirectoryID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not delete user")
	}

	deleteQuery = table.Directory.DELETE().WHERE(table.Directory.ID.EQ(jet.UUID(n.DirectoryID)))
	if _, err := deleteQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Delete user %v by %v failed | reason: %v", n.DirectoryID, n.CurrentUser.DirectoryID, err))
	}

	return nil
}

func (n *directory) updateUser() error {
	columnsToUpdate := make(jet.ColumnList, 0)
	for _, column := range n.DirectoryUpdate.Columns {
		switch column {
		case table.Directory.Name.Name():
			if len(n.DirectoryUpdate.Directory.Name) >= 3 {
				columnsToUpdate = append(columnsToUpdate, table.Directory.Name)
			}
		case table.Directory.Contacts.Name():
			if len(n.DirectoryUpdate.Directory.Contacts) > 0 {
				columnsToUpdate = append(columnsToUpdate, table.Directory.Contacts)
			}
		}
	}

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	isRequestProcessed := false
	if len(columnsToUpdate) > 0 && len(columnsToUpdate) < 3 {
		whereCondition := table.Directory.ID.EQ(jet.UUID(n.DirectoryID))
		if n.ProjectID != uuid.Nil {
			whereCondition.AND(
				table.DirectoryProjectsRoles.DirectoryID.EQ(jet.UUID(n.DirectoryID)).
					AND(table.DirectoryProjectsRoles.ProjectID.EQ(jet.UUID(n.ProjectID))),
			)
		}

		updateQuery := table.Directory.
			UPDATE(columnsToUpdate).
			MODEL(n.DirectoryUpdate.Directory).
			WHERE(whereCondition).
			RETURNING(table.Directory.ID, table.Directory.LastUpdatedOn)

		if err = updateQuery.Query(db, &n.DirectoryUpdate.Directory); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Update user %v by %v failed | reason: %v", n.DirectoryID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not update user")
		}
		isRequestProcessed = true
	}

	if (n.DirectoryUpdate.IsSystemUser == "true" || n.DirectoryUpdate.IsSystemUser == "false") && lib.IsUserAuthorized(true, uuid.Nil, []string{}, n.CurrentUser, nil) {
		if n.DirectoryUpdate.IsSystemUser == "true" {
			insertQuery := table.DirectorySystemUsers.
				INSERT(table.DirectorySystemUsers.DirectoryID).
				MODEL(model.DirectorySystemUsers{DirectoryID: n.DirectoryID})
			if _, err := insertQuery.Exec(db); err != nil {
				intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Add system user %v by %v failed | reason: %v", n.DirectoryID, n.CurrentUser.DirectoryID, err))
				return lib.NewError(http.StatusInternalServerError, "Could not update system user")
			}
		} else {
			deleteQuery := table.DirectorySystemUsers.DELETE().WHERE(table.DirectorySystemUsers.DirectoryID.EQ(jet.UUID(n.DirectoryID)))
			if _, err := deleteQuery.Exec(db); err != nil {
				intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete system user %v by %v failed | reason: %v", n.DirectoryID, n.CurrentUser.DirectoryID, err))
				return lib.NewError(http.StatusInternalServerError, "Could not update system user")
			}
		}
		isRequestProcessed = true
	}

	if !isRequestProcessed {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	return nil
}

func (n *directory) getUsers() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	if n.DirectoryID != uuid.Nil {
		var fromTables jet.ReadableTable = table.Directory
		columnsToSelect := make(jet.ProjectionList, 0)
		columnsToSelect = append(columnsToSelect, jet.ProjectionList{
			table.Directory.ID.AS("retrieve_user.id"),
			table.Directory.Name.AS("retrieve_user.name"),
			table.Directory.Contacts.AS("retrieve_user.contacts"),
			table.Directory.CreatedOn.AS("retrieve_user.created_on"),
			table.Directory.LastUpdatedOn.AS("retrieve_user.last_updated_on"),
		})
		if lib.IsUserAuthorized(true, uuid.Nil, []string{}, n.CurrentUser, nil) {
			fromTables = fromTables.LEFT_JOIN(table.DirectorySystemUsers, table.Directory.ID.EQ(table.DirectorySystemUsers.DirectoryID))
			columnsToSelect = append(columnsToSelect, table.DirectorySystemUsers.CreatedOn.AS("retrieve_user.system_user_created_on"))
		}
		if lib.IsUserAuthorized(true, uuid.Nil, []string{}, n.CurrentUser, nil) || (n.ProjectID != uuid.Nil && lib.IsUserAuthorized(false, n.ProjectID, []string{lib.ROLE_PROJECT_ADMIN}, n.CurrentUser, nil)) {
			fromTables = fromTables.LEFT_JOIN(table.DirectoryIam, table.Directory.ID.EQ(table.DirectoryIam.DirectoryID))
			columnsToSelect = append(columnsToSelect, jet.ProjectionList{
				table.DirectoryIam.Email.AS("retrieve_user.iam_email"),
				table.DirectoryIam.CreatedOn.AS("retrieve_user.iam_created_on"),
				table.DirectoryIam.LastUpdatedOn.AS("retrieve_user.iam_last_updated_on"),
			})
		}
		if (n.ProjectID != uuid.Nil && lib.IsUserAuthorized(false, n.ProjectID, []string{}, n.CurrentUser, nil)) || lib.IsUserAuthorized(true, uuid.Nil, []string{}, n.CurrentUser, nil) {
			fromTables = fromTables.LEFT_JOIN(
				table.DirectoryProjectsRoles.INNER_JOIN(
					table.Projects,
					table.DirectoryProjectsRoles.ProjectID.EQ(table.Projects.ID),
				),
				table.DirectoryProjectsRoles.DirectoryID.EQ(table.DirectoryIam.DirectoryID),
			)
			columnsToSelect = append(columnsToSelect, jet.ProjectionList{
				table.Projects.ID.AS("project.project_id"),
				table.Projects.Name.AS("project.project_name"),
				table.Projects.Description.AS("project.project_description"),
				table.Projects.CreatedOn.AS("project.project_created_on"),
				table.DirectoryProjectsRoles.ProjectRoleID.AS("project_roles.project_role_id"),
				table.DirectoryProjectsRoles.CreatedOn.AS("project_roles.project_role_created_on"),
			})
		}
		selectQuery := jet.SELECT(columnsToSelect).FROM(fromTables).WHERE(table.Directory.ID.EQ(jet.UUID(n.DirectoryID)))
		n.RetrieveUser = RetrieveUser{}
		if err = selectQuery.Query(db, &n.RetrieveUser); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get user %v by %v failed | reason: %v", n.DirectoryID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get user")
		}
	} else {
		usersSelectQuery := jet.SELECT(
			table.Directory.ID.AS("retrieve_user.id"),
			table.Directory.Name.AS("retrieve_user.name"),
			table.Directory.Contacts.AS("retrieve_user.contacts"),
			table.Directory.CreatedOn.AS("retrieve_user.created_on"),
			table.Directory.LastUpdatedOn.AS("retrieve_user.last_updated_on"),
		).FROM(table.Directory)
		var usersSelectwhereCondition jet.BoolExpression
		isWhereClauseSet := false
		if n.SearchQuery != "" {
			usersSelectwhereCondition = jet.BoolExp(lib.GetTextSearchBoolExp(table.Directory.DirectoryVector.Name(), n.SearchQuery))
			isWhereClauseSet = true
		}
		if n.CreatedOnGreaterThan != "" {
			cogtCondition := table.Directory.CreatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnGreaterThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				usersSelectwhereCondition = usersSelectwhereCondition.AND(cogtCondition)
			} else {
				usersSelectwhereCondition = cogtCondition
				isWhereClauseSet = true
			}
		}
		if n.CreatedOnLessThan != "" {
			coltCondition := table.Directory.CreatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnLessThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				usersSelectwhereCondition = usersSelectwhereCondition.AND(coltCondition)
			} else {
				usersSelectwhereCondition = coltCondition
				isWhereClauseSet = true
			}
		}
		if n.LastUpdatedOnGreaterThan != "" {
			lugtCondition := table.Directory.LastUpdatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnGreaterThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				usersSelectwhereCondition = usersSelectwhereCondition.AND(lugtCondition)
			} else {
				usersSelectwhereCondition = lugtCondition
				isWhereClauseSet = true
			}
		}
		if n.LastUpdatedOnLessThan != "" {
			lultCondition := table.Directory.LastUpdatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnLessThan), jet.String("YYYY-MM-DD")))
			if isWhereClauseSet {
				usersSelectwhereCondition = usersSelectwhereCondition.AND(lultCondition)
			} else {
				usersSelectwhereCondition = lultCondition
				isWhereClauseSet = true
			}
		}
		if isWhereClauseSet {
			usersSelectQuery = usersSelectQuery.WHERE(usersSelectwhereCondition)
		}
		if n.Limit > 0 {
			usersSelectQuery = usersSelectQuery.LIMIT(int64(n.Limit))
		}
		if n.Offset > 0 {
			usersSelectQuery = usersSelectQuery.OFFSET(int64(n.Offset))
		}
		if n.SortyBy != "" {
			switch n.SortyBy {
			case table.Directory.CreatedOn.Name():
				if n.SortByOrder == "asc" {
					usersSelectQuery = usersSelectQuery.ORDER_BY(table.Directory.CreatedOn.ASC())
				} else {
					usersSelectQuery = usersSelectQuery.ORDER_BY(table.Directory.CreatedOn.DESC())
				}
			case table.Directory.LastUpdatedOn.Name():
				if n.SortByOrder == "asc" {
					usersSelectQuery = usersSelectQuery.ORDER_BY(table.Directory.LastUpdatedOn.ASC())
				} else {
					usersSelectQuery = usersSelectQuery.ORDER_BY(table.Directory.LastUpdatedOn.DESC())
				}
			}
		}
		selectedUsers := usersSelectQuery.AsTable(table.Directory.TableName())
		var fromTables jet.ReadableTable = selectedUsers
		columnsToSelect := make(jet.ProjectionList, 0)
		columnsToSelect = append(columnsToSelect, selectedUsers.AllColumns())
		if lib.IsUserAuthorized(true, uuid.Nil, []string{}, n.CurrentUser, nil) {
			fromTables = fromTables.LEFT_JOIN(table.DirectorySystemUsers, table.DirectorySystemUsers.DirectoryID.EQ(jet.StringColumn("retrieve_user.id").From(selectedUsers)))
			columnsToSelect = append(columnsToSelect, table.DirectorySystemUsers.CreatedOn.AS("retrieve_user.system_user_created_on"))
		}
		if lib.IsUserAuthorized(true, uuid.Nil, []string{}, n.CurrentUser, nil) || (n.ProjectID != uuid.Nil && lib.IsUserAuthorized(false, n.ProjectID, []string{lib.ROLE_PROJECT_ADMIN}, n.CurrentUser, nil)) {
			fromTables = fromTables.LEFT_JOIN(table.DirectoryIam, table.DirectoryIam.DirectoryID.EQ(jet.StringColumn("retrieve_user.id").From(selectedUsers)))
			columnsToSelect = append(columnsToSelect, jet.ProjectionList{
				table.DirectoryIam.Email.AS("retrieve_user.iam_email"),
				table.DirectoryIam.CreatedOn.AS("retrieve_user.iam_created_on"),
				table.DirectoryIam.LastUpdatedOn.AS("retrieve_user.iam_last_updated_on"),
			})
		}
		if (n.ProjectID != uuid.Nil && lib.IsUserAuthorized(false, n.ProjectID, []string{}, n.CurrentUser, nil)) || lib.IsUserAuthorized(true, uuid.Nil, []string{}, n.CurrentUser, nil) {
			var projectRolesWhereCondition jet.BoolExpression
			if n.ProjectID != uuid.Nil {
				projectRolesWhereCondition = table.DirectoryProjectsRoles.ProjectID.EQ(jet.UUID(n.ProjectID))
				if n.ProjectRole != "" {
					projectRolesWhereCondition = projectRolesWhereCondition.AND(table.DirectoryProjectsRoles.ProjectRoleID.EQ(jet.String(n.ProjectRole)))
				}
			} else {
				projectRolesWhereCondition = table.DirectoryProjectsRoles.ProjectID.EQ(table.Projects.ID)
			}

			fromTables = fromTables.LEFT_JOIN(
				table.DirectoryProjectsRoles.INNER_JOIN(
					table.Projects,
					projectRolesWhereCondition,
				),
				table.DirectoryProjectsRoles.DirectoryID.EQ(jet.StringColumn("retrieve_user.id").From(selectedUsers)),
			)
			columnsToSelect = append(columnsToSelect, jet.ProjectionList{
				table.Projects.ID.AS("project.project_id"),
				table.Projects.Name.AS("project.project_name"),
				table.Projects.Description.AS("project.project_description"),
				table.Projects.CreatedOn.AS("project.project_created_on"),
				table.DirectoryProjectsRoles.ProjectRoleID.AS("project_roles.project_role_id"),
				table.DirectoryProjectsRoles.CreatedOn.AS("project_roles.project_role_created_on"),
			})
		}
		selectQuery := jet.SELECT(columnsToSelect).FROM(fromTables)
		n.RetrieveUsers = []RetrieveUser{}
		if err = selectQuery.Query(db, &n.RetrieveUsers); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get users by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get users")
		}
	}

	return nil
}

func (n *directory) createUser() error {
	if len(n.DirectoryCreate.Name) < 3 || len(n.DirectoryCreate.Contacts) < 1 {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	var newDirectoryIam model.DirectoryIam
	if len(n.DirectoryCreate.Email) > 7 {
		newDirectoryIam.Email = &n.DirectoryCreate.Email
	}

	insertQuery := table.Directory.
		INSERT(table.Directory.Name, table.Directory.Contacts).
		MODEL(n.DirectoryCreate).
		RETURNING(table.Directory.ID)

	n.Directory = model.Directory{}
	if err = insertQuery.Query(db, &n.Directory); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Creating user by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not create new user")
	}

	if len(n.DirectoryCreate.Email) > 7 {
		newDirectoryIam.DirectoryID = n.Directory.ID
		insertQuery = table.DirectoryIam.
			INSERT(table.DirectoryIam.DirectoryID, table.DirectoryIam.Email).
			MODEL(newDirectoryIam)

		if _, err = insertQuery.Exec(db); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Creating user iam for %v by %v failed | reason: %v", newDirectoryIam.DirectoryID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not create new user iam")
		}
	}

	if n.DirectoryCreate.ProjectID != uuid.Nil {
		newDirectoryProjectRoles := model.DirectoryProjectsRoles{
			DirectoryID:   n.Directory.ID,
			ProjectID:     n.DirectoryCreate.ProjectID,
			ProjectRoleID: lib.ROLE_EXPLORER,
		}
		insertQuery = table.DirectoryProjectsRoles.
			INSERT(table.DirectoryProjectsRoles.ProjectID, table.DirectoryProjectsRoles.DirectoryID, table.DirectoryProjectsRoles.ProjectRoleID).
			MODEL(newDirectoryProjectRoles)
		if _, err = insertQuery.Exec(db); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Add explorer role for %v in project %v by %v failed | reason: %v", newDirectoryProjectRoles.DirectoryID, newDirectoryProjectRoles.ProjectID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not add default explorer role")
		}
	}

	if n.DirectoryCreate.IsSystemUser {
		newSystemUser := model.DirectorySystemUsers{
			DirectoryID: n.Directory.ID,
		}

		insertQuery = table.DirectorySystemUsers.
			INSERT(table.DirectorySystemUsers.DirectoryID).
			MODEL(newSystemUser)
		if _, err := insertQuery.Exec(db); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Add system user %v by %v failed | reason: %v", newSystemUser.DirectoryID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not add system user")
		}
	}

	return nil
}
