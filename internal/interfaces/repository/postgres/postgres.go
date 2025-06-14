package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/brunoga/deep"
	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
	intlibjson "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib/json"
	intlibmmodel "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib/metadata_model"
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func MetadataModelExtractFullTextSearchValue(metadatamodel any, data any) []string {
	pathToValues := make([]string, 0)
	intlibmmodel.ForEachFieldGroup(metadatamodel, func(property map[string]any) bool {
		if value, ok := property[intlibmmodel.FIELD_GROUP_PROP_DATABASE_FIELD_ADD_DATA_TO_FULL_TEXT_SEARCH_INDEX].(bool); ok && value {
			if fgKey, ok := property[intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
				pathToValues = append(pathToValues, fmt.Sprintf("%s%s", intlibmmodel.GetPathToValue(fgKey, true, intlibmmodel.ARRAY_PATH_PLACEHOLDER), intlibmmodel.ARRAY_PATH_PLACEHOLDER))
			}
		}

		return false
	})

	fullTextSearch := make([]string, 0)
	for _, value := range pathToValues {
		intlibjson.ForEachValueInObject(data, value, func(currentValuePathKeyArrayIndexes []any, valueFound any) bool {
			if valueFound != nil {
				fullTextSearch = append(fullTextSearch, fmt.Sprintf("%+v", valueFound))
			}
			return false
		})
	}

	return fullTextSearch
}

type AuthIDsSelectQueryPKey struct {
	Name      string
	ProcessAs string
}

func (n *PostgresSelectQuery) AuthorizationIDsGetSelectQuery(
	ctx context.Context,
	metadataModel map[string]any,
	metadataModelParentPath string,
	tableName string,
	primaryKeyColumns []AuthIDsSelectQueryPKey,
	creationIamGroupAuthorizationsIDColumnName string,
	deactivationIamGroupAuthorizationsIDColumnName string,
) (*SelectQuery, error) {
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
		TableName: tableName,
		Query:     "",
		Where:     make(map[string]map[int][][]string),
		Join:      make(map[string]*SelectQuery),
	}

	if tableUid, ok := metadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID].(string); ok && len(tableUid) > 0 {
		selectQuery.TableUid = tableUid
	} else {
		return nil, intlib.FunctionNameAndError(n.AuthorizationIDsGetSelectQuery, errors.New("tableUid is empty"))
	}

	if value, err := intlibmmodel.DatabaseGetColumnFields(metadataModel, selectQuery.TableUid, false, false); err != nil {
		return nil, intlib.FunctionNameAndError(n.AuthorizationIDsGetSelectQuery, fmt.Errorf("extract database column fields failed, error: %v", err))
	} else {
		selectQuery.Columns = value
	}

	for _, pkc := range primaryKeyColumns {
		if fgKeyString, ok := selectQuery.Columns.Fields[pkc.Name][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
			if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", pkc.Name, fgKeyString, pkc.ProcessAs, ""); len(value) > 0 {
				selectQuery.Where[pkc.Name] = value
			}
		}
	}

	if _, ok := selectQuery.Columns.Fields[creationIamGroupAuthorizationsIDColumnName][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", creationIamGroupAuthorizationsIDColumnName, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[creationIamGroupAuthorizationsIDColumnName] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[deactivationIamGroupAuthorizationsIDColumnName][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", deactivationIamGroupAuthorizationsIDColumnName, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[deactivationIamGroupAuthorizationsIDColumnName] = value
		}
	}

	creationIamGroupAuthorizationsIDJoinIamGroupAuthorizations := intlib.MetadataModelGenJoinKey(creationIamGroupAuthorizationsIDColumnName, intdoment.IamGroupAuthorizationsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, creationIamGroupAuthorizationsIDJoinIamGroupAuthorizations); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", creationIamGroupAuthorizationsIDJoinIamGroupAuthorizations, err))
	} else {
		if sq, err := n.IamGroupAuthorizationsGetSelectQuery(ctx, value, metadataModelParentPath); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", creationIamGroupAuthorizationsIDJoinIamGroupAuthorizations, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.IamGroupAuthorizationsRepository().ID, true),      //1
				GetJoinColumnName(selectQuery.TableUid, creationIamGroupAuthorizationsIDColumnName, false), //2
			)

			selectQuery.Join[creationIamGroupAuthorizationsIDJoinIamGroupAuthorizations] = sq
		}
	}

	deactivationIamGroupAuthorizationsIDJoinIamGroupAuthorizations := intlib.MetadataModelGenJoinKey(deactivationIamGroupAuthorizationsIDColumnName, intdoment.IamGroupAuthorizationsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, deactivationIamGroupAuthorizationsIDJoinIamGroupAuthorizations); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", deactivationIamGroupAuthorizationsIDJoinIamGroupAuthorizations, err))
	} else {
		if sq, err := n.IamGroupAuthorizationsGetSelectQuery(ctx, value, metadataModelParentPath); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", deactivationIamGroupAuthorizationsIDJoinIamGroupAuthorizations, err))
		} else {
			if len(sq.Where) == 0 {
				sq.JoinType = JOIN_LEFT
			} else {
				sq.JoinType = JOIN_INNER
			}
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.IamGroupAuthorizationsRepository().ID, true),          //1
				GetJoinColumnName(selectQuery.TableUid, deactivationIamGroupAuthorizationsIDColumnName, false), //2
			)

			selectQuery.Join[deactivationIamGroupAuthorizationsIDJoinIamGroupAuthorizations] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}

func GetSelectQuery(selectQuery *SelectQuery, whereAfterJoin bool) (string, *SelectQueryExtract) {
	sqe := NewExtractSelectQuery(whereAfterJoin)
	selectQueryExtract := sqe.Extract(selectQuery)
	selectQueryWhere := NewSelectQueryWhere()
	where := ""
	if whereAfterJoin {
		where = selectQueryWhere.Merge(selectQuery)
	}
	query := ""
	if len(selectQueryExtract.DirectoryGroupsSubGroupsCTE) > 0 {
		query += selectQueryExtract.DirectoryGroupsSubGroupsCTE + " "
	}
	query += selectQueryExtract.Query
	whereSet := false
	if len(selectQueryExtract.DirectoryGroupsSubGroupsCTECondition) > 0 {
		query += " WHERE " + selectQueryExtract.DirectoryGroupsSubGroupsCTECondition
		whereSet = true
	}

	if len(selectQueryExtract.WhereAnd) > 0 {
		if whereSet {
			query += " AND "
		} else {
			query += " WHERE "
		}
		query += selectQueryExtract.WhereAnd
		whereSet = true
	}

	if len(where) > 0 {
		if whereSet {
			query += " AND "
		} else {
			query += " WHERE "
		}
		query += where
	}
	query = appendSortToQuery(query, selectQuery.SortAsc)
	query = appendLimitOffsetToQuery(query, selectQueryExtract.Limit, selectQueryExtract.Offset)
	query += ";"
	return query, selectQueryExtract
}

