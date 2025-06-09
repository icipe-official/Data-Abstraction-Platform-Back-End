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

func (n *PostrgresRepository) RepoAbstractionsDirectoryGroupsDeleteOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.AbstractionsDirectoryGroups) error {
	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsDeleteOne, fmt.Errorf("start transaction to delete %s failed, error: %v", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s = $1;",
		intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().RepositoryName,    //1
		intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().DirectoryGroupsID, //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDirectoryGroupsDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.DirectoryGroupsID[0]); err == nil {
		query = fmt.Sprintf(
			"DELETE FROM %[1]s WHERE %[2]s = $1;",
			intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //1
			intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //2
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDirectoryGroupsDeleteOne))
		if _, err := transaction.Exec(ctx, query, datum.DirectoryGroupsID[0]); err == nil {
			if err := transaction.Commit(ctx); err != nil {
				return intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsDeleteOne, fmt.Errorf("commit transaction to delete %s failed, error: %v", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, err))
			}
			return nil
		} else {
			transaction.Rollback(ctx)
		}
	}

	transaction, err = n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsDeleteOne, fmt.Errorf("start transaction to deactivate %s failed, error: %v", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s = $2;",
		intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().RepositoryName,                       //1
		intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID, //2
		intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().DirectoryGroupsID,                    //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDirectoryGroupsDeleteOne))
	if _, err := transaction.Exec(ctx, query, iamAuthRule.ID, datum.DirectoryGroupsID[0]); err != nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = NOW() WHERE %[3]s = $1;",
		intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //1
		intdoment.AbstractionsDirectoryGroupsRepository().DeactivatedOn,     //2
		intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDirectoryGroupsDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.DirectoryGroupsID[0]); err != nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsDeleteOne, fmt.Errorf("commit transaction to update deactivation of %s failed, error: %v", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoAbstractionsDirectoryGroupsUpdateOne(ctx context.Context, datum *intdoment.AbstractionsDirectoryGroups) error {
	valuesToUpdate := make([]any, 0)
	columnsToUpdate := make([]string, 0)
	if v, c, err := n.RepoAbstractionsDirectoryGroupsValidateAndGetColumnsAndData(datum, false); err != nil {
		return err
	} else if len(c) == 0 || len(v) == 0 {
		return intlib.NewError(http.StatusBadRequest, "no values to update")
	} else {
		valuesToUpdate = append(valuesToUpdate, v...)
		columnsToUpdate = append(columnsToUpdate, c...)
	}

	nextPlaceholder := 1
	query := fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s WHERE %[3]s = %[4]s AND %[5]s IS NULL;",
		intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //1
		GetUpdateSetColumns(columnsToUpdate, &nextPlaceholder),              //2
		intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //3
		GetandUpdateNextPlaceholder(&nextPlaceholder),                       //4
		intdoment.AbstractionsDirectoryGroupsRepository().DeactivatedOn,     //5
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDirectoryGroupsUpdateOne))
	valuesToUpdate = append(valuesToUpdate, datum.DirectoryGroupsID[0])

	if _, err := n.db.Exec(ctx, query, valuesToUpdate...); err != nil {
		return intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsUpdateOne, fmt.Errorf("update %s failed, err: %v", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoAbstractionsDirectoryGroupsInsertOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.AbstractionsDirectoryGroups, columns []string) (*intdoment.AbstractionsDirectoryGroups, error) {
	abstractionsDirectoryGroupsMModel, err := intlib.MetadataModelGet(intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(abstractionsDirectoryGroupsMModel, intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID) {
		columns = append(columns, intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID)
	}

	valuesToInsert := []any{datum.DirectoryGroupsID[0]}
	columnsToInsert := []string{intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID}
	if v, c, err := n.RepoAbstractionsDirectoryGroupsValidateAndGetColumnsAndData(datum, true); err != nil {
		return nil, err
	} else if len(c) == 0 || len(v) == 0 {
		return nil, intlib.NewError(http.StatusBadRequest, "no values to insert")
	} else {
		valuesToInsert = append(valuesToInsert, v...)
		columnsToInsert = append(columnsToInsert, c...)
	}

	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, fmt.Errorf("start transaction to create %s failed, error: %v", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s) VALUES (%[3]s) RETURNING %[4]s;",
		intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, //1
		strings.Join(columnsToInsert, " , "),                             //2
		GetQueryPlaceholderString(len(valuesToInsert), &[]int{1}[0]),     //3
		strings.Join(columns, " , "),                                     //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDirectoryGroupsInsertOne))

	rows, err := transaction.Query(ctx, query, valuesToInsert...)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(abstractionsDirectoryGroupsMModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		transaction.Rollback(ctx)
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		transaction.Rollback(ctx)
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoAbstractionsDirectoryGroupsInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, fmt.Errorf("more than one %s found", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName))
	}

	abtractionsDirectoryGroup := new(intdoment.AbstractionsDirectoryGroups)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing abtractionsDirectoryGroup", "abtractionsDirectoryGroup", string(jsonData), "function", intlib.FunctionName(n.RepoAbstractionsDirectoryGroupsInsertOne))
		if err := json.Unmarshal(jsonData, abtractionsDirectoryGroup); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, err)
		}
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().RepositoryName,                   //1
		intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().DirectoryGroupsID,                //2
		intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().CreationIamGroupAuthorizationsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDirectoryGroupsInsertOne))

	if _, err := transaction.Exec(ctx, query, abtractionsDirectoryGroup.DirectoryGroupsID[0], iamAuthRule.ID); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsInsertOne, fmt.Errorf("commit transaction to create %s failed, error: %v", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, err))
	}

	return abtractionsDirectoryGroup, nil

}

