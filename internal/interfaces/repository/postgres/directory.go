package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/brunoga/deep"
	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
	intlibmmodel "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib/metadata_model"
	"github.com/jackc/pgx/v5"
)

func (n *PostrgresRepository) RepoDirectoryFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID, columns []string) (*intdoment.Directory, error) {
	directoryGroupsMModel, err := intlib.MetadataModelGet(intdoment.DirectoryRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneByIamCredentialID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(directoryGroupsMModel, intdoment.DirectoryRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneByIamCredentialID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, fmt.Sprintf("%s.%s", intdoment.DirectoryRepository().RepositoryName, intdoment.DirectoryRepository().ID)) {
		columns = append(columns, intdoment.DirectoryRepository().ID)
	}

	selectColumns := make([]string, len(columns))
	for cIndex, cValue := range columns {
		selectColumns[cIndex] = intdoment.DirectoryRepository().RepositoryName + "." + cValue
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s INNER JOIN %[3]s ON %[3]s.%[4]s = $1 AND %[3]s.%[4]s = %[2]s.%[5]s;",
		strings.Join(selectColumns, ","),                    //1
		intdoment.DirectoryRepository().RepositoryName,      //2
		intdoment.IamCredentialsRepository().RepositoryName, //3
		intdoment.IamCredentialsRepository().DirectoryID,    //4
		intdoment.DirectoryRepository().ID,                  //5
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryFindOneByIamCredentialID))

	rows, err := n.db.Query(ctx, query, iamCredentialID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneByIamCredentialID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneByIamCredentialID, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(directoryGroupsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneByIamCredentialID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneByIamCredentialID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoDirectoryFindOneByIamCredentialID))
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneByIamCredentialID, fmt.Errorf("more than one %s found", intdoment.DirectoryRepository().RepositoryName))
	}

	directory := new(intdoment.Directory)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneByIamCredentialID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing directory", "directory", string(jsonData), "function", intlib.FunctionName(n.RepoDirectoryFindOneByIamCredentialID))
		if err := json.Unmarshal(jsonData, directory); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneByIamCredentialID, err)
		}
	}

	return directory, nil
}

