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

func (n *PostrgresRepository) RepoAbstractionsReviewsUpdateOne(
	ctx context.Context,
	datum *intdoment.AbstractionsReviews,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
) error {
	authWhere := ""
	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_CREATE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS_REVIEWS,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		whereQuery := fmt.Sprintf(
			"%[1]s IN (SELECT %[2]s FROM %[3]s WHERE %[4]s = '%[5]s' AND %[6]s = '%[7]s')",
			intdoment.AbstractionsReviewsRepository().AbstractionsID,         //1
			intdoment.AbstractionsRepository().ID,                            //2
			intdoment.AbstractionsRepository().RepositoryName,                //3
			intdoment.AbstractionsRepository().ID,                            //4
			datum.AbstractionsID[0].String(),                                 //5
			intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, //6
			authContextDirectoryGroupID.String(),                             //7
		)

		if len(authWhere) > 0 {
			authWhere += " OR " + whereQuery
		} else {
			authWhere = whereQuery
		}
	}

	if len(authWhere) == 0 {
		return intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}

	query := fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s = $2 AND %[4]s = $3 AND %[3]s IN (SELECT %[5]s.%[6]s FROM %[5]s INNER JOIN %[7]s ON %[5]s.%[8]s = %[7]s.%[9]s AND %[5]s.%[10]s IS NULL AND %[7]s.%[11]s IS NULL) AND (%[12]s);",
		intdoment.AbstractionsReviewsRepository().RepositoryName,            //1
		intdoment.AbstractionsReviewsRepository().ReviewPass,                //2
		intdoment.AbstractionsReviewsRepository().AbstractionsID,            //3
		intdoment.AbstractionsReviewsRepository().DirectoryID,               //4
		intdoment.AbstractionsRepository().RepositoryName,                   //5
		intdoment.AbstractionsRepository().ID,                               //6
		intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //7
		intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID,    //8
		intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //9
		intdoment.AbstractionsRepository().DeactivatedOn,                    //10
		intdoment.AbstractionsDirectoryGroupsRepository().DeactivatedOn,     //11
		authWhere, //12
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsReviewsUpdateOne))

	if _, err := n.db.Exec(ctx, query, datum.ReviewPass[0], datum.AbstractionsID[0], datum.DirectoryID[0]); err != nil {
		return intlib.FunctionNameAndError(n.RepoAbstractionsReviewsUpdateOne, fmt.Errorf("update %s failed, err: %v", intdoment.AbstractionsReviewsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoAbstractionsReviewsInsertOne(
	ctx context.Context,
	datum *intdoment.AbstractionsReviews,
	columns []string,
) (*intdoment.AbstractionsReviews, error) {
	abstractionsReviewsMetadataModel, err := intlib.MetadataModelGet(intdoment.AbstractionsReviewsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(abstractionsReviewsMetadataModel, intdoment.AbstractionsReviewsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.AbstractionsReviewsRepository().AbstractionsID) {
		columns = append(columns, intdoment.AbstractionsReviewsRepository().AbstractionsID)
	}

	if !slices.Contains(columns, intdoment.AbstractionsReviewsRepository().DirectoryID) {
		columns = append(columns, intdoment.AbstractionsReviewsRepository().DirectoryID)
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s , %[3]s , %[4]s) VALUES($1 , $2 , $3) RETURNING %[5]s;",
		intdoment.AbstractionsReviewsRepository().RepositoryName, //1
		intdoment.AbstractionsReviewsRepository().AbstractionsID, //2
		intdoment.AbstractionsReviewsRepository().DirectoryID,    //3
		intdoment.AbstractionsReviewsRepository().ReviewPass,     //4
		strings.Join(columns, " , "),                             //5
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsReviewsInsertOne))

	rows, err := n.db.Query(ctx, query, datum.AbstractionsID[0], datum.DirectoryID[0], datum.ReviewPass[0])
	if err != nil {

		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsInsertOne, fmt.Errorf("upsert %s failed, err: %v", intdoment.AbstractionsReviewsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(abstractionsReviewsMetadataModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.AbstractionsReviewsRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoAbstractionsReviewsInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsInsertOne, fmt.Errorf("more than one %s found", intdoment.AbstractionsReviewsRepository().RepositoryName))
	}

	abstractionsReviews := new(intdoment.AbstractionsReviews)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing abstractionsReviews", "abstractionsReviews", string(jsonData), "function", intlib.FunctionName(n.RepoAbstractionsReviewsInsertOne))
		if err := json.Unmarshal(jsonData, abstractionsReviews); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsInsertOne, err)
		}
	}

	return abstractionsReviews, nil
}

