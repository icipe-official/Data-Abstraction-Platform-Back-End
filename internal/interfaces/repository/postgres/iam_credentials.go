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
)

func (n *PostrgresRepository) RepoIamCredentialsUpdateOne(ctx context.Context, datum *intdoment.IamCredentials) error {
	valuesToUpdate := make([]any, 0)
	columnsToUpdate := make([]string, 0)
	if v, c, err := n.RepoIamCredentialsValidateAndGetColumnsAndData(datum); err != nil {
		return err
	} else if len(c) == 0 || len(v) == 0 {
		return intlib.NewError(http.StatusBadRequest, "no values to update")
	} else {
		valuesToUpdate = append(valuesToUpdate, v...)
		columnsToUpdate = append(columnsToUpdate, c...)
	}

	nextPlaceholder := 1
	query := fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s WHERE %[3]s = %[4]s AND %[5]s IS NULL AND %[6]s IS NULL;",
		intdoment.IamCredentialsRepository().RepositoryName,    //1
		GetUpdateSetColumns(columnsToUpdate, &nextPlaceholder), //2
		intdoment.IamCredentialsRepository().ID,                //3
		GetandUpdateNextPlaceholder(&nextPlaceholder),          //4
		intdoment.IamCredentialsRepository().DeactivatedOn,     //5
		intdoment.IamCredentialsRepository().DirectoryID,       //6
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryGroupsUpdateOne))
	valuesToUpdate = append(valuesToUpdate, datum.ID[0])
	if _, err := n.db.Exec(ctx, query, valuesToUpdate...); err != nil {
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsUpdateOne, fmt.Errorf("update %s failed, err: %v", intdoment.IamCredentialsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoIamCredentialsValidateAndGetColumnsAndData(datum *intdoment.IamCredentials) ([]any, []string, error) {
	values := make([]any, 0)
	columns := make([]string, 0)

	if len(datum.DirectoryID) > 0 {
		values = append(values, datum.DirectoryID[0])
		columns = append(columns, intdoment.IamCredentialsRepository().DirectoryID)
	}

	return values, columns, nil
}

func (n *PostrgresRepository) RepoIamCredentialsSearch(
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
	selectQuery, err := pSelectQuery.IamCredentialsGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamCredentialsSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.IamCredentialsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsSearch, err)
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

func (n *PostgresSelectQuery) IamCredentialsGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_IAM_CREDENTIALS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_IAM_CREDENTIALS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_IAM_CREDENTIALS,
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
		TableName: intdoment.IamCredentialsRepository().RepositoryName,
		Query:     "",
		Where:     make(map[string]map[int][][]string),
		Join:      make(map[string]*SelectQuery),
		JoinQuery: make([]string, 0),
	}

	if tableUid, ok := metadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID].(string); ok && len(tableUid) > 0 {
		selectQuery.TableUid = tableUid
	} else {
		return nil, intlib.FunctionNameAndError(n.IamCredentialsGetSelectQuery, errors.New("tableUid is empty"))
	}

	if value, err := intlibmmodel.DatabaseGetColumnFields(metadataModel, selectQuery.TableUid, false, false); err != nil {
		return nil, intlib.FunctionNameAndError(n.IamCredentialsGetSelectQuery, fmt.Errorf("extract database column fields failed, error: %v", err))
	} else {
		selectQuery.Columns = value
	}

	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().ID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().ID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().ID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().DirectoryID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().DirectoryID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().DirectoryID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().OpenidSub][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().OpenidSub, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().OpenidSub] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().OpenidPreferredUsername][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().OpenidPreferredUsername, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().OpenidPreferredUsername] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().OpenidEmail][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().OpenidEmail, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().OpenidEmail] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().OpenidEmailVerified][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().OpenidEmailVerified, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().OpenidEmailVerified] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().OpenidGivenName][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().OpenidGivenName, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().OpenidGivenName] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().OpenidFamilyName][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().OpenidFamilyName, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().OpenidFamilyName] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().LastUpdatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().LastUpdatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().LastUpdatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.IamCredentialsRepository().DeactivatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.IamCredentialsRepository().DeactivatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.IamCredentialsRepository().DeactivatedOn] = value
		}
	}
	if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, selectQuery.TableName, "", "", "", intdoment.IamCredentialsRepository().FullTextSearch); len(value) > 0 {
		selectQuery.Where[intdoment.IamCredentialsRepository().RepositoryName] = value
	}

	directoryIDJoinDirectory := intlib.MetadataModelGenJoinKey(intdoment.IamCredentialsRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, directoryIDJoinDirectory); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", directoryIDJoinDirectory, err))
	} else {
		if sq, err := n.DirectoryGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", directoryIDJoinDirectory, err))
		} else {
			if len(sq.Where) == 0 {
				sq.JoinType = JOIN_LEFT
			} else {
				sq.JoinType = JOIN_INNER
			}
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryRepository().ID, true),                         //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.IamCredentialsRepository().DirectoryID, false), //2
			)

			selectQuery.Join[directoryIDJoinDirectory] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}

