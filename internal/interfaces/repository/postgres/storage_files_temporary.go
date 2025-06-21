package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (n *PostrgresRepository) RepoStorageFilesDeleteTemporaryFiles(
	ctx context.Context,
	fileService intdomint.FileService,
) (*intdoment.StorageFilesTemporaryDelete, error) {
	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE (NOW() - %[3]s) > '30 minutes';",
		intdoment.StorageFilesTemporaryRepository().ID,             //1
		intdoment.StorageFilesTemporaryRepository().RepositoryName, //2
		intdoment.StorageFilesTemporaryRepository().CreatedOn,      //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesDeleteTemporaryFiles))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporarySearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.StorageFilesTemporaryRepository().RepositoryName, err))
	}
	defer rows.Close()

	result := new(intdoment.StorageFilesTemporaryDelete)
	result.Success = 0
	result.Failed = make([]error, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			result.Failed = append(result.Failed, intlib.FunctionNameAndError(n.RepoStorageFilesTemporarySearch, err))
		} else {
			if fileID, ok := r[0].(uuid.UUID); ok {
				storageFileTemporary := new(intdoment.StorageFilesTemporary)
				storageFileTemporary.ID = []uuid.UUID{fileID}

				if err := n.RepoStorageFilesTemporaryDeleteOne(ctx, nil, fileService, storageFileTemporary); err != nil {
					result.Failed = append(result.Failed, err)
				} else {
					result.Success += 1
				}
			}
		}
	}

	return result, nil
}

func (n *PostrgresRepository) RepoStorageFilesTemporaryDeleteOne(
	ctx context.Context,
	iamAuthRule *intdoment.IamAuthorizationRule,
	fileService intdomint.FileService,
	datum *intdoment.StorageFilesTemporary,
) error {
	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryDeleteOne, fmt.Errorf("start transaction to delete %s failed, error: %v", intdoment.StorageFilesTemporaryRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s = $1;",
		intdoment.StorageFilesTemporaryRepository().RepositoryName, //1
		intdoment.StorageFilesTemporaryRepository().ID,             //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesTemporaryDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		sf := new(intdoment.StorageFiles)
		sf.ID = datum.ID
		if err := fileService.Delete(ctx, sf); err != nil {
			transaction.Rollback(ctx)
			return err
		}

		if err := transaction.Commit(ctx); err != nil {
			return intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryDeleteOne, fmt.Errorf("commit transaction to delete %s failed, error: %v", intdoment.StorageFilesTemporaryRepository().RepositoryName, err))
		}

		return nil
	} else {
		transaction.Rollback(ctx)
	}

	if err := transaction.Commit(ctx); err != nil {
		return intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryDeleteOne, fmt.Errorf("commit transaction to update deactivation of %s failed, error: %v", intdoment.StorageFilesTemporaryRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoStorageFilesTemporaryInsertOne(
	ctx context.Context,
	fileService intdomint.FileService,
	datum *intdoment.StorageFilesTemporary,
	file io.Reader,
	columns []string,
) (*intdoment.StorageFilesTemporary, error) {
	storageFilesMetadataModel, err := intlib.MetadataModelGet(intdoment.StorageFilesTemporaryRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(storageFilesMetadataModel, intdoment.StorageFilesTemporaryRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.StorageFilesTemporaryRepository().ID) {
		columns = append(columns, intdoment.StorageFilesTemporaryRepository().ID)
	}

	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, fmt.Errorf("start transaction to create %s failed, error: %v", intdoment.StorageFilesTemporaryRepository().RepositoryName, err))
	}

	valuesToInsert := make([]any, 0)
	columnsToInsert := make([]string, 0)
	if v, c, err := n.RepoStorageFilesTemporaryValidateAndGetColumnsAndData(datum, true); err != nil {
		return nil, err
	} else if len(c) == 0 || len(v) == 0 {
		return nil, intlib.NewError(http.StatusBadRequest, "no values to insert")
	} else {
		valuesToInsert = append(valuesToInsert, v...)
		columnsToInsert = append(columnsToInsert, c...)
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s) VALUES (%[3]s) RETURNING %[4]s;",
		intdoment.StorageFilesTemporaryRepository().RepositoryName,   //1
		strings.Join(columnsToInsert, " , "),                         //2
		GetQueryPlaceholderString(len(valuesToInsert), &[]int{1}[0]), //3
		strings.Join(columns, " , "),                                 //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesTemporaryInsertOne))

	rows, err := transaction.Query(ctx, query, valuesToInsert...)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.StorageFilesTemporaryRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(storageFilesMetadataModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		transaction.Rollback(ctx)
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.StorageFilesTemporaryRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		transaction.Rollback(ctx)
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoStorageFilesTemporaryInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, fmt.Errorf("more than one %s found", intdoment.StorageFilesTemporaryRepository().RepositoryName))
	}

	storageFileTemporary := new(intdoment.StorageFilesTemporary)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing storageFileTemporary", "storageFileTemporary", string(jsonData), "function", intlib.FunctionName(n.RepoStorageFilesTemporaryInsertOne))
		if err := json.Unmarshal(jsonData, storageFileTemporary); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, err)
		}
	}

	sf := new(intdoment.StorageFiles)
	sf.ID = storageFileTemporary.ID
	sf.OriginalName = storageFileTemporary.OriginalName
	sf.StorageFileMimeType = storageFileTemporary.StorageFileMimeType
	sf.Tags = storageFileTemporary.Tags
	sf.SizeInBytes = storageFileTemporary.SizeInBytes
	sf.CreatedOn = storageFileTemporary.CreatedOn

	if err := fileService.Create(ctx, sf, file); err != nil {
		transaction.Rollback(ctx)
		return nil, err
	}

	if err := transaction.Commit(ctx); err != nil {
		if err := fileService.Delete(ctx, sf); err != nil {
			n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("Delete file failed: error %v", err), "function", intlib.FunctionName(n.RepoStorageFilesTemporaryInsertOne))
		}
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporaryInsertOne, fmt.Errorf("commit transaction to create %s failed, error: %v", intdoment.StorageFilesTemporaryRepository().RepositoryName, err))
	}

	return storageFileTemporary, nil
}

