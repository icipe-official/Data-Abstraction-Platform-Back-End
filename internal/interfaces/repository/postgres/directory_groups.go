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

func (n *PostrgresRepository) RepoDirectoryGroupsDeleteOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.DirectoryGroups) error {
	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryGroupsDeleteOne, fmt.Errorf("start transaction to delete %s failed, error: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s = $1;",
		intdoment.DirectoryGroupsAuthorizationIDsRepository().RepositoryName, //1
		intdoment.DirectoryGroupsAuthorizationIDsRepository().ID,             //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		query = fmt.Sprintf(
			"DELETE FROM %[1]s WHERE %[2]s = $1;",
			intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName,    //1
			intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID, //2
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsDeleteOne))
		if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
			query = fmt.Sprintf(
				"DELETE FROM %[1]s WHERE %[2]s = $1;",
				intdoment.MetadataModelsDirectoryRepository().RepositoryName,    //1
				intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID, //2
			)
			n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsDeleteOne))
			if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
				query = fmt.Sprintf(
					"DELETE FROM %[1]s WHERE %[2]s = $1;",
					intdoment.DirectoryGroupsSubGroupsRepository().RepositoryName, //1
					intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID,     //2
				)
				n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsDeleteOne))
				if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
					query = fmt.Sprintf(
						"DELETE FROM %[1]s WHERE %[2]s = $1;",
						intdoment.DirectoryGroupsRepository().RepositoryName, //1
						intdoment.DirectoryGroupsRepository().ID,             //2
						intdoment.DirectoryGroupsRepository().Data,           //3
					)
					n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsDeleteOne))
					if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
						if err := transaction.Commit(ctx); err != nil {
							return intlib.FunctionNameAndError(n.RepoDirectoryGroupsDeleteOne, fmt.Errorf("commit transaction to delete %s failed, error: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
						}
						return nil
					} else {
						transaction.Rollback(ctx)
					}
				}
			}
		}
	}

	transaction, err = n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryGroupsDeleteOne, fmt.Errorf("start transaction to deactivate %s failed, error: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s = $2;",
		intdoment.DirectoryGroupsAuthorizationIDsRepository().RepositoryName,                       //1
		intdoment.DirectoryGroupsAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID, //2
		intdoment.DirectoryGroupsAuthorizationIDsRepository().ID,                                   //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsDeleteOne))
	if _, err := transaction.Exec(ctx, query, iamAuthRule.ID, datum.ID[0]); err != nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoDirectoryGroupsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.DirectoryGroupsAuthorizationIDsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = NOW() WHERE %[3]s = $1 AND %[4]s IS NOT NULL;",
		intdoment.DirectoryGroupsRepository().RepositoryName, //1
		intdoment.DirectoryGroupsRepository().DeactivatedOn,  //2
		intdoment.DirectoryGroupsRepository().ID,             //3
		intdoment.DirectoryGroupsRepository().Data,           //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err != nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoDirectoryGroupsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return intlib.FunctionNameAndError(n.RepoDirectoryGroupsDeleteOne, fmt.Errorf("commit transaction to update deactivation of %s failed, error: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoDirectoryGroupsCheckIfSystemGroup(ctx context.Context, directoryGroupID uuid.UUID) (bool, error) {
	query := fmt.Sprintf(
		"SELECT %[3]s.%[5]s FROM %[1]s INNER JOIN %[3]s ON %[1]s.%[2]s = %[3]s.%[5]s AND %[3]s.%[5]s = $1 AND %[3]s.%[4]s IS NOT NULL;",
		intdoment.DirectoryGroupsSubGroupsRepository().RepositoryName, //1
		intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID,     //2
		intdoment.DirectoryGroupsRepository().RepositoryName,          //3
		intdoment.DirectoryGroupsRepository().Data,                    //4
		intdoment.DirectoryGroupsRepository().ID,                      //5
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsCheckIfSystemGroup))
	var dest any
	if err := n.db.QueryRow(ctx, query, directoryGroupID).Scan(dest); err != nil {
		if err == pgx.ErrNoRows {
			return true, nil
		} else {
			return false, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCheckIfSystemGroup, fmt.Errorf("get non-system %s failed, err: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
		}
	} else {
		return false, nil
	}
}