const (
	RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME string = "group_sub_groups"
)

func RecursiveDirectoryGroupsSubGroupsCte(parentGroupID uuid.UUID, cteName string) string {
	return fmt.Sprintf("WITH RECURSIVE %[1]s AS (SELECT %[2]s,%[3]s FROM %[4]s WHERE %[2]s = '%[5]s' UNION SELECT dgsb.%[2]s, dgsb.%[3]s FROM %[4]s dgsb INNER JOIN %[1]s ON %[1]s.%[3]s = dgsb.%[2]s)",
		cteName, //1
		intdoment.DirectoryGroupsSubGroupsRepository().ParentGroupID,  //2
		intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID,     //3
		intdoment.DirectoryGroupsSubGroupsRepository().RepositoryName, //4
		parentGroupID.String(), //5
	)
}

type SelectQueryExtract struct {
	DirectoryGroupsSubGroupsCTE          string
	DirectoryGroupsSubGroupsCTECondition string
	WhereAnd                             string
	Query                                string
	JoinQuery                            []string
	SelectColumns                        []string
	Columns                              []string
	Fields                               []any
	Join                                 map[string]*SelectQueryExtract
	Limit                                string
	Offset                               string
	SortAsc                              map[string]bool
	Distinct                             bool
}

type ExtractSelectQuery struct {
	whereAfterJoin bool
}

func NewExtractSelectQuery(whereAfterJoin bool) *ExtractSelectQuery {
	n := new(ExtractSelectQuery)
	n.whereAfterJoin = whereAfterJoin
	return n
}

func (n *ExtractSelectQuery) Extract(selectQuery *SelectQuery) *SelectQueryExtract {
	return n.extract(selectQuery, false)
}

func (n *ExtractSelectQuery) extract(selectQuery *SelectQuery, nested bool) *SelectQueryExtract {
	psq := new(SelectQueryExtract)
	psq.Columns = make([]string, 0)
	psq.Fields = make([]any, 0)
	psq.Join = make(map[string]*SelectQueryExtract)
	psq.JoinQuery = make([]string, 0)

	for _, columnName := range selectQuery.Columns.ColumnFieldsReadOrder {
		if field, ok := selectQuery.Columns.Fields[columnName]; ok {
			psq.SelectColumns = append(psq.SelectColumns, fmt.Sprintf("%s as %s", columnName, GetJoinColumnName(selectQuery.TableUid, columnName, true)))
			psq.Columns = append(psq.Columns, GetJoinColumnName(selectQuery.TableUid, columnName, true))
			psq.Fields = append(psq.Fields, field)
		}
	}
	// Process joins
	for key, value := range selectQuery.Join {
		psq.Join[key] = n.extract(value, true)
		psq.SelectColumns = append(psq.SelectColumns, psq.Join[key].Columns...)
		psq.Columns = append(psq.Columns, psq.Join[key].Columns...)
		psq.Fields = append(psq.Fields, psq.Join[key].Fields...)
		psq.JoinQuery = append(psq.JoinQuery, fmt.Sprintf("%s (%s) ON %s", value.JoinType, psq.Join[key].Query, strings.Join(value.JoinQuery, "AND")))
	}

	if nested && len(selectQuery.DirectoryGroupsSubGroupsCTE) > 0 {
		psq.Query = selectQuery.DirectoryGroupsSubGroupsCTE + " "
	} else {
		psq.DirectoryGroupsSubGroupsCTE = selectQuery.DirectoryGroupsSubGroupsCTE
	}

	psq.Query += "SELECT"

	if len(selectQuery.Distinct) > 0 {
		psq.Query += fmt.Sprintf(" DISTINCT ON (%s)", strings.Join(selectQuery.Distinct, " , "))
	}

	psq.Query += fmt.Sprintf(" %s FROM %s as %s", strings.Join(psq.SelectColumns, ","), selectQuery.TableName, selectQuery.TableUid)
	psq.Query += " " + strings.Join(psq.JoinQuery, " ")
	whereSet := false

	if (nested || !n.whereAfterJoin) && len(selectQuery.DirectoryGroupsSubGroupsCTECondition) > 0 {
		psq.Query += " WHERE " + selectQuery.DirectoryGroupsSubGroupsCTECondition
		whereSet = true
	} else {
		psq.DirectoryGroupsSubGroupsCTECondition = selectQuery.DirectoryGroupsSubGroupsCTECondition
	}

	if (nested || !n.whereAfterJoin) && len(selectQuery.WhereAnd) > 0 {
		if whereSet {
			psq.Query += " AND " + strings.Join(selectQuery.WhereAnd, " AND ")
		} else {
			psq.Query += " WHERE " + strings.Join(selectQuery.WhereAnd, " AND ")
		}
		whereSet = true
	} else {
		psq.WhereAnd = strings.Join(selectQuery.WhereAnd, " AND ")
	}

	if !n.whereAfterJoin {
		sqCopy := deep.MustCopy(selectQuery)
		sqCopy.Join = nil
		selectQueryWhere := NewSelectQueryWhere()
		where := selectQueryWhere.Merge(sqCopy)
		if len(where) > 0 {
			if whereSet {
				psq.Query += " AND " + where
			} else {
				psq.Query += " WHERE " + where
			}
		}
	}

	// Process Limit, Offset, Sort
	if nested && len(selectQuery.SortAsc) > 0 {
		psq.Query = appendSortToQuery(psq.Query, selectQuery.SortAsc)
	} else {
		psq.SortAsc = deep.MustCopy(selectQuery.SortAsc)
	}

	if nested && len(selectQuery.Limit) > 0 || len(selectQuery.Offset) > 0 {
		psq.Query = appendLimitOffsetToQuery(psq.Query, psq.Limit, psq.Offset)
	} else {
		psq.Limit = selectQuery.Limit
		psq.Offset = selectQuery.Offset
	}

	return psq
}

func appendSortToQuery(query string, sortAsc map[string]bool) string {
	sort := make([]string, 0)
	for column, value := range sortAsc {
		if value {
			sort = append(sort, fmt.Sprintf("%s ASC", column))
		} else {
			sort = append(sort, fmt.Sprintf("%s DESC", column))
		}
	}

	if len(sort) > 0 {
		return query + " ORDER BY " + strings.Join(sort, " , ")
	}

	return query
}