func (n *PostrgresRepository) RepoStorageFilesTemporaryValidateAndGetColumnsAndData(datum *intdoment.StorageFilesTemporary, insert bool) ([]any, []string, error) {
	values := make([]any, 0)
	columns := make([]string, 0)

	if insert {
		if len(datum.StorageFileMimeType) == 0 || len(datum.StorageFileMimeType[0]) == 0 {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.StorageFilesTemporaryRepository().StorageFileMimeType)
		} else {
			values = append(values, datum.StorageFileMimeType[0])
			columns = append(columns, intdoment.StorageFilesTemporaryRepository().StorageFileMimeType)
		}

		if len(datum.OriginalName) == 0 || len(datum.OriginalName[0]) == 0 {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.StorageFilesTemporaryRepository().OriginalName)
		} else {
			values = append(values, datum.OriginalName[0])
			columns = append(columns, intdoment.StorageFilesTemporaryRepository().OriginalName)
		}
	}

	if insert {
		if len(datum.Tags) > 0 {
			values = append(values, datum.Tags)
			columns = append(columns, intdoment.StorageFilesTemporaryRepository().Tags)
		}
	} else {
		if datum.Tags != nil && len(datum.Tags) >= 0 {
			values = append(values, datum.Tags)
			columns = append(columns, intdoment.StorageFilesTemporaryRepository().Tags)
		}
	}

	if insert {
		if len(datum.SizeInBytes) > 0 {

			values = append(values, datum.SizeInBytes[0])
			columns = append(columns, intdoment.StorageFilesTemporaryRepository().SizeInBytes)
		}
	}

	return values, columns, nil
}

func (n *PostrgresRepository) RepoStorageFilesTemporarySearch(
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
	selectQuery, err := pSelectQuery.StorageFilesTemporaryGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporarySearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesTemporarySearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporarySearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.StorageFilesTemporaryRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporarySearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporarySearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesTemporarySearch, err)
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

func (n *PostgresSelectQuery) StorageFilesTemporaryGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
	quoteColumns := true
	if len(metadataModelParentPath) == 0 {
		metadataModelParentPath = "$"
		quoteColumns = false
	}
	if !n.whereAfterJoin {
		quoteColumns = false
	}

	selectQuery := SelectQuery{
		TableName: intdoment.StorageFilesTemporaryRepository().RepositoryName,
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

	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesTemporaryRepository().ID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesTemporaryRepository().ID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesTemporaryRepository().ID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesTemporaryRepository().StorageFileMimeType][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesTemporaryRepository().StorageFileMimeType, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesTemporaryRepository().StorageFileMimeType] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesTemporaryRepository().OriginalName][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesTemporaryRepository().OriginalName, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesTemporaryRepository().OriginalName] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesTemporaryRepository().Tags][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesTemporaryRepository().Tags, "", PROCESS_QUERY_CONDITION_AS_ARRAY, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesTemporaryRepository().Tags] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesTemporaryRepository().SizeInBytes][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesTemporaryRepository().SizeInBytes, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesTemporaryRepository().SizeInBytes] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesTemporaryRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesTemporaryRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesTemporaryRepository().CreatedOn] = value
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}