func (n *PostrgresRepository) RepoDirectoryUpdateOne(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	fieldAnyMetadataModelGet intdomint.FieldAnyMetadataModel,
	authContextDirectoryGroupID uuid.UUID,
	datum *intdoment.Directory,
) error {
	where := ""
	datum.DirectoryGroupsID = make([]uuid.UUID, 1)
	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1;",
		intdoment.DirectoryRepository().DirectoryGroupsID, //1
		intdoment.DirectoryRepository().RepositoryName,    //2
		intdoment.DirectoryRepository().ID,                //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryUpdateOne))
	if err := n.db.QueryRow(ctx, query, datum.ID[0]).Scan(&datum.DirectoryGroupsID[0]); err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryUpdateOne, fmt.Errorf("get %s for %s failed, err: %v", intdoment.DirectoryRepository().DirectoryGroupsID, intdoment.DirectoryRepository().RepositoryName, err))
	}

	query = ""
	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		cteName := fmt.Sprintf("%s_%s", intdoment.DirectoryRepository().RepositoryName, RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME)
		query = RecursiveDirectoryGroupsSubGroupsCte(authContextDirectoryGroupID, cteName)
		where = fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.DirectoryRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName)
	}
	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		whereQuery := fmt.Sprintf("%s = '%s'", intdoment.DirectoryRepository().DirectoryGroupsID, authContextDirectoryGroupID.String())
		if len(where) > 0 {
			where += " OR " + whereQuery
		} else {
			where = whereQuery
		}
	}
	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		whereQuery := ""
		if len(iamCredential.DirectoryID) > 0 {
			whereQuery = fmt.Sprintf("%s = '%s'", intdoment.DirectoryRepository().ID, iamCredential.DirectoryID[0].String())
		} else {
			return intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		}

		if len(where) > 0 {
			where += " OR " + whereQuery
		} else {
			where = whereQuery
		}
	}
	if len(where) == 0 {
		return intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}

	valuesToUpdate := make([]any, 0)
	valueToUpdateQuery := make([]string, 0)
	columnsToUpdate := make([]string, 0)
	nextPlaceholder := 1
	if v, vQ, c := n.RepoDirectoryValidateAndGetColumnsAndData(ctx, datum.DirectoryGroupsID[0], &nextPlaceholder, fieldAnyMetadataModelGet, datum, false); len(c) == 0 || len(v) == 0 || len(vQ) == 0 {
		return intlib.NewError(http.StatusBadRequest, "no values to update")
	} else {
		valuesToUpdate = append(valuesToUpdate, v...)
		columnsToUpdate = append(columnsToUpdate, c...)
		valueToUpdateQuery = append(valueToUpdateQuery, vQ...)
	}

	query += fmt.Sprintf(
		" UPDATE %[1]s SET %[2]s WHERE %[3]s = %[4]s AND %[5]s IS NULL AND (%[6]s);",
		intdoment.DirectoryRepository().RepositoryName,                     //1
		GetUpdateSetColumnsWithVQuery(columnsToUpdate, valueToUpdateQuery), //2
		intdoment.DirectoryRepository().ID,                                 //3
		GetandUpdateNextPlaceholder(&nextPlaceholder),                      //4
		intdoment.DirectoryRepository().DeactivatedOn,                      //5
		where, //6
	)
	query = strings.TrimLeft(query, " \n")
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryUpdateOne))

	valuesToUpdate = append(valuesToUpdate, datum.ID[0])
	if _, err := n.db.Exec(ctx, query, valuesToUpdate...); err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryUpdateOne, fmt.Errorf("update %s failed, err: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoDirectoryInsertOne(
	ctx context.Context,
	datum *intdoment.Directory,
	authContextDirectoryGroupID uuid.UUID,
	iamAuthorizationRule *intdoment.IamAuthorizationRule,
	fieldAnyMetadataModelGet intdomint.FieldAnyMetadataModel,
	columns []string,
) (*intdoment.Directory, error) {
	directoryMetadataModel, err := intlib.MetadataModelGet(intdoment.DirectoryRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(directoryMetadataModel, intdoment.DirectoryRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.DirectoryRepository().ID) {
		columns = append(columns, intdoment.DirectoryRepository().ID)
	}

	valuesToInsert := make([]any, 0)
	valueToInsertQuery := make([]string, 0)
	columnsToInsert := make([]string, 0)
	nextPlaceholder := 1
	if v, vQ, c := n.RepoDirectoryValidateAndGetColumnsAndData(ctx, authContextDirectoryGroupID, &nextPlaceholder, fieldAnyMetadataModelGet, datum, true); len(c) == 0 || len(v) == 0 || len(vQ) == 0 {
		return nil, intlib.NewError(http.StatusBadRequest, "no values to insert")
	} else {
		valuesToInsert = append(valuesToInsert, v...)
		columnsToInsert = append(columnsToInsert, c...)
		valueToInsertQuery = append(valueToInsertQuery, vQ...)
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s) VALUES(%[3]s) RETURNING %[4]s;",
		intdoment.DirectoryRepository().RepositoryName, //1
		strings.Join(columnsToInsert, " , "),           //2
		strings.Join(valueToInsertQuery, " , "),        //3
		strings.Join(columns, " , "),                   //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryInsertOne))

	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, fmt.Errorf("start transaction to create %s failed, error: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}

	rows, err := transaction.Query(ctx, query, valuesToInsert...)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(directoryMetadataModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryInsertOne, err)
	}

	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		transaction.Rollback(ctx)
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.DirectoryRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		transaction.Rollback(ctx)
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoDirectoryInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryInsertOne, fmt.Errorf("more than one %s found", intdoment.DirectoryRepository().RepositoryName))
	}

	directory := new(intdoment.Directory)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing directory", "directory", string(jsonData), "function", intlib.FunctionName(n.RepoDirectoryInsertOne))
		if err := json.Unmarshal(jsonData, directory); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryInsertOne, err)
		}
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.DirectoryAuthorizationIDsRepository().RepositoryName,                   //1
		intdoment.DirectoryAuthorizationIDsRepository().ID,                               //2
		intdoment.DirectoryAuthorizationIDsRepository().CreationIamGroupAuthorizationsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryInsertOne))

	if _, err := transaction.Exec(ctx, query, directory.ID[0], iamAuthorizationRule.ID); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.DirectoryAuthorizationIDsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, fmt.Errorf("commit transaction to create %s failed, error: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}

	return directory, nil
}

