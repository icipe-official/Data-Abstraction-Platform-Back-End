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

func (n *PostrgresRepository) RepoAbstractionsReviewsCommentsInsertOne(
	ctx context.Context,
	datum *intdoment.AbstractionsReviewsComments,
	columns []string,
) (*intdoment.AbstractionsReviewsComments, error) {
	abstractionsReviewsCommentsMetadataModel, err := intlib.MetadataModelGet(intdoment.AbstractionsReviewsCommentsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(abstractionsReviewsCommentsMetadataModel, intdoment.AbstractionsReviewsCommentsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID) {
		columns = append(columns, intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID)
	}

	if !slices.Contains(columns, intdoment.AbstractionsReviewsCommentsRepository().DirectoryID) {
		columns = append(columns, intdoment.AbstractionsReviewsCommentsRepository().DirectoryID)
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s , %[3]s , %[4]s) VALUES($1 , $2 , $3) RETURNING %[5]s;",
		intdoment.AbstractionsReviewsCommentsRepository().RepositoryName, //1
		intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID, //2
		intdoment.AbstractionsReviewsCommentsRepository().DirectoryID,    //3
		intdoment.AbstractionsReviewsCommentsRepository().Comment,        //4
		strings.Join(columns, " , "),                                     //5
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsReviewsCommentsInsertOne))

	rows, err := n.db.Query(ctx, query, datum.AbstractionsID[0], datum.DirectoryID[0], datum.Comment[0])
	if err != nil {

		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsInsertOne, fmt.Errorf("upsert %s failed, err: %v", intdoment.AbstractionsReviewsCommentsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(abstractionsReviewsCommentsMetadataModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.AbstractionsReviewsCommentsRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoAbstractionsReviewsCommentsInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsInsertOne, fmt.Errorf("more than one %s found", intdoment.AbstractionsReviewsCommentsRepository().RepositoryName))
	}

	abstractionsReviewsComments := new(intdoment.AbstractionsReviewsComments)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing abstractionsReviewsComments", "abstractionsReviewsComments", string(jsonData), "function", intlib.FunctionName(n.RepoAbstractionsReviewsCommentsInsertOne))
		if err := json.Unmarshal(jsonData, abstractionsReviewsComments); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsInsertOne, err)
		}
	}

	return abstractionsReviewsComments, nil
}

func (n *PostrgresRepository) RepoAbstractionsReviewsCommentsSearch(
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
	selectQuery, err := pSelectQuery.AbstractionsReviewsCommentsGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsReviewsCommentsSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.AbstractionsReviewsCommentsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsReviewsCommentsSearch, err)
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

func (n *PostgresSelectQuery) AbstractionsReviewsCommentsGetSelectQuery(ctx context.Context, metadataModel map[string]any, abstractionParentPath string) (*SelectQuery, error) {
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
		TableName: intdoment.AbstractionsReviewsCommentsRepository().RepositoryName,
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
				intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID, //1
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

	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsReviewsCommentsRepository().ID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsReviewsCommentsRepository().ID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsReviewsCommentsRepository().ID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsReviewsCommentsRepository().DirectoryID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsReviewsCommentsRepository().DirectoryID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsReviewsCommentsRepository().DirectoryID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsReviewsCommentsRepository().Comment][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsReviewsCommentsRepository().Comment, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsReviewsCommentsRepository().Comment] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsReviewsCommentsRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsReviewsCommentsRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsReviewsCommentsRepository().CreatedOn] = value
		}
	}

	abstractionsIDJoinAbstractions := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID, intdoment.AbstractionsRepository().RepositoryName)
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
				GetJoinColumnName(sq.TableUid, intdoment.AbstractionsRepository().ID, true),                                      //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID, false), //2
			)

			selectQuery.Join[abstractionsIDJoinAbstractions] = sq
		}
	}

	directoryIDJoinDirectory := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsReviewsCommentsRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)
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
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryRepository().ID, true),                                      //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsReviewsCommentsRepository().DirectoryID, false), //2
			)

			selectQuery.Join[directoryIDJoinDirectory] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}
