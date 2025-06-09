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

func (n *PostrgresRepository) RepoIamGroupAuthorizationsDeleteOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.IamGroupAuthorizations) error {
	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsDeleteOne, fmt.Errorf("start transaction to delete %s failed, error: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s = $1;",
		intdoment.IamGroupAuthorizationsIDsRepository().RepositoryName, //1
		intdoment.IamGroupAuthorizationsIDsRepository().ID,             //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsDeleteOne))

	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		query := fmt.Sprintf(
			"DELETE FROM %[1]s WHERE %[2]s = $1;",
			intdoment.IamGroupAuthorizationsRepository().RepositoryName, //1
			intdoment.IamGroupAuthorizationsRepository().ID,             //2
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsDeleteOne))
		if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
			if err := transaction.Commit(ctx); err != nil {
				return intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsDeleteOne, fmt.Errorf("commit transaction to delete %s failed, error: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
			}
			return nil
		} else {
			transaction.Rollback(ctx)
		}
	}

	transaction, err = n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsDeleteOne, fmt.Errorf("start transaction to deactivate %s failed, error: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s = $2;",
		intdoment.IamGroupAuthorizationsIDsRepository().RepositoryName,                       //1
		intdoment.IamGroupAuthorizationsIDsRepository().DeactivationIamGroupAuthorizationsID, //2
		intdoment.IamGroupAuthorizationsIDsRepository().ID,                                   //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsDeleteOne))
	if _, err := transaction.Exec(ctx, query, iamAuthRule.ID, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.IamGroupAuthorizationsIDsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = NOW() WHERE %[3]s = $1;",
		intdoment.IamGroupAuthorizationsRepository().RepositoryName, //1
		intdoment.IamGroupAuthorizationsRepository().DeactivatedOn,  //2
		intdoment.IamGroupAuthorizationsRepository().ID,             //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsDeleteOne, fmt.Errorf("commit transaction to update deactivation of %s failed, error: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoIamGroupAuthorizationsFindOneInactiveRule(ctx context.Context, iamGroupAuthorizationID uuid.UUID, columns []string) (*intdoment.IamGroupAuthorizations, error) {
	iamGroupAuthorizationsMModel, err := intlib.MetadataModelGet(intdoment.IamGroupAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneInactiveRule, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(iamGroupAuthorizationsMModel, intdoment.IamGroupAuthorizationsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneInactiveRule, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.IamGroupAuthorizationsRepository().ID) {
		columns = append(columns, intdoment.IamGroupAuthorizationsRepository().ID)
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s IS NOT NULL;",
		strings.Join(columns, "  , "),                               //1
		intdoment.IamGroupAuthorizationsRepository().RepositoryName, //2
		intdoment.IamGroupAuthorizationsRepository().ID,             //3
		intdoment.IamGroupAuthorizationsRepository().DeactivatedOn,  //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsFindOneInactiveRule))

	rows, err := n.db.Query(ctx, query, iamGroupAuthorizationID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneInactiveRule, fmt.Errorf("retrieve %s failed, err: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneInactiveRule, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(iamGroupAuthorizationsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneInactiveRule, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneInactiveRule, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsFindOneInactiveRule))
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneInactiveRule, fmt.Errorf("more than one %s found", intdoment.IamGroupAuthorizationsRepository().RepositoryName))
	}

	iamGroupAuthorization := new(intdoment.IamGroupAuthorizations)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneInactiveRule, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing iamGroupAuthorization", "iamGroupAuthorization", string(jsonData), "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsFindOneInactiveRule))
		if err := json.Unmarshal(jsonData, iamGroupAuthorization); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneInactiveRule, err)
		}
	}

	return iamGroupAuthorization, nil
}