func (n *PostrgresRepository) RepoDirectoryValidateAndGetColumnsAndData(ctx context.Context, directoryGroupID uuid.UUID, nextPlaceholder *int, fieldAnyMetadataModelGet intdomint.FieldAnyMetadataModel, datum *intdoment.Directory, insert bool) ([]any, []string, []string) {
	values := make([]any, 0)
	valuesQuery := make([]string, 0)
	columns := make([]string, 0)

	if insert {
		columns = append(columns, intdoment.DirectoryRepository().DirectoryGroupsID)
		values = append(values, directoryGroupID)
		valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
	}

	fullTextSearchValue := make([]string, 0)
	if len(datum.Data) > 0 {
		if dMap, ok := datum.Data[0].(map[string]any); ok {
			columns = append(columns, intdoment.DirectoryRepository().Data)
			values = append(values, dMap)
			valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
		} else {
			if insert {
				datum.Data[0] = map[string]any{}
				columns = append(columns, intdoment.DirectoryRepository().Data)
				values = append(values, datum.Data[0])
				valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
			}
		}
	} else {
		if insert {
			datum.Data = []any{map[string]any{}}
			columns = append(columns, intdoment.DirectoryRepository().Data)
			values = append(values, datum.Data[0])
			valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
		}
	}

	if len(datum.DisplayName) > 0 {
		columns = append(columns, intdoment.DirectoryRepository().DisplayName)
		values = append(values, datum.DisplayName[0])
		valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
		fullTextSearchValue = append(fullTextSearchValue, datum.DisplayName...)
	}

	if slices.Contains(columns, intdoment.DirectoryRepository().Data) {
		if mm, err := fieldAnyMetadataModelGet.GetMetadataModel(ctx, intdoment.MetadataModelsDirectoryRepository().RepositoryName, "$", intdoment.MetadataModelsDirectoryRepository().RepositoryName, []any{directoryGroupID}); err == nil {
			if value := MetadataModelExtractFullTextSearchValue(mm, datum.Data[0]); len(value) > 0 {
				fullTextSearchValue = append(fullTextSearchValue, value...)
			}
		}
	}

	if len(fullTextSearchValue) > 0 {
		columns = append(columns, intdoment.DirectoryRepository().FullTextSearch)
		values = append(values, strings.Join(fullTextSearchValue, " "))
		valuesQuery = append(valuesQuery, fmt.Sprintf("to_tsvector(%s)", GetandUpdateNextPlaceholder(nextPlaceholder)))
	}

	return values, valuesQuery, columns
}