func (n *PostrgresRepository) RepoAbstractionsDirectoryGroupsValidateAndGetColumnsAndData(datum *intdoment.AbstractionsDirectoryGroups, insert bool) ([]any, []string, error) {
	values := make([]any, 0)
	columns := make([]string, 0)

	if len(datum.MetadataModelsID) == 0 || len(datum.MetadataModelsID[0]) < 4 {
		if insert {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.AbstractionsDirectoryGroupsRepository().MetadataModelsID)
		}
	} else {
		values = append(values, datum.MetadataModelsID[0])
		columns = append(columns, intdoment.AbstractionsDirectoryGroupsRepository().MetadataModelsID)
	}

	if len(datum.Description) > 0 && len(datum.Description[0]) > 0 {
		values = append(values, datum.Description[0])
		columns = append(columns, intdoment.AbstractionsDirectoryGroupsRepository().Description)
	}

	if len(datum.AbstractionReviewQuorum) > 0 {
		values = append(values, datum.AbstractionReviewQuorum[0])
		columns = append(columns, intdoment.AbstractionsDirectoryGroupsRepository().AbstractionReviewQuorum)
	}

	if len(datum.ViewAuthorized) > 0 {
		values = append(values, datum.ViewAuthorized[0])
		columns = append(columns, intdoment.AbstractionsDirectoryGroupsRepository().ViewAuthorized)
	}

	if len(datum.ViewUnauthorized) > 0 {
		values = append(values, datum.ViewUnauthorized[0])
		columns = append(columns, intdoment.AbstractionsDirectoryGroupsRepository().ViewUnauthorized)
	}

	return values, columns, nil
}

func (n *PostrgresRepository) RepoAbstractionsDirectoryGroupsSearch(
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
	selectQuery, err := pSelectQuery.AbstractionsDirectoryGroupsGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDirectoryGroupsSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsDirectoryGroupsSearch, err)
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