func (n *PostrgresRepository) RepoDirectoryGroupsUpdateOne(ctx context.Context, datum *intdoment.DirectoryGroups, fieldAnyMetadataModelGet intdomint.FieldAnyMetadataModel) error {
	valuesToUpdate := make([]any, 0)
	valueToUpdateQuery := make([]string, 0)
	columnsToUpdate := make([]string, 0)
	nextPlaceholder := 1
	if v, vQ, c := n.RepoDirectoryGroupsValidateAndGetColumnsAndData(ctx, datum.ID[0], &nextPlaceholder, fieldAnyMetadataModelGet, datum, false); len(c) == 0 || len(v) == 0 || len(vQ) == 0 {
		return intlib.NewError(http.StatusBadRequest, "no values to update")
	} else {
		valuesToUpdate = append(valuesToUpdate, v...)
		columnsToUpdate = append(columnsToUpdate, c...)
		valueToUpdateQuery = append(valueToUpdateQuery, vQ...)
	}

	query := fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s WHERE %[3]s = %[4]s AND %[5]s IS NULL AND %[6]s IS NOT NULL;",
		intdoment.DirectoryGroupsRepository().RepositoryName,               //1
		GetUpdateSetColumnsWithVQuery(columnsToUpdate, valueToUpdateQuery), //2
		intdoment.DirectoryGroupsRepository().ID,                           //3
		GetandUpdateNextPlaceholder(&nextPlaceholder),                      //4
		intdoment.DirectoryGroupsRepository().DeactivatedOn,                //5
		intdoment.DirectoryGroupsRepository().Data,                         //6
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsUpdateOne))

	valuesToUpdate = append(valuesToUpdate, datum.ID[0])
	if _, err := n.db.Exec(ctx, query, valuesToUpdate...); err != nil {
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryUpdateOne, fmt.Errorf("update %s failed, err: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoDirectoryGroupsInsertOne(
	ctx context.Context,
	datum *intdoment.DirectoryGroups,
	authContextDirectoryGroupID uuid.UUID,
	iamAuthorizationRule *intdoment.IamAuthorizationRule,
	fieldAnyMetadataModelGet intdomint.FieldAnyMetadataModel,
	columns []string,
) (*intdoment.DirectoryGroups, error) {
	directoryGroupsMetadataModel, err := intlib.MetadataModelGet(intdoment.DirectoryGroupsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(directoryGroupsMetadataModel, intdoment.DirectoryGroupsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.DirectoryGroupsRepository().ID) {
		columns = append(columns, intdoment.DirectoryGroupsRepository().ID)
	}

	valuesToInsert := make([]any, 0)
	valueToInsertQuery := make([]string, 0)
	columnsToInsert := make([]string, 0)
	nextPlaceholder := 1
	if v, vQ, c := n.RepoDirectoryGroupsValidateAndGetColumnsAndData(ctx, authContextDirectoryGroupID, &nextPlaceholder, fieldAnyMetadataModelGet, datum, true); len(c) == 0 || len(v) == 0 || len(vQ) == 0 {
		return nil, intlib.NewError(http.StatusBadRequest, "no values to insert")
	} else {
		valuesToInsert = append(valuesToInsert, v...)
		columnsToInsert = append(columnsToInsert, c...)
		valueToInsertQuery = append(valueToInsertQuery, vQ...)
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s) VALUES(%[3]s) RETURNING %[4]s;",
		intdoment.DirectoryGroupsRepository().RepositoryName, //1
		strings.Join(columnsToInsert, " , "),                 //2
		strings.Join(valueToInsertQuery, " , "),              //3
		strings.Join(columns, " , "),                         //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsInsertOne))

	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, fmt.Errorf("start transaction to create %s failed, error: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}

	rows, err := transaction.Query(ctx, query, valuesToInsert...)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(directoryGroupsMetadataModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		transaction.Rollback(ctx)
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.DirectoryGroupsRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		transaction.Rollback(ctx)
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoDirectoryGroupsInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, fmt.Errorf("more than one %s found", intdoment.DirectoryGroupsRepository().RepositoryName))
	}

	directoryGroup := new(intdoment.DirectoryGroups)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing directoryGroup", "directoryGroup", string(jsonData), "function", intlib.FunctionName(n.RepoDirectoryGroupsInsertOne))
		if err := json.Unmarshal(jsonData, directoryGroup); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsInsertOne, err)
		}
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.DirectoryGroupsAuthorizationIDsRepository().RepositoryName,                   //1
		intdoment.DirectoryGroupsAuthorizationIDsRepository().ID,                               //2
		intdoment.DirectoryGroupsAuthorizationIDsRepository().CreationIamGroupAuthorizationsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsInsertOne))

	if _, err := transaction.Exec(ctx, query, directoryGroup.ID[0], iamAuthorizationRule.ID); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.DirectoryGroupsAuthorizationIDsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.DirectoryGroupsSubGroupsRepository().RepositoryName, //1
		intdoment.DirectoryGroupsSubGroupsRepository().ParentGroupID,  //2
		intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID,     //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsInsertOne))

	if _, err := transaction.Exec(ctx, query, authContextDirectoryGroupID, directoryGroup.ID[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.DirectoryGroupsSubGroupsRepository().RepositoryName, err))
	}

	//TODO: Set default metadata_models

	if err := transaction.Commit(ctx); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, fmt.Errorf("commit transaction to create %s failed, error: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}

	return directoryGroup, nil
}