func (n *PostrgresRepository) RepoDirectoryDeleteOne(
	ctx context.Context,
	iamAuthRule *intdoment.IamAuthorizationRule,
	datum *intdoment.Directory,
) error {
	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryDeleteOne, fmt.Errorf("start transaction to delete %s failed, error: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s = $1;",
		intdoment.DirectoryAuthorizationIDsRepository().RepositoryName, //1
		intdoment.DirectoryAuthorizationIDsRepository().ID,             //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryDeleteOne))

	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		query := fmt.Sprintf(
			"DELETE FROM %[1]s WHERE %[2]s = $1;",
			intdoment.DirectoryRepository().RepositoryName, //1
			intdoment.DirectoryRepository().ID,             //2
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryDeleteOne))
		if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
			if err := transaction.Commit(ctx); err != nil {
				return intlib.FunctionNameAndError(n.RepoDirectoryDeleteOne, fmt.Errorf("commit transaction to delete %s failed, error: %v", intdoment.DirectoryRepository().RepositoryName, err))
			}
			return nil
		} else {
			transaction.Rollback(ctx)
		}
	}

	transaction, err = n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryDeleteOne, fmt.Errorf("start transaction to deactivate %s failed, error: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s = $2;",
		intdoment.DirectoryAuthorizationIDsRepository().RepositoryName,                       //1
		intdoment.DirectoryAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID, //2
		intdoment.DirectoryAuthorizationIDsRepository().ID,                                   //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryDeleteOne))
	if _, err := transaction.Exec(ctx, query, iamAuthRule.ID, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoDirectoryDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.DirectoryAuthorizationIDsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = NOW() WHERE %[3]s = $1;",
		intdoment.DirectoryRepository().RepositoryName, //1
		intdoment.DirectoryRepository().DeactivatedOn,  //2
		intdoment.DirectoryRepository().ID,             //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoDirectoryDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryDeleteOne, fmt.Errorf("commit transaction to update deactivation of %s failed, error: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoDirectoryFindOneForDeletionByID(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	datum *intdoment.Directory,
	columns []string,
) (*intdoment.Directory, *intdoment.IamAuthorizationRule, error) {
	directoryMModel, err := intlib.MetadataModelGet(intdoment.DirectoryRepository().RepositoryName)
	if err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(directoryMModel, intdoment.DirectoryRepository().RepositoryName, false, false); err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.DirectoryRepository().ID) {
		columns = append(columns, intdoment.DirectoryRepository().ID)
	}

	dataRows := make([]any, 0)
	iamAuthRule := new(intdoment.IamAuthorizationRule)
	if len(dataRows) == 0 {
		if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			iamCredential,
			authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_DELETE,
					RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
				},
			},
			iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			query := fmt.Sprintf(
				"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s = $2;",
				strings.Join(columns, " , "),                      //1
				intdoment.DirectoryRepository().RepositoryName,    //2
				intdoment.DirectoryRepository().ID,                //3
				intdoment.DirectoryRepository().DirectoryGroupsID, //4
			)
			n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryFindOneForDeletionByID))

			rows, err := n.db.Query(ctx, query, datum.ID[0], authContextDirectoryGroupID)
			if err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.DirectoryRepository().RepositoryName, err))
			}
			defer rows.Close()
			for rows.Next() {
				if r, err := rows.Values(); err != nil {
					return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, err)
				} else {
					dataRows = append(dataRows, r)
				}
			}
			if len(dataRows) > 0 {
				iamAuthRule = iamAuthorizationRule[0]
			}
		}
	}

	if len(dataRows) == 0 {
		if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			iamCredential,
			authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_DELETE_OTHERS,
					RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
				},
			},
			iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			cteName := fmt.Sprintf("%s_%s", intdoment.DirectoryRepository().RepositoryName, RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME)
			query := fmt.Sprintf(
				"%[1]s SELECT %[2]s FROM %[3]s WHERE %[4]s = $1 AND %[5]s;",
				RecursiveDirectoryGroupsSubGroupsCte(authContextDirectoryGroupID, cteName), //1
				strings.Join(columns, " , "),                                               //2
				intdoment.DirectoryRepository().RepositoryName,                             //3
				intdoment.DirectoryRepository().ID,                                         //4
				fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.DirectoryRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName), //5
			)
			n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryFindOneForDeletionByID))

			rows, err := n.db.Query(ctx, query, datum.ID[0])
			if err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.DirectoryRepository().RepositoryName, err))
			}
			defer rows.Close()
			for rows.Next() {
				if r, err := rows.Values(); err != nil {
					return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, err)
				} else {
					dataRows = append(dataRows, r)
				}
			}
			if len(dataRows) > 0 {
				iamAuthRule = iamAuthorizationRule[0]
			}
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(directoryMModel, nil, false, false, columns)
	if err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoDirectoryFindOneForDeletionByID))
		return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, fmt.Errorf("more than one %s found", intdoment.DirectoryRepository().RepositoryName))
	}

	directory := new(intdoment.Directory)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing directory", "directory", string(jsonData), "function", intlib.FunctionName(n.RepoDirectoryFindOneForDeletionByID))
		if err := json.Unmarshal(jsonData, directory); err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoDirectoryFindOneForDeletionByID, err)
		}
	}

	return directory, iamAuthRule, nil
}