func (n *PostrgresRepository) RepoIamCredentialsInsertOpenIDUserInfo(ctx context.Context, openIDUserInfo *intdoment.OpenIDUserInfo, columns []string) (*intdoment.IamCredentials, error) {
	iamCredentialsMModel, err := intlib.MetadataModelGet(intdoment.IamCredentialsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(iamCredentialsMModel, intdoment.IamCredentialsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.IamCredentialsRepository().ID) {
		columns = append(columns, intdoment.IamCredentialsRepository().ID)
	}

	columnsToInsert := make([]string, 0)
	valuesToInsert := make([]any, 0)

	if openIDUserInfo.IsSubValid() {
		columnsToInsert = append(columnsToInsert, intdoment.IamCredentialsRepository().OpenidSub)
		valuesToInsert = append(valuesToInsert, openIDUserInfo.Sub)
	} else {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, errors.New("openIDUserInfo.Sub is empty"))
	}

	if openIDUserInfo.IsPreferredUsernameValid() {
		columnsToInsert = append(columnsToInsert, intdoment.IamCredentialsRepository().OpenidPreferredUsername)
		valuesToInsert = append(valuesToInsert, openIDUserInfo.PreferredUsername)
	} else {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, errors.New("openIDUserInfo.PreferredUsername is empty"))
	}

	if openIDUserInfo.IsEmailValid() {
		columnsToInsert = append(columnsToInsert, intdoment.IamCredentialsRepository().OpenidEmail)
		valuesToInsert = append(valuesToInsert, openIDUserInfo.Email)
	} else {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, errors.New("openIDUserInfo.Email is empty"))
	}

	columnsToInsert = append(columnsToInsert, intdoment.IamCredentialsRepository().OpenidEmailVerified)
	valuesToInsert = append(valuesToInsert, openIDUserInfo.EmailVerified)

	if openIDUserInfo.IsGivenNameValid() {
		columnsToInsert = append(columnsToInsert, intdoment.IamCredentialsRepository().OpenidGivenName)
		valuesToInsert = append(valuesToInsert, openIDUserInfo.GivenName)
	}

	if openIDUserInfo.IsFamilyNameValid() {
		columnsToInsert = append(columnsToInsert, intdoment.IamCredentialsRepository().OpenidFamilyName)
		valuesToInsert = append(valuesToInsert, openIDUserInfo.FamilyName)
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s) VALUES (%[3]s) RETURNING %[4]s;",
		intdoment.IamCredentialsRepository().RepositoryName,          //1
		strings.Join(columnsToInsert, " , "),                         //2
		GetQueryPlaceholderString(len(valuesToInsert), &[]int{1}[0]), //3
		intdoment.IamCredentialsRepository().ID,                      //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamCredentialsInsertOpenIDUserInfo))

	rows, err := n.db.Query(ctx, query, valuesToInsert...)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, fmt.Errorf("insert %s failed, err: %v", intdoment.IamCredentialsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(iamCredentialsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoIamCredentialsInsertOpenIDUserInfo))
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, errors.New("convert inserted rows return empty"))
	}

	iamCredential := new(intdoment.IamCredentials)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing iamCredential", "iamCredential", string(jsonData), "function", intlib.FunctionName(n.RepoIamCredentialsInsertOpenIDUserInfo))
		if err := json.Unmarshal(jsonData, iamCredential); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsInsertOpenIDUserInfo, err)
		}
	}

	return iamCredential, nil
}

func (n *PostrgresRepository) RepoIamCredentialsFindOneByID(ctx context.Context, column string, value any, columns []string) (*intdoment.IamCredentials, error) {
	iamCredentialsMModel, err := intlib.MetadataModelGet(intdoment.IamCredentialsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsFindOneByID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(iamCredentialsMModel, intdoment.IamCredentialsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsFindOneByID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.IamCredentialsRepository().ID) {
		columns = append(columns, intdoment.IamCredentialsRepository().ID)
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1;",
		strings.Join(columns, ","),                          //1
		intdoment.IamCredentialsRepository().RepositoryName, //2
		column, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoIamCredentialsFindOneByID))

	rows, err := n.db.Query(ctx, query, value)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsFindOneByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.IamCredentialsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsFindOneByID, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(iamCredentialsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsFindOneByID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsFindOneByID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoIamCredentialsFindOneByID))
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsFindOneByID, fmt.Errorf("more than one %s found", intdoment.IamCredentialsRepository().RepositoryName))
	}

	iamCredential := new(intdoment.IamCredentials)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsFindOneByID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing iamCredential", "iamCredential", string(jsonData), "function", intlib.FunctionName(n.RepoIamCredentialsFindOneByID))
		if err := json.Unmarshal(jsonData, iamCredential); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoIamCredentialsFindOneByID, err)
		}
	}

	return iamCredential, nil
}