func (n *PostrgresRepository) RepoDirectoryGroupsValidateAndGetColumnsAndData(ctx context.Context, directoryGroupID uuid.UUID, nextPlaceholder *int, fieldAnyMetadataModelGet intdomint.FieldAnyMetadataModel, datum *intdoment.DirectoryGroups, insert bool) ([]any, []string, []string) {
	values := make([]any, 0)
	valuesQuery := make([]string, 0)
	columns := make([]string, 0)

	fullTextSearchValue := make([]string, 0)
	if len(datum.Data) > 0 {
		if dMap, ok := datum.Data[0].(map[string]any); ok {
			columns = append(columns, intdoment.DirectoryGroupsRepository().Data)
			values = append(values, dMap)
			valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
		} else {
			if insert {
				datum.Data[0] = map[string]any{}
				columns = append(columns, intdoment.DirectoryGroupsRepository().Data)
				values = append(values, datum.Data[0])
				valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
			}
		}
	} else {
		if insert {
			datum.Data = []any{map[string]any{}}
			columns = append(columns, intdoment.DirectoryGroupsRepository().Data)
			values = append(values, datum.Data[0])
			valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
		}
	}

	if len(datum.DisplayName) > 0 {
		columns = append(columns, intdoment.DirectoryGroupsRepository().DisplayName)
		values = append(values, datum.DisplayName[0])
		valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
		fullTextSearchValue = append(fullTextSearchValue, datum.DisplayName...)
	}

	if len(datum.Description) > 0 {
		columns = append(columns, intdoment.DirectoryGroupsRepository().Description)
		values = append(values, datum.Description[0])
		valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
		fullTextSearchValue = append(fullTextSearchValue, datum.Description...)
	}

	if slices.Contains(columns, intdoment.DirectoryGroupsRepository().Data) {
		if mm, err := fieldAnyMetadataModelGet.GetMetadataModel(ctx, intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, "$", intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, []any{directoryGroupID}); err == nil {
			if value := MetadataModelExtractFullTextSearchValue(mm, datum.Data[0]); len(value) > 0 {
				fullTextSearchValue = append(fullTextSearchValue, value...)
			}
		}
	}

	if len(fullTextSearchValue) > 0 {
		columns = append(columns, intdoment.DirectoryGroupsRepository().FullTextSearch)
		values = append(values, strings.Join(fullTextSearchValue, " "))
		valuesQuery = append(valuesQuery, fmt.Sprintf("to_tsvector(%s)", GetandUpdateNextPlaceholder(nextPlaceholder)))
	}

	return values, valuesQuery, columns
}