func (n *PostrgresRepository) RepoDirectorySearch(
	ctx context.Context,
	mmsearch *intdoment.MetadataModelSearch,
	repo intdomint.IamRepository,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	startSearchDirectoryGroupID uuid.UUID,
	authContextDirectoryGroupID uuid.UUID,
	skipIfFGDisabled bool,
	skipIfDataExtraction bool,
	whereAfterJoin bool,
) (*intdoment.MetadataModelSearchResults, error) {
	pSelectQuery := NewPostgresSelectQuery(
		n.logger,
		repo,
		iamCredential,
		iamAuthorizationRules,
		startSearchDirectoryGroupID,
		authContextDirectoryGroupID,
		mmsearch.QueryConditions,
		skipIfFGDisabled,
		skipIfDataExtraction,
		whereAfterJoin,
	)
	selectQuery, err := pSelectQuery.DirectoryGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectorySearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectorySearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectorySearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectorySearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectorySearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectorySearch, err)
	}

	mmSearchResults := new(intdoment.MetadataModelSearchResults)
	mmSearchResults.MetadataModel = deep.MustCopy(mmsearch.MetadataModel)
	if len(array2DToObject.Objects()) > 0 {
		mmSearchResults.Data = array2DToObject.Objects()
	} else {
		mmSearchResults.Data = make([]any, 0)
	}

	return mmSearchResults, nil
}