func appendLimitOffsetToQuery(query string, limit string, offset string) string {
	if len(limit) > 0 {
		query += " LIMIT " + limit
	}
	if len(offset) > 0 {
		query += " OFFSET" + offset
	}
	return query
}

type SelecteQueryWhere struct {
	groupedqueryconditions map[int][][][]string
}

func (n *SelecteQueryWhere) Merge(selectQuery *SelectQuery) string {
	n.Group(selectQuery)

	if len(n.groupedqueryconditions) == 0 {
		return ""
	}

	oneResult := make([]string, 0)
	for _, oneValue := range n.groupedqueryconditions {
		if len(oneValue) == 0 {
			continue
		}

		twoResult := make([]string, 0)
		for _, twoValue := range oneValue {
			if len(twoValue) == 0 {
				continue
			}

			threeResult := make([]string, 0)
			for _, threeValue := range twoValue {
				if len(threeValue) == 0 {
					continue
				}

				noOfFour := 0
				for _, andCondition := range threeValue {
					if len(andCondition) > 0 {
						noOfFour += 1
					}
				}

				if noOfFour == 0 {
					continue
				}

				fourResult := ""
				if noOfFour > 1 {
					fourResult += "("
				}
				fourResult += strings.Join(threeValue, " OR ")
				if noOfFour > 1 {
					fourResult += ")"
				}
				threeResult = append(threeResult, fourResult)
			}

			if len(threeResult) == 0 {
				continue
			}

			threeResultString := ""
			if len(threeResult) > 1 {
				threeResultString += "("
			}
			threeResultString += strings.Join(threeResult, " AND ")
			if len(threeResult) > 1 {
				threeResultString += ")"
			}
			twoResult = append(twoResult, threeResultString)
		}

		if len(twoResult) == 0 {
			continue
		}

		twoResultString := ""
		if len(twoResult) > 1 {
			twoResultString += "("
		}
		twoResultString += strings.Join(twoResult, " AND ")
		if len(twoResult) > 1 {
			twoResultString += ")"
		}
		oneResult = append(oneResult, twoResultString)
	}

	if len(oneResult) == 0 {
		return ""
	}

	oneResultString := ""
	if len(oneResult) > 1 {
		oneResultString += "("
	}
	oneResultString += strings.Join(oneResult, " OR ")
	if len(oneResult) > 1 {
		oneResultString += ")"
	}

	return oneResultString
}

func (n *SelecteQueryWhere) Group(selectQuery *SelectQuery) {
	for _, qConditions := range selectQuery.Where {
		for qKey, qCondition := range qConditions {
			if _, ok := n.groupedqueryconditions[qKey]; !ok {
				n.groupedqueryconditions[qKey] = make([][][]string, 0)
			}
			n.groupedqueryconditions[qKey] = append(n.groupedqueryconditions[qKey], qCondition)
		}
	}

	if len(selectQuery.Join) > 0 {
		for _, sq := range selectQuery.Join {
			n.Group(sq)
		}
	}
}

func (n *SelecteQueryWhere) GroupQueryConditionsResult() map[int][][][]string {
	return n.groupedqueryconditions
}

func NewSelectQueryWhere() *SelecteQueryWhere {
	n := new(SelecteQueryWhere)

	n.groupedqueryconditions = make(map[int][][][]string)

	return n
}

const (
	PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE string = "PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE"
	PROCESS_QUERY_CONDITION_AS_ARRAY        string = "PROCESS_QUERY_CONDITION_AS_ARRAY"
	PROCESS_QUERY_CONDITION_AS_JSONB        string = "PROCESS_QUERY_CONDITION_AS_JSONB"
)

const (
	COLUMN_GREATER_THAN string = ">"
	COLUMN_LESS_THAN    string = "<"
	COLUMN_EQUAL_TO     string = "="
)

const (
	JOIN_INNER string = "INNER JOIN"
	JOIN_LEFT  string = "LEFT JOIN"
	JOIN_RIGHT string = "RIGHT JOIN"
)

type SelectQuery struct {
	DirectoryGroupsSubGroupsCTEName      string
	DirectoryGroupsSubGroupsCTE          string
	DirectoryGroupsSubGroupsCTECondition string
	TableUid                             string
	TableName                            string
	Query                                string
	Columns                              intlibmmodel.DatabaseColumnFields
	Where                                map[string]map[int][][]string
	WhereAnd                             []string
	SortAsc                              map[string]bool
	Limit                                string
	Offset                               string
	JoinQuery                            []string
	Join                                 map[string]*SelectQuery
	JoinType                             string
	Distinct                             []string
}

func (n *SelectQuery) appendLimitOffset(selectQueryMetadataModel map[string]any) {
	if limit, ok := selectQueryMetadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_LIMIT].(float64); ok && limit > 0 {
		n.Limit = fmt.Sprintf("%f", limit)
	}

	if offset, ok := selectQueryMetadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_OFFSET].(float64); ok && offset > 0 {
		n.Offset = fmt.Sprintf("%f", offset)
	}
}

func (n *SelectQuery) appendSort() {
	for columnName, fg := range n.Columns.Fields {
		if sortByAsc, ok := fg[intlibmmodel.FIELD_GROUP_PROP_DATABASE_SORT_BY_ASC].(bool); ok {
			if n.SortAsc == nil {
				n.SortAsc = make(map[string]bool)
			}
			n.SortAsc[columnName] = sortByAsc
		}
	}
}

func (n *SelectQuery) appendDistinct() {
	for columnName, fg := range n.Columns.Fields {
		if distinct, ok := fg[intlibmmodel.FIELD_GROUP_PROP_DATABASE_DISTINCT].(bool); ok && distinct {
			if n.Distinct == nil {
				n.Distinct = make([]string, 0)
			}
			n.Distinct = append(n.Distinct, columnName)
		}
	}
}

func (n *PostgresSelectQuery) extractChildMetadataModel(parentMetadataModel map[string]any, childMetadataModelFgKeySuffix string) (map[string]any, error) {
	childMetadataModel := make(map[string]any)
	childMetadataModelFgKey := ""
	parentMetadataModelFgKey := ""

	intlibmmodel.ForEachFieldGroup(parentMetadataModel, func(property map[string]any) bool {
		if fgKeyString, ok := property[intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
			fgStringArray := strings.Split(fgKeyString, ".")
			if fgStringArray[len(fgStringArray)-1] == childMetadataModelFgKeySuffix {
				childMetadataModel = deep.MustCopy(property)
				childMetadataModelFgKey = fgKeyString
				parentMetadataModelFgKey = fgStringArray[len(fgStringArray)-2]
				return true
			}
		}
		return false
	})

	if len(childMetadataModelFgKey) == 0 && len(parentMetadataModelFgKey) == 0 {
		return nil, errors.New("childMetadataModel not found")
	}

	return childMetadataModel, nil
}