func (n *PostrgresRepository) RepoDirectoryGroupsSearch(
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
	selectQuery, err := pSelectQuery.DirectoryGroupsGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSearch, err)
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

func (n *PostgresSelectQuery) DirectoryGroupsGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY_GROUPS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY_GROUPS,
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
		TableName: intdoment.DirectoryGroupsRepository().RepositoryName,
		Query:     "",
		Where:     make(map[string]map[int][][]string),
		Join:      make(map[string]*SelectQuery),
		JoinQuery: make([]string, 0),
	}

	if tableUid, ok := metadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID].(string); ok && len(tableUid) > 0 {
		selectQuery.TableUid = tableUid
	} else {
		return nil, intlib.FunctionNameAndError(n.DirectoryGroupsGetSelectQuery, errors.New("tableUid is empty"))
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
					RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY_GROUPS,
				},
			},
			n.iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			selectQuery.DirectoryGroupsSubGroupsCTEName = cteName
			selectQuery.DirectoryGroupsSubGroupsCTE = RecursiveDirectoryGroupsSubGroupsCte(n.startSearchDirectoryGroupID, cteName)
			cteWhere = append(cteWhere, fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.DirectoryGroupsRepository().ID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName))
		}

		if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			n.iamCredential,
			n.authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE,
					RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY_GROUPS,
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
			cteWhere = append(cteWhere, fmt.Sprintf("%s = '%s'", intdoment.DirectoryGroupsRepository().ID, n.startSearchDirectoryGroupID.String()))
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

	if value, err := intlibmmodel.DatabaseGetColumnFields(metadataModel, selectQuery.TableUid, false, false); err != nil {
		return nil, intlib.FunctionNameAndError(n.DirectoryGroupsGetSelectQuery, fmt.Errorf("extract database column fields failed, error: %v", err))
	} else {
		selectQuery.Columns = value
	}

	if _, ok := selectQuery.Columns.Fields[intdoment.DirectoryGroupsRepository().ID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryGroupsRepository().ID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryGroupsRepository().ID] = value
		}
	}
	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.DirectoryGroupsRepository().DisplayName][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryGroupsRepository().DisplayName, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryGroupsRepository().DisplayName] = value
		}
	}
	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.DirectoryGroupsRepository().Description][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryGroupsRepository().Description, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryGroupsRepository().Description] = value
		}
	}
	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.DirectoryGroupsRepository().Data][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryGroupsRepository().Data, fgKeyString, PROCESS_QUERY_CONDITION_AS_JSONB, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryGroupsRepository().Data] = value
		}
	}
	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.DirectoryGroupsRepository().Data][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryGroupsRepository().Data, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryGroupsRepository().Data] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.DirectoryGroupsRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryGroupsRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryGroupsRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.DirectoryGroupsRepository().LastUpdatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryGroupsRepository().LastUpdatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryGroupsRepository().LastUpdatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.DirectoryGroupsRepository().DeactivatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.DirectoryGroupsRepository().DeactivatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.DirectoryGroupsRepository().DeactivatedOn] = value
		}
	}
	if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, selectQuery.TableName, "", "", "", intdoment.DirectoryGroupsRepository().FullTextSearch); len(value) > 0 {
		selectQuery.Where[intdoment.DirectoryGroupsRepository().RepositoryName] = value
	}

	directoryGroupsJoinDirectoryGroupAuthorizationIDs := intlib.MetadataModelGenJoinKey(intdoment.DirectoryGroupsRepository().RepositoryName, intdoment.DirectoryGroupsAuthorizationIDsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, directoryGroupsJoinDirectoryGroupAuthorizationIDs); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", directoryGroupsJoinDirectoryGroupAuthorizationIDs, err))
	} else {
		if sq, err := n.AuthorizationIDsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
			intdoment.DirectoryGroupsAuthorizationIDsRepository().RepositoryName,
			[]AuthIDsSelectQueryPKey{{Name: intdoment.DirectoryGroupsAuthorizationIDsRepository().ID, ProcessAs: PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE}},
			intdoment.DirectoryGroupsAuthorizationIDsRepository().CreationIamGroupAuthorizationsID,
			intdoment.DirectoryGroupsAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", directoryGroupsJoinDirectoryGroupAuthorizationIDs, err))
		} else {
			if len(sq.Where) == 0 {
				sq.JoinType = JOIN_LEFT
			} else {
				sq.JoinType = JOIN_INNER
			}
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryGroupsAuthorizationIDsRepository().ID, true), //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.DirectoryGroupsRepository().ID, false),       //2
			)

			selectQuery.Join[directoryGroupsJoinDirectoryGroupAuthorizationIDs] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}

