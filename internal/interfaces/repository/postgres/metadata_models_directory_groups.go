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

func (n *PostrgresRepository) RepoMetadataModelsDirectoryGroupsDeleteOne(ctx context.Context, datum *intdoment.MetadataModelsDirectoryGroups) error {
	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s= $1;",
		intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName,    //1
		intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID, //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryGroupsDeleteOne))
	if _, err := n.db.Exec(ctx, query, datum.DirectoryGroupsID[0]); err != nil {
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoMetadataModelsDirectoryGroupsUpdateOne(ctx context.Context, datum *intdoment.MetadataModelsDirectoryGroups) error {
	valuesToUpdate := make([]any, 0)
	columnsToUpdate := make([]string, 0)
	if v, c, err := n.RepoMetadataModelsDirectoryGroupsValidateAndGetColumnsAndData(datum, true); err != nil {
		return err
	} else if len(c) == 0 || len(v) == 0 {
		return intlib.NewError(http.StatusBadRequest, "no values to update")
	} else {
		valuesToUpdate = append(valuesToUpdate, v...)
		columnsToUpdate = append(columnsToUpdate, c...)
	}

	nextPlaceholder := 1
	query := fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s WHERE %[3]s = %[4]s;",
		intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName,    //1
		GetUpdateSetColumns(columnsToUpdate, &nextPlaceholder),                //2
		intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID, //3
		GetandUpdateNextPlaceholder(&nextPlaceholder),                         //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryGroupsUpdateOne))
	valuesToUpdate = append(valuesToUpdate, datum.DirectoryGroupsID[0])
	if _, err := n.db.Exec(ctx, query, valuesToUpdate...); err != nil {
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsUpdateOne, fmt.Errorf("update %s failed, err: %v", intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoMetadataModelsDirectoryGroupsInsertOne(ctx context.Context, datum *intdoment.MetadataModelsDirectoryGroups, columns []string) (*intdoment.MetadataModelsDirectoryGroups, error) {
	metadataModelsDgMModel, err := intlib.MetadataModelGet(intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(metadataModelsDgMModel, intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID) {
		columns = append(columns, intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID)
	}

	valuesToInsert := make([]any, 0)
	columnsToInsert := make([]string, 0)
	if v, c, err := n.RepoMetadataModelsDirectoryGroupsValidateAndGetColumnsAndData(datum, true); err != nil {
		return nil, err
	} else if len(c) == 0 || len(v) == 0 {
		return nil, intlib.NewError(http.StatusBadRequest, "no values to insert")
	} else {
		valuesToInsert = append(valuesToInsert, v...)
		columnsToInsert = append(columnsToInsert, c...)
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s) VALUES (%[3]s) RETURNING %[4]s;",
		intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, //1
		strings.Join(columnsToInsert, " , "),                               //2
		GetQueryPlaceholderString(len(valuesToInsert), &[]int{1}[0]),       //3
		strings.Join(columns, " , "),                                       //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryGroupsInsertOne))

	rows, err := n.db.Query(ctx, query, valuesToInsert...)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(metadataModelsDgMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryGroupsInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsInsertOne, fmt.Errorf("more than one %s found", intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName))
	}

	metadataModelDg := new(intdoment.MetadataModelsDirectoryGroups)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing metadataModelDg", "metadataModelDg", string(jsonData), "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryGroupsInsertOne))
		if err := json.Unmarshal(jsonData, metadataModelDg); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsInsertOne, err)
		}
	}

	return metadataModelDg, nil
}

func (n *PostrgresRepository) RepoMetadataModelsDirectoryGroupsValidateAndGetColumnsAndData(mm *intdoment.MetadataModelsDirectoryGroups, insert bool) ([]any, []string, error) {
	values := make([]any, 0)
	columns := make([]string, 0)

	if insert {
		if len(mm.DirectoryGroupsID) > 0 {
			values = append(values, mm.DirectoryGroupsID[0])
			columns = append(columns, intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID)
		} else {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID)
		}
	}

	if len(mm.MetadataModelsID) == 0 || len(mm.MetadataModelsID[0]) < 4 {
		if insert {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.MetadataModelsDirectoryGroupsRepository().MetadataModelsID)
		}
	} else {
		values = append(values, mm.MetadataModelsID[0])
		columns = append(columns, intdoment.MetadataModelsDirectoryGroupsRepository().MetadataModelsID)
	}

	return values, columns, nil
}

func (n *PostrgresRepository) RepoMetadataModelsDirectoryGroupsSearch(
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
	selectQuery, err := pSelectQuery.MetadataModelsDirectoryGroupsGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryGroupsSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryGroupsSearch, err)
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

func (n *PostgresSelectQuery) MetadataModelsDirectoryGroupsGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        "",
				RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS_DIRECTORY_GROUPS,
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
		TableName: intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName,
		Query:     "",
		Where:     make(map[string]map[int][][]string),
		Join:      make(map[string]*SelectQuery),
		JoinQuery: make([]string, 0),
	}

	if tableUid, ok := metadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID].(string); ok && len(tableUid) > 0 {
		selectQuery.TableUid = tableUid
	} else {
		return nil, intlib.FunctionNameAndError(n.MetadataModelsDirectoryGroupsGetSelectQuery, errors.New("tableUid is empty"))
	}

	if value, err := intlibmmodel.DatabaseGetColumnFields(metadataModel, selectQuery.TableUid, false, false); err != nil {
		return nil, intlib.FunctionNameAndError(n.MetadataModelsDirectoryGroupsGetSelectQuery, fmt.Errorf("extract database column fields failed, error: %v", err))
	} else {
		selectQuery.Columns = value
	}

	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsDirectoryGroupsRepository().MetadataModelsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsDirectoryGroupsRepository().MetadataModelsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsDirectoryGroupsRepository().MetadataModelsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsDirectoryGroupsRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsDirectoryGroupsRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsDirectoryGroupsRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsDirectoryGroupsRepository().LastUpdatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsDirectoryGroupsRepository().LastUpdatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsDirectoryGroupsRepository().LastUpdatedOn] = value
		}
	}

	directoryGroupsIDJoinDirectoryGroups := intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)
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
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryGroupsRepository().ID, true),                                        //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID, false), //2
			)

			selectQuery.Join[directoryGroupsIDJoinDirectoryGroups] = sq
		}
	}

	metadataModelsIDJoinMetadataModels := intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsDirectoryGroupsRepository().MetadataModelsID, intdoment.MetadataModelsRepository().RepositoryName)
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
				GetJoinColumnName(sq.TableUid, intdoment.MetadataModelsRepository().ID, true),                                        //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.MetadataModelsDirectoryGroupsRepository().MetadataModelsID, false), //2
			)

			selectQuery.Join[metadataModelsIDJoinMetadataModels] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}