func (n *PostrgresRepository) RepoIamGroupAuthorizationsInsertOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.IamGroupAuthorizations, columns []string) (*intdoment.IamGroupAuthorizations, error) {
	iamGroupAuthorizationsMModel, err := intlib.MetadataModelGet(intdoment.IamGroupAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(iamGroupAuthorizationsMModel, intdoment.IamGroupAuthorizationsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.IamGroupAuthorizationsRepository().ID) {
		columns = append(columns, intdoment.IamGroupAuthorizationsRepository().ID)
	}

	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, fmt.Errorf("start transaction to create %s failed, error: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2) RETURNING %[4]s;",
		intdoment.IamGroupAuthorizationsRepository().RepositoryName,            //1
		intdoment.IamGroupAuthorizationsRepository().IamCredentialsID,          //2
		intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID, //3
		strings.Join(columns, "  , "),                                          //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsInsertOne))

	rows, err := transaction.Query(ctx, query, datum.IamCredentialsID[0], datum.GroupRuleAuthorizationsID[0])
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(iamGroupAuthorizationsMModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		transaction.Rollback(ctx)
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.IamGroupAuthorizationsRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		transaction.Rollback(ctx)
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, fmt.Errorf("more than one %s found", intdoment.IamGroupAuthorizationsRepository().RepositoryName))
	}

	iamGroupAuthorization := new(intdoment.IamGroupAuthorizations)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing iamGroupAuthorization", "iamGroupAuthorization", string(jsonData), "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsInsertOne))
		if err := json.Unmarshal(jsonData, iamGroupAuthorization); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, err)
		}
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.IamGroupAuthorizationsIDsRepository().RepositoryName,                   //1
		intdoment.IamGroupAuthorizationsIDsRepository().ID,                               //2
		intdoment.IamGroupAuthorizationsIDsRepository().CreationIamGroupAuthorizationsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsInsertOne))

	if _, err := transaction.Exec(ctx, query, iamGroupAuthorization.ID[0], iamAuthRule.ID); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.IamGroupAuthorizationsIDsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsInsertOne, fmt.Errorf("commit transaction to create %s failed, error: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
	}

	return iamGroupAuthorization, nil
}

func (n *PostrgresRepository) RepoIamGroupAuthorizationsFindOneActiveRule(ctx context.Context, iamCredentialID uuid.UUID, groupRuleAuthorizationID uuid.UUID, columns []string) (*intdoment.IamGroupAuthorizations, error) {
	iamGroupAuthorizationsMModel, err := intlib.MetadataModelGet(intdoment.IamGroupAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneActiveRule, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(iamGroupAuthorizationsMModel, intdoment.IamGroupAuthorizationsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneActiveRule, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.IamGroupAuthorizationsRepository().ID) {
		columns = append(columns, intdoment.IamGroupAuthorizationsRepository().ID)
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s = $2 AND %[5]s IS NULL;",
		strings.Join(columns, "  , "),                                          //1
		intdoment.IamGroupAuthorizationsRepository().RepositoryName,            //2
		intdoment.IamGroupAuthorizationsRepository().IamCredentialsID,          //3
		intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID, //4
		intdoment.IamGroupAuthorizationsRepository().DeactivatedOn,             //5
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsFindOneActiveRule))

	rows, err := n.db.Query(ctx, query, iamCredentialID, groupRuleAuthorizationID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneActiveRule, fmt.Errorf("retrieve %s failed, err: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneActiveRule, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(iamGroupAuthorizationsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneActiveRule, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneActiveRule, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsFindOneActiveRule))
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneActiveRule, fmt.Errorf("more than one %s found", intdoment.IamGroupAuthorizationsRepository().RepositoryName))
	}

	iamGroupAuthorization := new(intdoment.IamGroupAuthorizations)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneActiveRule, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing iamGroupAuthorization", "iamGroupAuthorization", string(jsonData), "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsFindOneActiveRule))
		if err := json.Unmarshal(jsonData, iamGroupAuthorization); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsFindOneActiveRule, err)
		}
	}

	return iamGroupAuthorization, nil
}