func (n *PostrgresRepository) RepoDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID, columns []string) (*intdoment.DirectoryGroups, error) {
	directoryGroupsMModel, err := intlib.MetadataModelGet(intdoment.DirectoryGroupsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindOneByIamCredentialID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(directoryGroupsMModel, intdoment.DirectoryGroupsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindOneByIamCredentialID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, fmt.Sprintf("%s.%s", intdoment.DirectoryGroupsRepository().RepositoryName, intdoment.DirectoryGroupsRepository().ID)) {
		columns = append(columns, intdoment.DirectoryGroupsRepository().ID)
	}

	selectColumns := make([]string, len(columns))
	for cIndex, cValue := range columns {
		selectColumns[cIndex] = intdoment.DirectoryGroupsRepository().RepositoryName + "." + cValue
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s INNER JOIN %[3]s INNER JOIN %[4]s ON %[4]s.%[5]s = $1 AND %[4]s.%[6]s = %[3]s.%[7]s ON %[3]s.%[8]s = %[2]s.%[9]s;",
		strings.Join(selectColumns, ","),                     //1
		intdoment.DirectoryGroupsRepository().RepositoryName, //2
		intdoment.DirectoryRepository().RepositoryName,       //3
		intdoment.IamCredentialsRepository().RepositoryName,  //4
		intdoment.IamCredentialsRepository().ID,              //5
		intdoment.IamCredentialsRepository().DirectoryID,     //6
		intdoment.DirectoryRepository().ID,                   //7
		intdoment.DirectoryRepository().DirectoryGroupsID,    //8
		intdoment.DirectoryGroupsRepository().ID,             //9
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsFindOneByIamCredentialID))

	rows, err := n.db.Query(ctx, query, iamCredentialID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindOneByIamCredentialID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindOneByIamCredentialID, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(directoryGroupsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindOneByIamCredentialID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindOneByIamCredentialID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoDirectoryGroupsFindOneByIamCredentialID))
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindOneByIamCredentialID, fmt.Errorf("more than one %s found", intdoment.DirectoryGroupsRepository().RepositoryName))
	}

	directoryGroup := new(intdoment.DirectoryGroups)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindOneByIamCredentialID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing directoryGroup", "directoryGroup", string(jsonData), "function", intlib.FunctionName(n.RepoDirectoryGroupsFindOneByIamCredentialID))
		if err := json.Unmarshal(jsonData, directoryGroup); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindOneByIamCredentialID, err)
		}
	}

	return directoryGroup, nil
}