type PostgresSelectQuery struct {
	logger                      intdomint.Logger
	repo                        intdomint.IamRepository
	iamCredential               *intdoment.IamCredentials
	iamAuthorizationRules       *intdoment.IamAuthorizationRules
	startSearchDirectoryGroupID uuid.UUID
	authContextDirectoryGroupID uuid.UUID
	queryConditions             []intdoment.MetadataModelQueryConditions
	skipIfFGDisabled            bool
	skipIfDataExtraction        bool
	whereAfterJoin              bool
}

func (n *PostgresSelectQuery) getWhereCondition(quoteColumns bool, tableUID string, tableName string, columnName string, columnFgKey string, processQueryConditionAs string, fullTextSearchColumnName string) map[int][][]string {
	var where map[int][][]string

	if len(n.queryConditions) == 0 {
		return where
	}

	initWhere := func() {
		if where == nil {
			where = make(map[int][][]string)
		}
	}

	initWhereQIndex := func(qIndex int) {
		if where[qIndex] == nil {
			where[qIndex] = make([][]string, 0)
		}
	}

	for qIndex, qValues := range n.queryConditions {
		for qPath, qValue := range qValues {
			if qValue.DatabaseTableCollectionUid != tableUID {
				continue
			}

			if len(tableName) > 0 && qValue.DatabaseTableCollectionName != tableName {
				continue
			}
			if len(columnName) > 0 && qValue.DatabaseFieldColumnName != columnName {
				continue
			}
			if len(tableName) > 0 && len(fullTextSearchColumnName) > 0 {
				if len(qValue.DatabaseFullTextSearchQuery) > 0 {
					initWhere()
					initWhereQIndex(qIndex)
					whereOr := make([][]string, 1)
					whereOr[0] = make([]string, 1)
					whereOr[0][0] = GetFullTextSearchQuery(GetJoinColumnName(tableUID, fullTextSearchColumnName, quoteColumns), qValue.DatabaseFullTextSearchQuery)
					where[qIndex] = whereOr
				}
				continue
			}

			if len(qValue.FilterCondition) < 1 {
				continue
			}
			switch processQueryConditionAs {
			case PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE:
				if len(columnFgKey) > 0 && qPath != intlibmmodel.GetPathToValue(columnFgKey, true, intlibmmodel.ARRAY_PATH_PLACEHOLDER) {
					continue
				}
				if value := GetWhereConditionForSingleValue(GetJoinColumnName(tableUID, columnName, quoteColumns), qValue.FilterCondition); len(value) > 0 {
					initWhere()
					initWhereQIndex(qIndex)
					where[qIndex] = value
				}
			case PROCESS_QUERY_CONDITION_AS_ARRAY:
				if value := GetWhereConditionForArrayValue(GetJoinColumnName(tableUID, columnName, quoteColumns), qValue.FilterCondition); len(value) > 0 {
					initWhere()
					initWhereQIndex(qIndex)
					where[qIndex] = value
				}
			case PROCESS_QUERY_CONDITION_AS_JSONB:
				if value := GetWhereConditionForJsonbValue(GetJoinColumnName(tableUID, columnName, quoteColumns), qPath, columnFgKey, qValue.FilterCondition); len(value) > 0 {
					initWhere()
					initWhereQIndex(qIndex)
					where[qIndex] = value
				}
			}
		}
	}

	return where
}