func (n *PostgresSelectQuery) DirectoryGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
			},
		},
		n.iamAuthorizationRules,
	); err != nil || iamAuthorizationRule == nil {
		return nil, intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}

	quoteColumns := true
	if len(metadataModelParentPath) == 0 {
		metadataModelParentPath = "$"
		quoteColumns = false
	}
	if !n.whereAfterJoin {
		quoteColumns = false
	}

	selectQuery := SelectQuery{
		TableName: intdoment.DirectoryRepository().RepositoryName,
		Query:     "",
		Where:     make(map[string]map[int][][]string),
		Join:      make(map[string]*SelectQuery),
		JoinQuery: make([]string, 0),
	}

	if tableUid, ok := metadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID].(string); ok && len(tableUid) > 0 {
		selectQuery.TableUid = tableUid
	} else {
		return nil, intlib.FunctionNameAndError(n.DirectoryGetSelectQuery, errors.New("tableUid is empty"))
	}

	if value, err := intlibmmodel.DatabaseGetColumnFields(metadataModel, selectQuery.TableUid, false, false); err != nil {
		return nil, intlib.FunctionNameAndError(n.DirectoryGetSelectQuery, fmt.Errorf("extract database column fields failed, error: %v", err))
	} else {
		selectQuery.Columns = value
	}

	if !n.startSearchDirectoryGroupID.IsNil() {
		cteName := fmt.Sprintf("%s_%s", selectQuery.TableUid, RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME)
		cteWhere := make([]string, 0)

		if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			n.iamCredential,
			n.authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE_OTHERS,
					RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
				},
			},
			n.iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			selectQuery.DirectoryGroupsSubGroupsCTEName = cteName
			selectQuery.DirectoryGroupsSubGroupsCTE = RecursiveDirectoryGroupsSubGroupsCte(n.startSearchDirectoryGroupID, cteName)
			cteWhere = append(cteWhere, fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.DirectoryRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName))
		}

		if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			n.iamCredential,
			n.authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE,
					RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
				},
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
					RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
				},
			},
			n.iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			if len(selectQuery.DirectoryGroupsSubGroupsCTEName) == 0 {
				selectQuery.DirectoryGroupsSubGroupsCTEName = cteName
			}
			if len(selectQuery.DirectoryGroupsSubGroupsCTE) == 0 {
				selectQuery.DirectoryGroupsSubGroupsCTE = RecursiveDirectoryGroupsSubGroupsCte(n.startSearchDirectoryGroupID, cteName)
			}
			cteWhere = append(cteWhere, fmt.Sprintf("%s = '%s'", intdoment.DirectoryRepository().DirectoryGroupsID, n.startSearchDirectoryGroupID.String()))
		}

		if len(cteWhere) > 0 {
			if len(cteWhere) > 1 {
				selectQuery.DirectoryGroupsSubGroupsCTECondition = fmt.Sprintf("(%s)", strings.Join(cteWhere, " OR "))
			} else {
				selectQuery.DirectoryGroupsSubGroupsCTECondition = cteWhere[0]
			}
		}

		n.startSearchDirectoryGroupID = uuid.Nil
	}

	if _, ok := selectQuery.Columns.Fields[intdoment.DirectoryRepository().ID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryRepository().ID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryRepository().ID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.DirectoryRepository().DirectoryGroupsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryRepository().DirectoryGroupsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryRepository().DirectoryGroupsID] = value
		}
	}
	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.DirectoryRepository().DisplayName][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryRepository().DisplayName, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryRepository().DisplayName] = value
		}
	}
	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.DirectoryRepository().Data][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryRepository().Data, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryRepository().Data] = value
		}
	}
	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.DirectoryRepository().Data][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryRepository().Data, fgKeyString, PROCESS_QUERY_CONDITION_AS_JSONB, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryRepository().Data] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.DirectoryRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.DirectoryRepository().LastUpdatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryRepository().LastUpdatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryRepository().LastUpdatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.DirectoryRepository().DeactivatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryRepository().DeactivatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryRepository().DeactivatedOn] = value
		}
	}
	if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, selectQuery.TableName, "", "", "", intdoment.DirectoryRepository().FullTextSearch); len(value) > 0 {
		selectQuery.Where[intdoment.DirectoryRepository().RepositoryName] = value
	}

	directoryGroupsIDJoinDirectoryGroups := intlib.MetadataModelGenJoinKey(intdoment.DirectoryRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, directoryGroupsIDJoinDirectoryGroups); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", directoryGroupsIDJoinDirectoryGroups, err))
	} else {
		if sq, err := n.DirectoryGroupsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", directoryGroupsIDJoinDirectoryGroups, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryGroupsRepository().ID, true),                    //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.DirectoryRepository().DirectoryGroupsID, false), //2
			)

			selectQuery.Join[directoryGroupsIDJoinDirectoryGroups] = sq
		}
	}

	directoryJoinDirectoryAuthorizationIDs := intlib.MetadataModelGenJoinKey(intdoment.DirectoryRepository().RepositoryName, intdoment.DirectoryAuthorizationIDsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, directoryJoinDirectoryAuthorizationIDs); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", directoryJoinDirectoryAuthorizationIDs, err))
	} else {
		if sq, err := n.AuthorizationIDsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
			intdoment.DirectoryAuthorizationIDsRepository().RepositoryName,
			[]AuthIDsSelectQueryPKey{{Name: intdoment.DirectoryAuthorizationIDsRepository().ID, ProcessAs: PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE}},
			intdoment.DirectoryAuthorizationIDsRepository().CreationIamGroupAuthorizationsID,
			intdoment.DirectoryAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", directoryJoinDirectoryAuthorizationIDs, err))
		} else {
			if len(sq.Where) == 0 {
				sq.JoinType = JOIN_LEFT
			} else {
				sq.JoinType = JOIN_INNER
			}
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryAuthorizationIDsRepository().ID, true), //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.DirectoryRepository().ID, false),       //2
			)

			selectQuery.Join[directoryJoinDirectoryAuthorizationIDs] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}