func (n *PostrgresRepository) RepoDirectoryGroupsFindSystemGroupRuleAuthorizations(ctx context.Context) ([]intdoment.GroupRuleAuthorizations, error) {
	systemGroup, err := n.RepoDirectoryGroupsFindSystemGroup(ctx, []string{intdoment.DirectoryGroupsRepository().ID})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroupRuleAuthorizations, err)
	}

	groupRuleAuthorizationMModel, err := intlib.MetadataModelGet(intdoment.GroupRuleAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroupRuleAuthorizations, err)
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s IS NULL AND (%[5]s = $2 OR %[5]s = $3 OR %[5]s = $4 OR %[5]s = $5 OR %[5]s = $6);",
		intdoment.GroupRuleAuthorizationsRepository().ID,                           //1
		intdoment.GroupRuleAuthorizationsRepository().RepositoryName,               //2
		intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID,            //3
		intdoment.GroupRuleAuthorizationsRepository().DeactivatedOn,                //4
		intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup, //5
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsFindSystemGroupRuleAuthorizations))

	rows, err := n.db.Query(ctx, query, systemGroup.ID[0], intdoment.AUTH_RULE_GROUP_GROUP_RULE_AUTHORIZATIONS, intdoment.AUTH_RULE_GROUP_IAM_GROUP_AUTHORIZATIONS, intdoment.AUTH_RULE_GROUP_DIRECTORY, intdoment.AUTH_RULE_GROUP_IAM_CREDENTIALS, intdoment.AUTH_RULE_GROUP_DIRECTORY_GROUPS)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroupRuleAuthorizations, fmt.Errorf("retrieve %s failed, err: %v", intdoment.IamCredentialsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroupRuleAuthorizations, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(groupRuleAuthorizationMModel, nil, false, false, []string{intdoment.GroupRuleAuthorizationsRepository().ID})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroupRuleAuthorizations, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroupRuleAuthorizations, err)
	}

	groupRuleAuthorizations := make([]intdoment.GroupRuleAuthorizations, 0)
	if jsonData, err := json.Marshal(array2DToObject.Objects()); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroupRuleAuthorizations, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing groupRuleAuthorizations", "groupRuleAuthorizations", string(jsonData), "function", intlib.FunctionName(n.RepoDirectoryGroupsSearch))
		if err := json.Unmarshal(jsonData, &groupRuleAuthorizations); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroupRuleAuthorizations, err)
		}
	}

	return groupRuleAuthorizations, nil
}