func GetWhereConditionForJsonbValue(columName string, queryPath string, columnFgKey string, filterCondition [][]intdoment.MetadataModelFilterCondition) [][]string {
	jsonbElementColumnName := "jsonb_element"

	parentPath := intlibmmodel.GetPathToValue(columnFgKey, true, intlibmmodel.ARRAY_PATH_PLACEHOLDER)
	queryPath = strings.Replace(queryPath, parentPath+intlibmmodel.ARRAY_PATH_PLACEHOLDER, "$", 1)
	if !strings.HasSuffix(queryPath, intlibmmodel.ARRAY_PATH_PLACEHOLDER) {
		queryPath += intlibmmodel.ARRAY_PATH_PLACEHOLDER
	}
	selectJsonbElementsQuery := fmt.Sprintf("jsonb_path_query(%s, '%s')", columName, queryPath)
	selectJsonbElementsQueryTrim := fmt.Sprintf(`trim(both '\"' FROM %s::text) as jsonb_element_text`, jsonbElementColumnName)
	jsonbElementTextColumnName := "jsonb_element_text"

	var whereOr [][]string
	whereOrInitialized := false
	initWhereOr := func() {
		if !whereOrInitialized {
			whereOr = make([][]string, 0)
			whereOrInitialized = true
		}
	}

	whereAndInitialized := make(map[int]bool)
	initWhereAnd := func(orIndex int, whereAnd *[]string) {
		if !whereAndInitialized[orIndex] {
			*whereAnd = make([]string, 0)
			whereAndInitialized[orIndex] = true
		}
	}

	for orIndex, orFilterConditions := range filterCondition {
		var whereAnd []string
		whereAndInitialized[orIndex] = false

		for _, andFilterCondition := range orFilterConditions {
			andCondition := ""

			switch andFilterCondition.Condition {
			case intlibmmodel.FILTER_CONDTION_NO_OF_ENTRIES_GREATER_THAN:
				c := ColumnNumberCondition(fmt.Sprintf("(SELECT COUNT(*) FROM %s)", selectJsonbElementsQuery), COLUMN_GREATER_THAN, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = c
			case intlibmmodel.FILTER_CONDTION_NO_OF_ENTRIES_LESS_THAN:
				c := ColumnNumberCondition(fmt.Sprintf("(SELECT COUNT(*) FROM %s)", selectJsonbElementsQuery), COLUMN_LESS_THAN, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = c
			case intlibmmodel.FILTER_CONDTION_NO_OF_ENTRIES_EQUAL_TO:
				c := ColumnNumberCondition(fmt.Sprintf("(SELECT COUNT(*) FROM %s)", selectJsonbElementsQuery), COLUMN_EQUAL_TO, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = c
			case intlibmmodel.FILTER_CONDTION_NUMBER_GREATER_THAN:
				c := ColumnNumberCondition(jsonbElementColumnName, COLUMN_GREATER_THAN, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s as %s WHERE %s)", selectJsonbElementsQuery, jsonbElementColumnName, c)
			case intlibmmodel.FILTER_CONDTION_NUMBER_LESS_THAN:
				c := ColumnNumberCondition(jsonbElementColumnName, COLUMN_LESS_THAN, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s as %s WHERE %s)", selectJsonbElementsQuery, jsonbElementColumnName, c)
			case intlibmmodel.FILTER_CONDTION_TIMESTAMP_GREATER_THAN:
				if valueString, ok := andFilterCondition.Value.(string); ok {
					c := ColumnTimestampCondition(jsonbElementTextColumnName, andFilterCondition.DateTimeFormat, COLUMN_GREATER_THAN, valueString)
					if len(c) == 0 {
						break
					}
					initWhereOr()
					initWhereAnd(orIndex, &whereAnd)
					if andFilterCondition.Negate {
						c = "NOT " + c
					}
					andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s, %s WHERE %s)", selectJsonbElementsQuery, selectJsonbElementsQueryTrim, c)
				}
			case intlibmmodel.FILTER_CONDTION_TIMESTAMP_LESS_THAN:
				if valueString, ok := andFilterCondition.Value.(string); ok {
					c := ColumnTimestampCondition(jsonbElementTextColumnName, andFilterCondition.DateTimeFormat, COLUMN_LESS_THAN, valueString)
					if len(c) == 0 {
						break
					}
					initWhereOr()
					initWhereAnd(orIndex, &whereAnd)
					if andFilterCondition.Negate {
						c = "NOT " + c
					}
					andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s, %s WHERE %s)", selectJsonbElementsQuery, selectJsonbElementsQueryTrim, c)
				}
			case intlibmmodel.FILTER_CONDTION_TEXT_BEGINS_WITH, intlibmmodel.FILTER_CONDTION_TEXT_CONTAINS, intlibmmodel.FILTER_CONDTION_TEXT_ENDS_WITH:
				if fValueString, ok := andFilterCondition.Value.(string); ok && len(fValueString) > 0 {
					c := ColumTextCondition(jsonbElementTextColumnName, andFilterCondition.Condition, fValueString)
					if len(c) == 0 {
						break
					}
					initWhereOr()
					initWhereAnd(orIndex, &whereAnd)
					if andFilterCondition.Negate {
						c = "NOT " + c
					}
					andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s AS %s, %s WHERE %s)", selectJsonbElementsQuery, jsonbElementColumnName, selectJsonbElementsQueryTrim, c)
				}
			case intlibmmodel.FILTER_CONDTION_EQUAL_TO:
				if valueEqual, ok := andFilterCondition.Value.(map[string]any); ok {
					if vValue, ok := valueEqual[intlibmmodel.FIELD_SELECT_PROP_VALUE]; ok {
						cName := jsonbElementColumnName
						if len(andFilterCondition.DateTimeFormat) > 0 || reflect.TypeOf(vValue).Kind() == reflect.String {
							cName = jsonbElementTextColumnName
						}
						c := ColumnEqualTo(cName, valueEqual)
						if len(c) == 0 {
							break
						}
						initWhereOr()
						initWhereAnd(orIndex, &whereAnd)
						if andFilterCondition.Negate {
							c = "NOT " + c
						}

						if len(andFilterCondition.DateTimeFormat) > 0 || reflect.TypeOf(vValue).Kind() == reflect.String {
							andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s as %s, %s WHERE %s)", selectJsonbElementsQuery, jsonbElementColumnName, selectJsonbElementsQueryTrim, c)
						} else {
							andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s as %s WHERE %s)", selectJsonbElementsQuery, jsonbElementColumnName, c)
						}
					}
				}

			}

			if len(andCondition) > 0 {
				whereAnd = append(whereAnd, andCondition)
			}
		}

		if len(whereAnd) > 0 {
			whereOr = append(whereOr, whereAnd)
		}
	}

	return whereOr
}