func (n *PostgresSelectQuery) AbstractionsDirectoryGroupsGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS_DIRECTORY_GROUPS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS_DIRECTORY_GROUPS,
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
		TableName: intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,
		Query:     "",
		Where:     make(map[string]map[int][][]string),
		Join:      make(map[string]*SelectQuery),
		JoinQuery: make([]string, 0),
	}

	if tableUid, ok := metadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID].(string); ok && len(tableUid) > 0 {
		selectQuery.TableUid = tableUid
	} else {
		return nil, intlib.FunctionNameAndError(n.AbstractionsDirectoryGroupsGetSelectQuery, errors.New("tableUid is empty"))
	}

	if value, err := intlibmmodel.DatabaseGetColumnFields(metadataModel, selectQuery.TableUid, false, false); err != nil {
		return nil, intlib.FunctionNameAndError(n.AbstractionsDirectoryGroupsGetSelectQuery, fmt.Errorf("extract database column fields failed, error: %v", err))
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
					RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS_DIRECTORY_GROUPS,
				},
			},
			n.iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			selectQuery.DirectoryGroupsSubGroupsCTEName = cteName
			selectQuery.DirectoryGroupsSubGroupsCTE = RecursiveDirectoryGroupsSubGroupsCte(n.startSearchDirectoryGroupID, cteName)
			cteWhere = append(cteWhere, fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName))
		}

		if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			n.iamCredential,
			n.authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE,
					RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS_DIRECTORY_GROUPS,
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
			cteWhere = append(cteWhere, fmt.Sprintf("%s = '%s'", intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, n.startSearchDirectoryGroupID.String()))
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

	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsDirectoryGroupsRepository().MetadataModelsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsDirectoryGroupsRepository().MetadataModelsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsDirectoryGroupsRepository().MetadataModelsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsDirectoryGroupsRepository().Description][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsDirectoryGroupsRepository().Description, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsDirectoryGroupsRepository().Description] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsDirectoryGroupsRepository().AbstractionReviewQuorum][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsDirectoryGroupsRepository().AbstractionReviewQuorum, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsDirectoryGroupsRepository().AbstractionReviewQuorum] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsDirectoryGroupsRepository().ViewAuthorized][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsDirectoryGroupsRepository().ViewAuthorized, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsDirectoryGroupsRepository().ViewAuthorized] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsDirectoryGroupsRepository().ViewUnauthorized][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsDirectoryGroupsRepository().ViewUnauthorized, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsDirectoryGroupsRepository().ViewUnauthorized] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsDirectoryGroupsRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsDirectoryGroupsRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsDirectoryGroupsRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsDirectoryGroupsRepository().LastUpdatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsDirectoryGroupsRepository().LastUpdatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsDirectoryGroupsRepository().LastUpdatedOn] = value
		}
	}

	directoryGroupsIDJoinDirectoryGroups := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)
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
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryGroupsRepository().ID, true),                                      //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, false), //2
			)

			selectQuery.Join[directoryGroupsIDJoinDirectoryGroups] = sq
		}
	}

	metadataModelsIDJoinMetadataModels := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsDirectoryGroupsRepository().MetadataModelsID, intdoment.MetadataModelsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, metadataModelsIDJoinMetadataModels); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", metadataModelsIDJoinMetadataModels, err))
	} else {
		if sq, err := n.MetadataModelsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", metadataModelsIDJoinMetadataModels, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.MetadataModelsRepository().ID, true),                                      //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsDirectoryGroupsRepository().MetadataModelsID, false), //2
			)

			selectQuery.Join[metadataModelsIDJoinMetadataModels] = sq
		}
	}

	abstractionsDirectoryGroupsJoinAbstractionsDirectoryGroupsAuthorizationIDs := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName, intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, abstractionsDirectoryGroupsJoinAbstractionsDirectoryGroupsAuthorizationIDs); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", abstractionsDirectoryGroupsJoinAbstractionsDirectoryGroupsAuthorizationIDs, err))
	} else {
		if sq, err := n.AuthorizationIDsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
			intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().RepositoryName,
			[]AuthIDsSelectQueryPKey{{Name: intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().DirectoryGroupsID, ProcessAs: PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE}},
			intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().CreationIamGroupAuthorizationsID,
			intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", abstractionsDirectoryGroupsJoinAbstractionsDirectoryGroupsAuthorizationIDs, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.AbstractionsDirectoryGroupsAuthorizationIDsRepository().DirectoryGroupsID, true), //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, false),       //2
			)

			selectQuery.Join[abstractionsDirectoryGroupsJoinAbstractionsDirectoryGroupsAuthorizationIDs] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}