func (n *PostrgresRepository) RepoAbstractionsReviewsSearch(
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
	selectQuery, err := pSelectQuery.AbstractionsReviewsGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsReviewsSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.AbstractionsReviewsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsSearch, err)
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

func (n *PostgresSelectQuery) AbstractionsReviewsGetSelectQuery(ctx context.Context, metadataModel map[string]any, abstractionParentPath string) (*SelectQuery, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        "",
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS_REVIEWS,
			},
		},
		n.iamAuthorizationRules,
	); err != nil || iamAuthorizationRule == nil {
		return nil, intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}

	quoteColumns := true
	if len(abstractionParentPath) == 0 {
		abstractionParentPath = "$"
		quoteColumns = false
	}
	if !n.whereAfterJoin {
		quoteColumns = false
	}

	selectQuery := SelectQuery{
		TableName: intdoment.AbstractionsReviewsRepository().RepositoryName,
		Query:     "",
		Where:     make(map[string]map[int][][]string),
		WhereAnd:  make([]string, 0),
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

		selectQuery.DirectoryGroupsSubGroupsCTEName = cteName
		selectQuery.DirectoryGroupsSubGroupsCTE = RecursiveDirectoryGroupsSubGroupsCte(n.startSearchDirectoryGroupID, cteName)
		cteWhere = append(
			cteWhere,
			fmt.Sprintf(
				"(%[1]s) IN (SELECT %[2]s FROM %[3]s WHERE %[4]s IN (SELECT %[5]s FROM %[6]s) OR %[4]s = '%[7]s')",
				intdoment.AbstractionsReviewsRepository().AbstractionsID,         //1
				intdoment.AbstractionsRepository().ID,                            //2
				intdoment.AbstractionsRepository().RepositoryName,                //3
				intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, //4
				intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID,        //5
				cteName,                                //6
				n.startSearchDirectoryGroupID.String(), //7
			),
		)

		if len(cteWhere) > 0 {
			if len(cteWhere) > 1 {
				selectQuery.DirectoryGroupsSubGroupsCTECondition = fmt.Sprintf("(%s)", strings.Join(cteWhere, " OR "))
			} else {
				selectQuery.DirectoryGroupsSubGroupsCTECondition = cteWhere[0]
			}
		}

		n.startSearchDirectoryGroupID = uuid.Nil
	}

	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsReviewsRepository().AbstractionsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsReviewsRepository().AbstractionsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsReviewsRepository().AbstractionsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsReviewsRepository().DirectoryID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsReviewsRepository().DirectoryID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsReviewsRepository().DirectoryID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsReviewsRepository().ReviewPass][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsReviewsRepository().ReviewPass, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsReviewsRepository().ReviewPass] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsReviewsRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsReviewsRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsReviewsRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsReviewsRepository().LastUpdatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsReviewsRepository().LastUpdatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsReviewsRepository().LastUpdatedOn] = value
		}
	}

	abstractionsIDJoinAbstractions := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsReviewsRepository().AbstractionsID, intdoment.AbstractionsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, abstractionsIDJoinAbstractions); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", abstractionsIDJoinAbstractions, err))
	} else {
		if sq, err := n.AbstractionsGetSelectQuery(
			ctx,
			value,
			abstractionParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", abstractionsIDJoinAbstractions, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.AbstractionsRepository().ID, true),                              //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsReviewsRepository().AbstractionsID, false), //2
			)

			selectQuery.Join[abstractionsIDJoinAbstractions] = sq
		}
	}

	directoryIDJoinDirectory := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsReviewsRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, directoryIDJoinDirectory); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", directoryIDJoinDirectory, err))
	} else {
		if sq, err := n.DirectoryGetSelectQuery(
			ctx,
			value,
			abstractionParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", directoryIDJoinDirectory, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryRepository().ID, true),                              //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsReviewsRepository().DirectoryID, false), //2
			)

			selectQuery.Join[directoryIDJoinDirectory] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}