func (n *PostrgresRepository) RepoIamGroupAuthorizationsSearch(
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
	selectQuery, err := pSelectQuery.IamGroupAuthorizationsGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.IamCredentialsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsSearch, err)
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

func (n *PostgresSelectQuery) IamGroupAuthorizationsGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_IAM_GROUP_AUTHORIZATIONS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_IAM_GROUP_AUTHORIZATIONS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_IAM_GROUP_AUTHORIZATIONS,
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
		TableName: intdoment.IamGroupAuthorizationsRepository().RepositoryName,
		Query:     "",
		Where:     make(map[string]map[int][][]string),
		Join:      make(map[string]*SelectQuery),
		JoinQuery: make([]string, 0),
	}

	if tableUid, ok := metadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID].(string); ok && len(tableUid) > 0 {
		selectQuery.TableUid = tableUid
	} else {
		return nil, intlib.FunctionNameAndError(n.IamGroupAuthorizationsGetSelectQuery, errors.New("tableUid is empty"))
	}

	if value, err := intlibmmodel.DatabaseGetColumnFields(metadataModel, selectQuery.TableUid, false, false); err != nil {
		return nil, intlib.FunctionNameAndError(n.DirectoryGroupsGetSelectQuery, fmt.Errorf("extract database column fields failed, error: %v", err))
	} else {
		selectQuery.Columns = value
	}

	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.IamGroupAuthorizationsRepository().ID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamGroupAuthorizationsRepository().ID, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamGroupAuthorizationsRepository().ID] = value
		}
	}

	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.IamGroupAuthorizationsRepository().IamCredentialsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamGroupAuthorizationsRepository().ID, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamGroupAuthorizationsRepository().IamCredentialsID] = value
		}
	}

	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamGroupAuthorizationsRepository().ID, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID] = value
		}
	}

	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.IamGroupAuthorizationsRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamGroupAuthorizationsRepository().ID, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamGroupAuthorizationsRepository().CreatedOn] = value
		}
	}

	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.IamGroupAuthorizationsRepository().DeactivatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamGroupAuthorizationsRepository().ID, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamGroupAuthorizationsRepository().DeactivatedOn] = value
		}
	}

	iamCredentialsIDJoinIamCredentials := intlib.MetadataModelGenJoinKey(intdoment.IamGroupAuthorizationsRepository().IamCredentialsID, intdoment.IamCredentialsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, iamCredentialsIDJoinIamCredentials); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", iamCredentialsIDJoinIamCredentials, err))
	} else {
		if sq, err := n.IamCredentialsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", iamCredentialsIDJoinIamCredentials, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.IamCredentialsRepository().ID, true),                                 //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.IamGroupAuthorizationsRepository().IamCredentialsID, false), //2
			)

			selectQuery.Join[iamCredentialsIDJoinIamCredentials] = sq
		}
	}

	groupRuleAuthorizationsIDJoinGroupRuleAuthorizations := intlib.MetadataModelGenJoinKey(intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID, intdoment.GroupRuleAuthorizationsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, groupRuleAuthorizationsIDJoinGroupRuleAuthorizations); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", groupRuleAuthorizationsIDJoinGroupRuleAuthorizations, err))
	} else {
		if sq, err := n.GroupRuleAuthorizationsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", groupRuleAuthorizationsIDJoinGroupRuleAuthorizations, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.GroupRuleAuthorizationsRepository().ID, true),                                 //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID, false), //2
			)

			selectQuery.Join[groupRuleAuthorizationsIDJoinGroupRuleAuthorizations] = sq
		}
	}

	iamGroupAuthorizationsJoinIamGroupAuthorizationIDs := intlib.MetadataModelGenJoinKey(intdoment.IamGroupAuthorizationsRepository().RepositoryName, intdoment.IamGroupAuthorizationsIDsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, iamGroupAuthorizationsJoinIamGroupAuthorizationIDs); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", iamGroupAuthorizationsJoinIamGroupAuthorizationIDs, err))
	} else {
		if sq, err := n.AuthorizationIDsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
			intdoment.IamGroupAuthorizationsIDsRepository().RepositoryName,
			[]AuthIDsSelectQueryPKey{{Name: intdoment.IamGroupAuthorizationsIDsRepository().ID, ProcessAs: PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE}},
			intdoment.IamGroupAuthorizationsIDsRepository().CreationIamGroupAuthorizationsID,
			intdoment.IamGroupAuthorizationsIDsRepository().DeactivationIamGroupAuthorizationsID,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", iamGroupAuthorizationsJoinIamGroupAuthorizationIDs, err))
		} else {
			if len(sq.Where) == 0 {
				sq.JoinType = JOIN_LEFT
			} else {
				sq.JoinType = JOIN_INNER
			}
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.IamGroupAuthorizationsIDsRepository().ID, true),        //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.IamGroupAuthorizationsRepository().ID, false), //2
			)

			selectQuery.Join[iamGroupAuthorizationsJoinIamGroupAuthorizationIDs] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}