func (n *PostrgresRepository) RepoDirectoryGroupsCreateSystemGroup(ctx context.Context, columns []string) (*intdoment.DirectoryGroups, error) {
	directoryGroupsMModel, err := intlib.MetadataModelGet(intdoment.DirectoryGroupsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(directoryGroupsMModel, intdoment.DirectoryGroupsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.DirectoryGroupsRepository().ID) {
		columns = append(columns, intdoment.DirectoryGroupsRepository().ID)
	}

	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, fmt.Errorf("start transaction to create system group failed, error: %v", err))
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s, %[4]s) VALUES ($1, NULL,to_tsvector($1)) RETURNING %[5]s;",
		intdoment.DirectoryGroupsRepository().RepositoryName, //1
		intdoment.DirectoryGroupsRepository().DisplayName,    //2
		intdoment.DirectoryGroupsRepository().Data,           //3
		intdoment.DirectoryGroupsRepository().FullTextSearch, //4
		strings.Join(columns, ","),                           //5
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsCreateSystemGroup))

	rows, err := transaction.Query(ctx, query, "system")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, fmt.Errorf("insert system %s failed, err: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}
	if len(dataRows) < 1 {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, fmt.Errorf("%v data is empty", intdoment.DirectoryGroupsRepository().RepositoryName))
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(directoryGroupsMModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, err)
	}
	systemGroup := new(intdoment.DirectoryGroups)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, err
	} else {
		if err := json.Unmarshal(jsonData, systemGroup); err != nil {
			transaction.Rollback(ctx)
			return nil, err
		}
	}

	groupAuthorizationRulesMModel, err := intlib.MetadataModelGet(intdoment.GroupAuthorizationRulesRepository().RepositoryName)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
	}
	columns = []string{intdoment.GroupAuthorizationRulesRepository().ID, intdoment.GroupAuthorizationRulesRepository().RuleGroup}
	query = fmt.Sprintf(
		"SELECT %[1]s, %[2]s FROM %[3]s WHERE %[2]s = $1 OR %[2]s = $2 OR %[2]s = $3 OR %[2]s = $4 OR %[2]s = $5;",
		intdoment.GroupAuthorizationRulesRepository().ID,             //1
		intdoment.GroupAuthorizationRulesRepository().RuleGroup,      //2
		intdoment.GroupAuthorizationRulesRepository().RepositoryName, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsCreateSystemGroup))

	rows, err = transaction.Query(ctx, query, intdoment.AUTH_RULE_GROUP_GROUP_RULE_AUTHORIZATIONS, intdoment.AUTH_RULE_GROUP_IAM_GROUP_AUTHORIZATIONS, intdoment.AUTH_RULE_GROUP_DIRECTORY, intdoment.AUTH_RULE_GROUP_IAM_CREDENTIALS, intdoment.AUTH_RULE_GROUP_DIRECTORY_GROUPS)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, fmt.Errorf("retrieve %s failed, err: %v", intdoment.GroupAuthorizationRulesRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows = make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}
	if len(dataRows) < 1 {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, fmt.Errorf("%v data is empty", intdoment.GroupAuthorizationRulesRepository().RepositoryName))
	}

	array2DToObject, err = intlibmmodel.NewConvert2DArrayToObjects(groupAuthorizationRulesMModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, err)
	}
	groupAuthorizationRules := make([]intdoment.GroupAuthorizationRules, 0)
	if jsonData, err := json.Marshal(array2DToObject.Objects()); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, err)
	} else {
		if err := json.Unmarshal(jsonData, &groupAuthorizationRules); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, err)
		}
	}

	for _, gar := range groupAuthorizationRules {
		query = fmt.Sprintf(
			"SELECT * FROM %[1]s WHERE %[2]s = $1 AND %[3]s = $2 AND %[4]s = $3;",
			intdoment.GroupRuleAuthorizationsRepository().RepositoryName,               //1
			intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID,            //2
			intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID,    //3
			intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup, //4
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsCreateSystemGroup))

		if rows := transaction.QueryRow(
			ctx,
			query,
			systemGroup.ID[0],
			gar.GroupAuthorizationRulesID[0].ID[0],
			gar.GroupAuthorizationRulesID[0].RuleGroup[0],
		); rows.Scan() == pgx.ErrNoRows {
			query = fmt.Sprintf(
				"INSERT INTO %[1]s (%[2]s, %[3]s, %[4]s) VALUES ($1, $2, $3);",
				intdoment.GroupRuleAuthorizationsRepository().RepositoryName,               //1
				intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID,            //2
				intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID,    //3
				intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup, //4
			)
			n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsCreateSystemGroup))

			if _, err := transaction.Exec(ctx, query, systemGroup.ID[0], gar.GroupAuthorizationRulesID[0].ID[0], gar.GroupAuthorizationRulesID[0].RuleGroup[0]); err != nil {
				transaction.Rollback(ctx)
				return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, fmt.Errorf("insert %s failed, err: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
			}
		} else {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, fmt.Errorf("get individual %s failed, err: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, rows.Scan()))
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsCreateSystemGroup, fmt.Errorf("commit transaction to create system group failed, error: %v", err))
	}

	return systemGroup, nil
}

func (n *PostrgresRepository) RepoDirectoryGroupsFindSystemGroup(ctx context.Context, columns []string) (*intdoment.DirectoryGroups, error) {
	directoryGroupsMModel, err := intlib.MetadataModelGet(intdoment.DirectoryGroupsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(directoryGroupsMModel, intdoment.DirectoryGroupsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE %[3]s IS NULL;",
		strings.Join(columns, ","),                           //1
		intdoment.DirectoryGroupsRepository().RepositoryName, //2
		intdoment.DirectoryGroupsRepository().Data,           //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsFindSystemGroup))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, fmt.Errorf("retrieve system %s failed, err: %v", intdoment.DirectoryGroupsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(directoryGroupsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())))
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, fmt.Errorf("more than one system %s found", intdoment.DirectoryGroupsRepository().RepositoryName))
	}

	systemGroup := new(intdoment.DirectoryGroups)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing systemGroup", "systemGroup", string(jsonData), "function", intlib.FunctionName(n.RepoDirectoryGroupsFindSystemGroup))
		if err := json.Unmarshal(jsonData, systemGroup); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsFindSystemGroup, err)
		}
	}

	return systemGroup, nil
}
