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

func (n *PostrgresRepository) RepoMetadataModelsDirectoryDeleteOne(ctx context.Context, datum *intdoment.MetadataModelsDirectory) error {
	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s= $1;",
		intdoment.MetadataModelsDirectoryRepository().RepositoryName,    //1
		intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID, //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryDeleteOne))
	if _, err := n.db.Exec(ctx, query, datum.DirectoryGroupsID[0]); err != nil {
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.MetadataModelsDirectoryRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoMetadataModelsDirectoryUpdateOne(ctx context.Context, datum *intdoment.MetadataModelsDirectory) error {
	valuesToUpdate := make([]any, 0)
	columnsToUpdate := make([]string, 0)
	if v, c, err := n.RepoMetadataModelsDirectoryValidateAndGetColumnsAndData(datum, true); err != nil {
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
		intdoment.MetadataModelsDirectoryRepository().RepositoryName,    //1
		GetUpdateSetColumns(columnsToUpdate, &nextPlaceholder),          //2
		intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID, //3
		GetandUpdateNextPlaceholder(&nextPlaceholder),                   //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryUpdateOne))
	valuesToUpdate = append(valuesToUpdate, datum.DirectoryGroupsID[0])
	if _, err := n.db.Exec(ctx, query, valuesToUpdate...); err != nil {
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryUpdateOne, fmt.Errorf("update %s failed, err: %v", intdoment.MetadataModelsDirectoryRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoMetadataModelsDirectoryInsertOne(ctx context.Context, datum *intdoment.MetadataModelsDirectory, columns []string) (*intdoment.MetadataModelsDirectory, error) {
	metadataModelsDgMModel, err := intlib.MetadataModelGet(intdoment.MetadataModelsDirectoryRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(metadataModelsDgMModel, intdoment.MetadataModelsDirectoryRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID) {
		columns = append(columns, intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID)
	}

	valuesToInsert := make([]any, 0)
	columnsToInsert := make([]string, 0)
	if v, c, err := n.RepoMetadataModelsDirectoryValidateAndGetColumnsAndData(datum, true); err != nil {
		return nil, err
	} else if len(c) == 0 || len(v) == 0 {
		return nil, intlib.NewError(http.StatusBadRequest, "no values to insert")
	} else {
		valuesToInsert = append(valuesToInsert, v...)
		columnsToInsert = append(columnsToInsert, c...)
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s) VALUES (%[3]s) RETURNING %[4]s;",
		intdoment.MetadataModelsDirectoryRepository().RepositoryName, //1
		strings.Join(columnsToInsert, " , "),                         //2
		GetQueryPlaceholderString(len(valuesToInsert), &[]int{1}[0]), //3
		strings.Join(columns, " , "),                                 //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryInsertOne))

	rows, err := n.db.Query(ctx, query, valuesToInsert...)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.MetadataModelsDirectoryRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(metadataModelsDgMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.MetadataModelsDirectoryRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryInsertOne, fmt.Errorf("more than one %s found", intdoment.MetadataModelsDirectoryRepository().RepositoryName))
	}

	metadataModelDg := new(intdoment.MetadataModelsDirectory)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing metadataModelDg", "metadataModelDg", string(jsonData), "function", intlib.FunctionName(n.RepoMetadataModelsDirectoryInsertOne))
		if err := json.Unmarshal(jsonData, metadataModelDg); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectoryInsertOne, err)
		}
	}

	return metadataModelDg, nil
}

func (n *PostrgresRepository) RepoMetadataModelsDirectoryValidateAndGetColumnsAndData(mm *intdoment.MetadataModelsDirectory, insert bool) ([]any, []string, error) {
	values := make([]any, 0)
	columns := make([]string, 0)

	if insert {
		if len(mm.DirectoryGroupsID) > 0 {
			values = append(values, mm.DirectoryGroupsID[0])
			columns = append(columns, intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID)
		} else {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID)
		}
	}

	if len(mm.MetadataModelsID) == 0 || len(mm.MetadataModelsID[0]) < 4 {
		if insert {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.MetadataModelsDirectoryRepository().MetadataModelsID)
		}
	} else {
		values = append(values, mm.MetadataModelsID[0])
		columns = append(columns, intdoment.MetadataModelsDirectoryRepository().MetadataModelsID)
	}

	return values, columns, nil
}

func (n *PostrgresRepository) RepoMetadataModelsDirectorySearch(
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
	selectQuery, err := pSelectQuery.MetadataModelsDirectoryGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectorySearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDirectorySearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectorySearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.MetadataModelsDirectoryRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectorySearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectorySearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsDirectorySearch, err)
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

func (n *PostgresSelectQuery) MetadataModelsDirectoryGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
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
		TableName: intdoment.MetadataModelsDirectoryRepository().RepositoryName,
		Query:     "",
		Where:     make(map[string]map[int][][]string),
		Join:      make(map[string]*SelectQuery),
		JoinQuery: make([]string, 0),
	}

	if tableUid, ok := metadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID].(string); ok && len(tableUid) > 0 {
		selectQuery.TableUid = tableUid
	} else {
		return nil, intlib.FunctionNameAndError(n.MetadataModelsDirectoryGetSelectQuery, errors.New("tableUid is empty"))
	}

	if value, err := intlibmmodel.DatabaseGetColumnFields(metadataModel, selectQuery.TableUid, false, false); err != nil {
		return nil, intlib.FunctionNameAndError(n.MetadataModelsDirectoryGetSelectQuery, fmt.Errorf("extract database column fields failed, error: %v", err))
	} else {
		selectQuery.Columns = value
	}

	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsDirectoryRepository().MetadataModelsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsDirectoryRepository().MetadataModelsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsDirectoryRepository().MetadataModelsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsDirectoryRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsDirectoryRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsDirectoryRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsDirectoryRepository().LastUpdatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsDirectoryRepository().LastUpdatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsDirectoryRepository().LastUpdatedOn] = value
		}
	}

	directoryGroupsIDJoinDirectoryGroups := intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)
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
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryGroupsRepository().ID, true),                                  //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID, false), //2
			)

			selectQuery.Join[directoryGroupsIDJoinDirectoryGroups] = sq
		}
	}

	metadataModelsIDJoinMetadataModels := intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsDirectoryRepository().MetadataModelsID, intdoment.MetadataModelsRepository().RepositoryName)
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
				GetJoinColumnName(sq.TableUid, intdoment.MetadataModelsRepository().ID, true),                                  //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.MetadataModelsDirectoryRepository().MetadataModelsID, false), //2
			)

			selectQuery.Join[metadataModelsIDJoinMetadataModels] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}