func (n *PostrgresRepository) RepoDirectoryInsertOneAndUpdateIamCredentials(ctx context.Context, iamCredential *intdoment.IamCredentials) error {
	query := fmt.Sprintf(
		"SELECT %[2]s FROM %[1]s WHERE %[2]s = $1 AND %[3]s IS NULL;",
		intdoment.IamCredentialsRepository().RepositoryName, //1
		intdoment.IamCredentialsRepository().ID,             //2
		intdoment.IamCredentialsRepository().DirectoryID,    //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryInsertOneAndUpdateIamCredentials))

	var id uuid.UUID
	if err := n.db.QueryRow(ctx, query, iamCredential.ID[0]).Scan(&id); err != nil {
		return nil
	}

	systemGroup, err := n.RepoDirectoryGroupsFindSystemGroup(ctx, []string{intdoment.DirectoryGroupsRepository().ID})
	if err != nil {
		return err
	}

	iamAuthorizationRule := new(intdoment.IamAuthorizationRule)
	if iar, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		systemGroup.ID[0],
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_CREATE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY,
			},
		},
		nil,
	); err != nil {
		return err
	} else if iar == nil {
		return intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	} else {
		iamAuthorizationRule = iar[0]
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s, %[4]s, %[5]s) VALUES ($1, $2, $3, to_tsvector($3)) RETURNING %[6]s;",
		intdoment.DirectoryRepository().RepositoryName,    //1
		intdoment.DirectoryRepository().DirectoryGroupsID, //2
		intdoment.DirectoryRepository().Data,              //3
		intdoment.DirectoryRepository().DisplayName,       //4
		intdoment.DirectoryRepository().FullTextSearch,    //5
		intdoment.DirectoryRepository().ID,                //6
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryInsertOneAndUpdateIamCredentials))
	if err := n.db.QueryRow(ctx, query, systemGroup.ID[0], map[string]any{}, iamCredential.OpenidUserInfo[0].OpenidPreferredUsername[0]).Scan(&id); err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryInsertOneAndUpdateIamCredentials, fmt.Errorf("insert %s failed, err: %v", intdoment.DirectoryRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.DirectoryAuthorizationIDsRepository().RepositoryName,                   //1
		intdoment.DirectoryAuthorizationIDsRepository().ID,                               //2
		intdoment.DirectoryAuthorizationIDsRepository().CreationIamGroupAuthorizationsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryInsertOneAndUpdateIamCredentials))
	if _, err := n.db.Exec(ctx, query, id, iamAuthorizationRule.ID); err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryInsertOneAndUpdateIamCredentials, fmt.Errorf("insert %s failed, err: %v", intdoment.DirectoryAuthorizationIDsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s = $2;",
		intdoment.IamCredentialsRepository().RepositoryName, //1
		intdoment.IamCredentialsRepository().DirectoryID,    //2
		intdoment.IamCredentialsRepository().ID,             //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryInsertOneAndUpdateIamCredentials))
	if _, err := n.db.Exec(ctx, query, id, iamCredential.ID[0]); err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryInsertOneAndUpdateIamCredentials, fmt.Errorf("update %s failed, err: %v", intdoment.IamCredentialsRepository().RepositoryName, err))
	}

	return nil
}