func (n *PostrgresRepository) RepoIamGroupAuthorizationsGetAuthorized(ctx context.Context, iamAuthInfo *intdoment.IamCredentials, authContextDirectoryGroupID uuid.UUID, groupAuthorizationRules []*intdoment.IamGroupAuthorizationRule, currentIamAuthorizationRules *intdoment.IamAuthorizationRules) ([]*intdoment.IamAuthorizationRule, error) {
	if iamAuthInfo == nil {
		return nil, nil
	}

	selectColumns := []string{
		fmt.Sprintf("%s.%s", intdoment.IamGroupAuthorizationsRepository().RepositoryName, intdoment.IamGroupAuthorizationsRepository().ID),                             //1
		fmt.Sprintf("%s.%s", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID),            //2
		fmt.Sprintf("%s.%s", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID),    //3
		fmt.Sprintf("%s.%s", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup), //4
	}

	nextPlaceholder := 1
	whereOrConditions := make([]string, 0)
	valuesForCondition := make([]any, 0)
	cacheIamAuthRules := make([]*intdoment.IamAuthorizationRule, 0)
	for _, gar := range groupAuthorizationRules {
		if currentIamAuthorizationRules != nil {
			if iamGroupAuthorizationID, ok := (*currentIamAuthorizationRules)[genIamGroupAuthorizationIDsKey(
				authContextDirectoryGroupID,
				gar.ID,
				gar.RuleGroup,
			)]; ok {
				cacheIamAuthRules = append(cacheIamAuthRules, iamGroupAuthorizationID)
				continue
			}
		}
		whereAndCondition := make([]string, 0)
		whereAndCondition = append(whereAndCondition, fmt.Sprintf("%s.%s = $%d", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID, nextPlaceholder))
		valuesForCondition = append(valuesForCondition, authContextDirectoryGroupID)
		nextPlaceholder += 1
		if gar.ID != "*" && len(gar.ID) > 0 {
			whereAndCondition = append(whereAndCondition, fmt.Sprintf("%s.%s = $%d", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID, nextPlaceholder))
			valuesForCondition = append(valuesForCondition, gar.ID)
			nextPlaceholder += 1
		}
		whereAndCondition = append(whereAndCondition, fmt.Sprintf("%s.%s = $%d", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup, nextPlaceholder))
		valuesForCondition = append(valuesForCondition, gar.RuleGroup)
		nextPlaceholder += 1
		whereOrConditions = append(whereOrConditions, strings.Join(whereAndCondition, " AND "))
	}

	if len(cacheIamAuthRules) > 0 {
		return cacheIamAuthRules, nil
	}

	if len(whereOrConditions) == 0 || len(valuesForCondition) == 0 {
		return nil, errors.New("groupAuthorizationRules invalid")
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s INNER JOIN %[3]s ON %[2]s.%[4]s = %[3]s.%[5]s WHERE %[6]s AND %[2]s.%[7]s IS NULL AND %[3]s.%[8]s IS NULL AND %[2]s.%[9]s = %[10]s;",
		strings.Join(selectColumns, " , "),                                     //1
		intdoment.IamGroupAuthorizationsRepository().RepositoryName,            //2
		intdoment.GroupRuleAuthorizationsRepository().RepositoryName,           //3
		intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID, //4
		intdoment.GroupRuleAuthorizationsRepository().ID,                       //5
		strings.Join(whereOrConditions, " OR "),                                //6
		intdoment.IamGroupAuthorizationsRepository().DeactivatedOn,             //7
		intdoment.GroupRuleAuthorizationsRepository().DeactivatedOn,            //8
		intdoment.IamGroupAuthorizationsRepository().IamCredentialsID,          //9
		GetandUpdateNextPlaceholder(&nextPlaceholder),                          //10
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsGetAuthorized))
	valuesForCondition = append(valuesForCondition, iamAuthInfo.ID[0])

	rows, err := n.db.Query(ctx, query, valuesForCondition...)
	if err != nil {
		errmsg := fmt.Errorf("get %s failed, error: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err)
		n.logger.Log(ctx, slog.LevelDebug, errmsg.Error())
		return nil, errmsg
	}
	newIamAuthorizationRules := make([]*intdoment.IamAuthorizationRule, 0)
	for rows.Next() {
		newIamAuthorizationRule := new(intdoment.IamAuthorizationRule)
		if err := rows.Scan(&newIamAuthorizationRule.ID, &newIamAuthorizationRule.DirectoryGroupID, &newIamAuthorizationRule.GroupAuthorizationRuleID, &newIamAuthorizationRule.GroupAuthorizationRuleGroup); err != nil {
			errmsg := fmt.Errorf("read %s row failed, error: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err)
			n.logger.Log(ctx, slog.LevelDebug, errmsg.Error())
			return nil, errmsg
		}
		newIamAuthorizationRules = append(newIamAuthorizationRules, newIamAuthorizationRule)
		if currentIamAuthorizationRules != nil {
			(*currentIamAuthorizationRules)[genIamGroupAuthorizationIDsKey(
				newIamAuthorizationRule.DirectoryGroupID,
				newIamAuthorizationRule.GroupAuthorizationRuleGroup,
				newIamAuthorizationRule.GroupAuthorizationRuleGroup,
			)] = newIamAuthorizationRule
		}
	}

	if len(newIamAuthorizationRules) > 0 {
		return newIamAuthorizationRules, nil
	} else {
		return nil, nil
	}
}

func genIamGroupAuthorizationIDsKey(groupRuleAuthorizationID uuid.UUID, groupRuleAuthorizationRuleID string, groupRuleAuthorizationRuleGroup string) string {
	return fmt.Sprintf("%s/%s/%s",
		groupRuleAuthorizationID,
		groupRuleAuthorizationRuleID,
		groupRuleAuthorizationRuleGroup,
	)
}

func (n *PostrgresRepository) RepoIamGroupAuthorizationsSystemRolesInsertMany(ctx context.Context, iamCredenial *intdoment.IamCredentials, groupRuleAuthorizations []intdoment.GroupRuleAuthorizations) (int, error) {
	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.IamGroupAuthorizationsRepository().RepositoryName,            //1
		intdoment.IamGroupAuthorizationsRepository().IamCredentialsID,          //2
		intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamGroupAuthorizationsSystemRolesInsertMany))

	successfulInserts := 0
	for _, gra := range groupRuleAuthorizations {
		if _, err := n.db.Exec(ctx, query, iamCredenial.ID[0], gra.ID[0]); err != nil {
			return successfulInserts, intlib.FunctionNameAndError(n.RepoIamGroupAuthorizationsSystemRolesInsertMany, fmt.Errorf("insert %v failed, error: %v", intdoment.IamGroupAuthorizationsRepository().RepositoryName, err))
		}
		successfulInserts += 1
	}

	return successfulInserts, nil
}