func GetWhereConditionForArrayValue(columName string, filterCondition [][]intdoment.MetadataModelFilterCondition) [][]string {
	selectArrayElementsQuery := fmt.Sprintf("unnest(%s)", columName)
	arrayElementColumnName := "array_element"

	var whereOr [][]string
	whereOrInitialized := false
	initWhereOr := func() {
		if !whereOrInitialized {
			whereOr = make([][]string, 0)
			whereOrInitialized = true
		}
	}

	whereAndInitialized := make(map[int]bool)
	initWhereAnd := func(orIndex int, whereAnd *[]string) {
		if !whereAndInitialized[orIndex] {
			*whereAnd = make([]string, 0)
			whereAndInitialized[orIndex] = true
		}
	}

	for orIndex, orFilterConditions := range filterCondition {
		var whereAnd []string
		whereAndInitialized[orIndex] = false

		for _, andFilterCondition := range orFilterConditions {
			andCondition := ""

			switch andFilterCondition.Condition {
			case intlibmmodel.FILTER_CONDTION_NO_OF_ENTRIES_GREATER_THAN:
				c := ColumnNumberCondition(fmt.Sprintf("(SELECT COUNT(*) FROM %s)", selectArrayElementsQuery), COLUMN_GREATER_THAN, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = c
			case intlibmmodel.FILTER_CONDTION_NO_OF_ENTRIES_LESS_THAN:
				c := ColumnNumberCondition(fmt.Sprintf("(SELECT COUNT(*) FROM %s)", selectArrayElementsQuery), COLUMN_LESS_THAN, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = c
			case intlibmmodel.FILTER_CONDTION_NO_OF_ENTRIES_EQUAL_TO:
				c := ColumnNumberCondition(fmt.Sprintf("(SELECT COUNT(*) FROM %s)", selectArrayElementsQuery), COLUMN_EQUAL_TO, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = c
			case intlibmmodel.FILTER_CONDTION_NUMBER_GREATER_THAN:
				c := ColumnNumberCondition(arrayElementColumnName, COLUMN_GREATER_THAN, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s as %s WHERE %s)", selectArrayElementsQuery, arrayElementColumnName, c)
			case intlibmmodel.FILTER_CONDTION_NUMBER_LESS_THAN:
				c := ColumnNumberCondition(arrayElementColumnName, COLUMN_LESS_THAN, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s as %s WHERE %s)", selectArrayElementsQuery, arrayElementColumnName, c)
			case intlibmmodel.FILTER_CONDTION_TIMESTAMP_GREATER_THAN:
				if valueString, ok := andFilterCondition.Value.(string); ok {
					c := ColumnTimestampCondition(arrayElementColumnName, andFilterCondition.DateTimeFormat, COLUMN_GREATER_THAN, valueString)
					if len(c) == 0 {
						break
					}
					initWhereOr()
					initWhereAnd(orIndex, &whereAnd)
					if andFilterCondition.Negate {
						c = "NOT " + c
					}
					andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s as %s WHERE %s)", selectArrayElementsQuery, arrayElementColumnName, c)
				}
			case intlibmmodel.FILTER_CONDTION_TIMESTAMP_LESS_THAN:
				if valueString, ok := andFilterCondition.Value.(string); ok {
					c := ColumnTimestampCondition(arrayElementColumnName, andFilterCondition.DateTimeFormat, COLUMN_LESS_THAN, valueString)
					if len(c) == 0 {
						break
					}
					initWhereOr()
					initWhereAnd(orIndex, &whereAnd)
					if andFilterCondition.Negate {
						c = "NOT " + c
					}
					andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s as %s WHERE %s)", selectArrayElementsQuery, arrayElementColumnName, c)
				}
			case intlibmmodel.FILTER_CONDTION_TEXT_BEGINS_WITH, intlibmmodel.FILTER_CONDTION_TEXT_CONTAINS, intlibmmodel.FILTER_CONDTION_TEXT_ENDS_WITH:
				if fValueString, ok := andFilterCondition.Value.(string); ok && len(fValueString) > 0 {
					c := ColumTextCondition(arrayElementColumnName, andFilterCondition.Condition, fValueString)
					if len(c) == 0 {
						break
					}
					initWhereOr()
					initWhereAnd(orIndex, &whereAnd)
					if andFilterCondition.Negate {
						c = "NOT " + c
					}
					andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s as %s WHERE %s)", selectArrayElementsQuery, arrayElementColumnName, c)
				}
			case intlibmmodel.FILTER_CONDTION_EQUAL_TO:
				c := ColumnEqualTo(arrayElementColumnName, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = fmt.Sprintf("EXISTS(SELECT 1 FROM %s as %s WHERE %s)", selectArrayElementsQuery, arrayElementColumnName, c)
			}

			if len(andCondition) > 0 {
				whereAnd = append(whereAnd, andCondition)
			}
		}

		if len(whereAnd) > 0 {
			whereOr = append(whereOr, whereAnd)
		}
	}

	return whereOr
}

func GetWhereConditionForSingleValue(columName string, filterCondition [][]intdoment.MetadataModelFilterCondition) [][]string {
	var whereOr [][]string
	whereOrInitialized := false
	initWhereOr := func() {
		if !whereOrInitialized {
			whereOr = make([][]string, 0)
			whereOrInitialized = true
		}
	}

	whereAndInitialized := make(map[int]bool)
	initWhereAnd := func(orIndex int, whereAnd *[]string) {
		if !whereAndInitialized[orIndex] {
			*whereAnd = make([]string, 0)
			whereAndInitialized[orIndex] = true
		}
	}

	for orIndex, orFilterConditions := range filterCondition {
		var whereAnd []string
		whereAndInitialized[orIndex] = false

		for _, andFilterCondition := range orFilterConditions {
			andCondition := ""

			switch andFilterCondition.Condition {
			case intlibmmodel.FILTER_CONDTION_NO_OF_ENTRIES_GREATER_THAN:
				if fValueFloat, ok := andFilterCondition.Value.(float64); ok {
					if andFilterCondition.Negate {
						if fValueFloat <= 0 {
							initWhereOr()
							initWhereAnd(orIndex, &whereAnd)
							andCondition = ColumnISNull(columName)
						}
					} else {
						if fValueFloat >= 0 {
							initWhereOr()
							initWhereAnd(orIndex, &whereAnd)
							andCondition = ColumnIsNotNull(columName)
						}
					}
				}
			case intlibmmodel.FILTER_CONDTION_NO_OF_ENTRIES_LESS_THAN:
				if fValueFloat, ok := andFilterCondition.Value.(float64); ok {
					if andFilterCondition.Negate {
						if fValueFloat >= 0 {
							initWhereOr()
							initWhereAnd(orIndex, &whereAnd)
							andCondition = ColumnIsNotNull(columName)
						}
					} else {
						if fValueFloat <= 0 {
							initWhereOr()
							initWhereAnd(orIndex, &whereAnd)
							andCondition = ColumnISNull(columName)
						}
					}

				}
			case intlibmmodel.FILTER_CONDTION_NO_OF_ENTRIES_EQUAL_TO:
				if fValueFloat, ok := andFilterCondition.Value.(float64); ok {
					initWhereOr()
					initWhereAnd(orIndex, &whereAnd)
					if andFilterCondition.Negate {
						if fValueFloat > 0 {
							andCondition = ColumnISNull(columName)
						} else {
							andCondition = ColumnIsNotNull(columName)
						}
					} else {
						if fValueFloat > 0 {
							andCondition = ColumnIsNotNull(columName)
						} else {
							andCondition = ColumnISNull(columName)
						}
					}
				}
			case intlibmmodel.FILTER_CONDTION_NUMBER_GREATER_THAN:
				c := ColumnNumberCondition(columName, COLUMN_GREATER_THAN, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = c
			case intlibmmodel.FILTER_CONDTION_NUMBER_LESS_THAN:
				c := ColumnNumberCondition(columName, COLUMN_LESS_THAN, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = c
			case intlibmmodel.FILTER_CONDTION_TIMESTAMP_GREATER_THAN:
				if valueString, ok := andFilterCondition.Value.(string); ok && len(valueString) > 0 {
					c := ColumnTimestampCondition(columName, andFilterCondition.DateTimeFormat, COLUMN_GREATER_THAN, valueString)
					if len(c) == 0 {
						break
					}
					initWhereOr()
					initWhereAnd(orIndex, &whereAnd)
					if andFilterCondition.Negate {
						c = "NOT " + c
					}
					andCondition = c
				}
			case intlibmmodel.FILTER_CONDTION_TIMESTAMP_LESS_THAN:
				if valueString, ok := andFilterCondition.Value.(string); ok {
					c := ColumnTimestampCondition(columName, andFilterCondition.DateTimeFormat, COLUMN_LESS_THAN, valueString)
					if len(c) == 0 {
						break
					}
					initWhereOr()
					initWhereAnd(orIndex, &whereAnd)
					if andFilterCondition.Negate {
						c = "NOT " + c
					}
					andCondition = c
				}
			case intlibmmodel.FILTER_CONDTION_TEXT_BEGINS_WITH, intlibmmodel.FILTER_CONDTION_TEXT_CONTAINS, intlibmmodel.FILTER_CONDTION_TEXT_ENDS_WITH:
				if fValueString, ok := andFilterCondition.Value.(string); ok && len(fValueString) > 0 {
					c := ColumTextCondition(columName, andFilterCondition.Condition, fValueString)
					if len(c) == 0 {
						break
					}
					initWhereOr()
					initWhereAnd(orIndex, &whereAnd)
					if andFilterCondition.Negate {
						c = "NOT " + c
					}
					andCondition = c
				}
			case intlibmmodel.FILTER_CONDTION_EQUAL_TO:
				c := ColumnEqualTo(columName, andFilterCondition.Value)
				if len(c) == 0 {
					break
				}
				initWhereOr()
				initWhereAnd(orIndex, &whereAnd)
				if andFilterCondition.Negate {
					c = "NOT " + c
				}
				andCondition = c
			}

			if len(andCondition) > 0 {
				whereAnd = append(whereAnd, andCondition)
			}
		}

		if len(whereAnd) > 0 {
			whereOr = append(whereOr, whereAnd)
		}
	}

	return whereOr
}

func ColumTextCondition(column string, filterCondition string, value string) string {
	switch filterCondition {
	case intlibmmodel.FILTER_CONDTION_TEXT_BEGINS_WITH:
		return fmt.Sprintf("%s LIKE '%s%%'", column, value)
	case intlibmmodel.FILTER_CONDTION_TEXT_CONTAINS:
		return fmt.Sprintf("%s LIKE '%%%s%%'", column, value)
	case intlibmmodel.FILTER_CONDTION_TEXT_ENDS_WITH:
		return fmt.Sprintf("%s LIKE '%%%s'", column, value)
	}

	return ""
}

func ColumnTimestampCondition(column string, dateTimeFormat string, columnCondition string, value string) string {
	condition := ""
	switch dateTimeFormat {
	case intlibmmodel.FIELD_DATE_TIME_FORMAT_YYYYMMDDHHMM:
		condition = "(" //year
		condition += fmt.Sprintf("date_part('year', %s::timestamp without time zone) %s date_part('year', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += " OR(" //month
		condition += fmt.Sprintf("date_part('year', %s::timestamp without time zone) = date_part('year', '%s'::timestamp without time zone)", column, value)
		condition += " AND "
		condition += fmt.Sprintf("date_part('month', %s::timestamp without time zone) %s date_part('month', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += " OR(" //day
		condition += fmt.Sprintf("date_part('month', %s::timestamp without time zone) = date_part('month', '%s'::timestamp without time zone)", column, value)
		condition += " AND "
		condition += fmt.Sprintf("date_part('day', %s::timestamp without time zone) %s date_part('day', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += " OR(" //hour
		condition += fmt.Sprintf("date_part('day', %s::timestamp without time zone) = date_part('day', '%s'::timestamp without time zone)", column, value)
		condition += " AND " //hour
		condition += fmt.Sprintf("date_part('hour', %s::timestamp without time zone) %s date_part('hour', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += " OR(" //minute
		condition += fmt.Sprintf("date_part('hour', %s::timestamp without time zone) = date_part('hour', '%s'::timestamp without time zone)", column, value)
		condition += " AND "
		condition += fmt.Sprintf("date_part('minute', %s::timestamp without time zone) %s date_part('minute', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += ")" //minute
		condition += ")" //hour
		condition += ")" //day
		condition += ")" //month
		condition += ")" //year
	case intlibmmodel.FIELD_DATE_TIME_FORMAT_YYYYMMDD:
		condition = "(" //year
		condition += fmt.Sprintf("date_part('year', %s::timestamp without time zone) %s date_part('year', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += " OR(" //month
		condition += fmt.Sprintf("date_part('year', %s::timestamp without time zone) = date_part('year', '%s'::timestamp without time zone)", column, value)
		condition += " AND "
		condition += fmt.Sprintf("date_part('month', %s::timestamp without time zone) %s date_part('month', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += " OR(" //day
		condition += fmt.Sprintf("date_part('month', %s::timestamp without time zone) = date_part('month', '%s'::timestamp without time zone)", column, value)
		condition += " AND "
		condition += fmt.Sprintf("date_part('day', %s::timestamp without time zone) %s date_part('day', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += ")" //day
		condition += ")" //month
		condition += ")" //year
	case intlibmmodel.FIELD_DATE_TIME_FORMAT_YYYYMM:
		condition = "(" //year
		condition += fmt.Sprintf("date_part('year', %s::timestamp without time zone) %s date_part('year', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += " OR(" //month
		condition += fmt.Sprintf("date_part('year', %s::timestamp without time zone) = date_part('year', '%s'::timestamp without time zone)", column, value)
		condition += " AND "
		condition += fmt.Sprintf("date_part('month', %s::timestamp without time zone) %s date_part('month', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += ")" //month
		condition += ")" //year
	case intlibmmodel.FIELD_DATE_TIME_FORMAT_HHMM:
		condition = "(" //hour
		condition += fmt.Sprintf("date_part('hour', %s::timestamp without time zone) %s date_part('hour', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += " OR(" //minute
		condition += fmt.Sprintf("date_part('hour', %s::timestamp without time zone) = date_part('hour', '%s'::timestamp without time zone)", column, value)
		condition += " AND "
		condition += fmt.Sprintf("date_part('minute', %s::timestamp without time zone) %s date_part('minute', '%s'::timestamp without time zone)", column, columnCondition, value)
		condition += ")" //minute
		condition += ")" //hour
	case intlibmmodel.FIELD_DATE_TIME_FORMAT_YYYY:
		condition = fmt.Sprintf("date_part('year', %s::timestamp without time zone) %s date_part('year', '%s'::timestamp without time zone)", column, columnCondition, value)
	case intlibmmodel.FIELD_DATE_TIME_FORMAT_MM:
		condition = fmt.Sprintf("date_part('month', %s::timestamp without time zone) %s date_part('month', '%s'::timestamp without time zone)", column, columnCondition, value)
	}

	return condition
}

func ColumnNumberCondition(column string, columnCondition string, value any) string {
	var valueFloat float64
	if vFloat, ok := value.(float64); ok {
		valueFloat = vFloat
	} else if vInt, ok := value.(int); ok {
		valueFloat = float64(vInt)
	} else if vInt64, ok := value.(int64); ok {
		valueFloat = float64(vInt64)
	} else {
		return ""
	}

	return fmt.Sprintf("%s %s %f", column, columnCondition, valueFloat)
}

func ColumnEqualTo(column string, value any) string {
	if valueEqual, ok := value.(map[string]any); ok {
		if vValue, ok := valueEqual[intlibmmodel.FIELD_SELECT_PROP_VALUE]; ok {
			if valueString, ok := vValue.(string); ok && len(valueString) > 0 {
				if vDateTimeFormat, ok := valueEqual[intlibmmodel.FIELD_SELECT_DATE_TIME_FORMAT].(string); ok {
					return ColumnTimestampCondition(column, vDateTimeFormat, COLUMN_EQUAL_TO, valueString)
				}
				return fmt.Sprintf("%s = '%v'", column, vValue)
			}
			return fmt.Sprintf("%s = %v", column, vValue)
		}
	}

	return ""
}

func ColumnIsNotNull(column string) string {
	return fmt.Sprintf("%s IS NOT NULL", column)
}

func ColumnISNull(column string) string {
	return fmt.Sprintf("%s IS NULL", column)
}

func NewPostgresSelectQuery(
	logger intdomint.Logger,
	repo intdomint.IamRepository,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	startSearchDirectoryGroupID uuid.UUID,
	authContextDirectoryGroupID uuid.UUID,
	queryConditions []intdoment.MetadataModelQueryConditions,
	skipIfFGDisabled bool,
	skipIfDataExtraction bool,
	whereAfterJoin bool,
) *PostgresSelectQuery {
	n := new(PostgresSelectQuery)
	n.logger = logger
	n.repo = repo
	n.iamCredential = iamCredential
	n.iamAuthorizationRules = iamAuthorizationRules
	n.startSearchDirectoryGroupID = startSearchDirectoryGroupID
	n.authContextDirectoryGroupID = authContextDirectoryGroupID
	n.queryConditions = queryConditions
	n.skipIfFGDisabled = skipIfFGDisabled
	n.skipIfDataExtraction = skipIfDataExtraction
	n.whereAfterJoin = whereAfterJoin

	return n
}

type PostrgresRepository struct {
	db     *pgxpool.Pool
	logger intdomint.Logger
}

func (n *PostrgresRepository) Ping(ctx context.Context) error {
	if err := n.db.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func NewPostgresRepository(ctx context.Context, logger intdomint.Logger) (*PostrgresRepository, error) {
	databaseUri, err := GetPsqlDatabaseUri()
	if err != nil {
		return nil, err
	}

	dbConfig, err := pgxpool.ParseConfig(databaseUri)
	if err != nil {
		return nil, err
	}
	dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}
	if dpool, err := pgxpool.NewWithConfig(ctx, dbConfig); err != nil {
		return nil, err
	} else {
		return &PostrgresRepository{
			db:     dpool,
			logger: logger,
		}, nil
	}
}

func GetPsqlDatabaseUri() (string, error) {
	uri := new(url.URL)
	uri.Scheme = "postgres"

	if user := os.Getenv(ENV_PSQL_USER); len(user) > 0 {
		if password := os.Getenv(ENV_PSQL_PASSWORD); len(password) > 0 {
			uri.User = url.UserPassword(user, password)
		} else {
			return "", fmt.Errorf("env %s not set", ENV_PSQL_PASSWORD)
		}
	} else {
		return "", fmt.Errorf("env %s not set", ENV_PSQL_USER)
	}

	if value := os.Getenv(ENV_PSQL_HOST); len(value) > 0 {
		uri.Host = value
	} else {
		return "", fmt.Errorf("env %s not set", ENV_PSQL_HOST)
	}

	if value := os.Getenv(ENV_PSQL_PORT); len(value) > 0 {
		uri.Host += ":" + value
	} else {
		return "", fmt.Errorf("env %s not set", ENV_PSQL_PORT)
	}

	if value := os.Getenv(ENV_PSQL_DATABASE); len(value) > 0 {
		uri.Path = value
	} else {
		return "", fmt.Errorf("env %s not set", ENV_PSQL_DATABASE)
	}

	params := url.Values{}
	if value := os.Getenv(ENV_PSQL_SCHEMA); len(value) > 0 {
		params.Set("search_path", value)
	} else {
		return "", fmt.Errorf("env %s not set", ENV_PSQL_SCHEMA)
	}

	if value := os.Getenv(ENV_PSQL_SEARCH_PARAMS); len(value) > 0 {
		for _, kv := range strings.Split(value, " ") {
			if splitKV := strings.Split(kv, "="); len(splitKV) == 2 {
				params.Set(splitKV[0], splitKV[1])
			}
		}
	}

	uri.RawQuery = params.Encode()

	return uri.String(), nil
}

const (
	ENV_PSQL_USER          string = "PSQL_USER"
	ENV_PSQL_PASSWORD      string = "PSQL_PASSWORD"
	ENV_PSQL_HOST          string = "PSQL_HOST"
	ENV_PSQL_PORT          string = "PSQL_PORT"
	ENV_PSQL_DATABASE      string = "PSQL_DATABASE"
	ENV_PSQL_SCHEMA        string = "PSQL_SCHEMA"
	ENV_PSQL_SEARCH_PARAMS string = "PSQL_SEARCH_PARAMS"
)

func GetJoinColumnName(prefix string, suffix string, quote bool) string {
	if quote {
		return fmt.Sprintf("\"%s.%s\"", prefix, suffix)
	}
	return fmt.Sprintf("%s.%s", prefix, suffix)
}

func GetFullTextSearchQuery(fullTextSearchColumn string, searchQuery string) string {
	sqSplitSpace := strings.Split(searchQuery, " ")
	if len(sqSplitSpace) > 0 {
		newQuery := "("
		newQuery += fmt.Sprintf("%v @@ to_tsquery('%v:*')", fullTextSearchColumn, sqSplitSpace[0])
		for i := 1; i < len(sqSplitSpace); i++ {
			newQuery = newQuery + " AND " + fmt.Sprintf("%v @@ to_tsquery('%v:*')", fullTextSearchColumn, sqSplitSpace[i])
		}
		newQuery += ")"
		return newQuery
	} else {
		return fmt.Sprintf("%v @@ to_tsquery('%v:*')", fullTextSearchColumn, searchQuery)
	}
}

func GetandUpdateNextPlaceholder(nextPlaceholder *int) string {
	defer func() {
		*nextPlaceholder += 1
	}()
	return fmt.Sprintf("$%d", *nextPlaceholder)
}

func GetUpdateSetColumnsWithVQuery(colums []string, vQuery []string) string {
	setColumns := make([]string, 0)
	for index, value := range colums {
		setColumns = append(setColumns, fmt.Sprintf("%s = %s", value, vQuery[index]))
	}
	return strings.Join(setColumns, ", ")
}

func GetUpdateSetColumns(colums []string, nextPlaceholder *int) string {
	setColumns := make([]string, 0)
	for _, value := range colums {
		setColumns = append(setColumns, fmt.Sprintf("%s = $%d", value, *nextPlaceholder))
		*nextPlaceholder += 1
	}
	return strings.Join(setColumns, ", ")
}

func GetQueryPlaceholderString(noOfPlaceHolders int, nextPlaceholder *int) string {
	placeholders := make([]string, 0)
	for range noOfPlaceHolders {
		placeholders = append(placeholders, fmt.Sprintf("$%d", *nextPlaceholder))
		*nextPlaceholder += 1
	}
	return strings.Join(placeholders, ", ")
}
