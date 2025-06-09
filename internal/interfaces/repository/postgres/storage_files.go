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

func (n *PostrgresRepository) RepoStorageFilesDeleteOne(
	ctx context.Context,
	iamAuthRule *intdoment.IamAuthorizationRule,
	fileService intdomint.FileService,
	datum *intdoment.StorageFiles,
) error {
	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoStorageFilesDeleteOne, fmt.Errorf("start transaction to delete %s failed, error: %v", intdoment.StorageFilesRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s = $1;",
		intdoment.StorageFilesAuthorizationIDsRepository().RepositoryName, //1
		intdoment.StorageFilesAuthorizationIDsRepository().ID,             //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesDeleteOne))

	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		query := fmt.Sprintf(
			"DELETE FROM %[1]s WHERE %[2]s = $1;",
			intdoment.StorageFilesRepository().RepositoryName, //1
			intdoment.StorageFilesRepository().ID,             //2
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesDeleteOne))
		if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
			if err := fileService.Delete(ctx, datum); err != nil {
				transaction.Rollback(ctx)
				return err
			}

			if err := transaction.Commit(ctx); err != nil {
				return intlib.FunctionNameAndError(n.RepoStorageFilesDeleteOne, fmt.Errorf("commit transaction to delete %s failed, error: %v", intdoment.StorageFilesRepository().RepositoryName, err))
			}

			return nil
		} else {
			transaction.Rollback(ctx)
		}
	}

	transaction, err = n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoStorageFilesDeleteOne, fmt.Errorf("start transaction to deactivate %s failed, error: %v", intdoment.StorageFilesRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s = $2;",
		intdoment.StorageFilesAuthorizationIDsRepository().RepositoryName,                       //1
		intdoment.StorageFilesAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID, //2
		intdoment.StorageFilesAuthorizationIDsRepository().ID,                                   //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesDeleteOne))
	if _, err := transaction.Exec(ctx, query, iamAuthRule.ID, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoStorageFilesDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.StorageFilesAuthorizationIDsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = NOW() WHERE %[3]s = $1;",
		intdoment.StorageFilesRepository().RepositoryName, //1
		intdoment.StorageFilesRepository().DeactivatedOn,  //2
		intdoment.StorageFilesRepository().ID,             //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoStorageFilesDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.StorageFilesRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return intlib.FunctionNameAndError(n.RepoStorageFilesDeleteOne, fmt.Errorf("commit transaction to update deactivation of %s failed, error: %v", intdoment.StorageFilesRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoStorageFilesFindOneForDeletionByID(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	datum *intdoment.StorageFiles,
	columns []string,
) (*intdoment.StorageFiles, *intdoment.IamAuthorizationRule, error) {
	storageFilesMModel, err := intlib.MetadataModelGet(intdoment.StorageFilesRepository().RepositoryName)
	if err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(storageFilesMModel, intdoment.StorageFilesRepository().RepositoryName, false, false); err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.StorageFilesRepository().ID) {
		columns = append(columns, intdoment.StorageFilesRepository().ID)
	}

	if !slices.Contains(columns, intdoment.StorageFilesRepository().DirectoryGroupsID) {
		columns = append(columns, intdoment.StorageFilesRepository().DirectoryGroupsID)
	}

	dataRows := make([]any, 0)
	iamAuthRule := new(intdoment.IamAuthorizationRule)
	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_DELETE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		if len(iamCredential.DirectoryID) == 0 {
			return nil, nil, intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		}
		query := fmt.Sprintf(
			"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s = $2;",
			strings.Join(columns, " , "),                      //1
			intdoment.StorageFilesRepository().RepositoryName, //2
			intdoment.StorageFilesRepository().ID,             //3
			intdoment.StorageFilesRepository().DirectoryID,    //4
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesFindOneForDeletionByID))

		rows, err := n.db.Query(ctx, query, datum.ID[0], iamCredential.DirectoryID[0])
		if err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.StorageFilesRepository().RepositoryName, err))
		}
		defer rows.Close()
		for rows.Next() {
			if r, err := rows.Values(); err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, err)
			} else {
				dataRows = append(dataRows, r)
			}
		}
		if len(dataRows) > 0 {
			iamAuthRule = iamAuthorizationRule[0]
		}
	}

	if len(dataRows) == 0 {
		if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			iamCredential,
			authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_DELETE,
					RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
				},
			},
			iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			query := fmt.Sprintf(
				"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s = $2;",
				strings.Join(columns, " , "),                         //1
				intdoment.StorageFilesRepository().RepositoryName,    //2
				intdoment.StorageFilesRepository().ID,                //3
				intdoment.StorageFilesRepository().DirectoryGroupsID, //4
			)
			n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesFindOneForDeletionByID))

			rows, err := n.db.Query(ctx, query, datum.ID[0], authContextDirectoryGroupID)
			if err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.StorageFilesRepository().RepositoryName, err))
			}
			defer rows.Close()
			for rows.Next() {
				if r, err := rows.Values(); err != nil {
					return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, err)
				} else {
					dataRows = append(dataRows, r)
				}
			}
			if len(dataRows) > 0 {
				iamAuthRule = iamAuthorizationRule[0]
			}
		}
	}

	if len(dataRows) == 0 {
		if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			iamCredential,
			authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_DELETE_OTHERS,
					RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
				},
			},
			iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			cteName := fmt.Sprintf("%s_%s", intdoment.StorageFilesRepository().RepositoryName, RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME)
			query := fmt.Sprintf(
				"%[1]s SELECT %[2]s FROM %[3]s WHERE %[4]s = $1 AND %[5]s;",
				RecursiveDirectoryGroupsSubGroupsCte(authContextDirectoryGroupID, cteName), //1
				strings.Join(columns, " , "),                                               //2
				intdoment.StorageFilesRepository().RepositoryName,                          //3
				intdoment.StorageFilesRepository().ID,                                      //4
				fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.StorageFilesRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName), //5
			)
			n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesFindOneForDeletionByID))

			rows, err := n.db.Query(ctx, query, datum.ID[0])
			if err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.StorageFilesRepository().RepositoryName, err))
			}
			defer rows.Close()
			for rows.Next() {
				if r, err := rows.Values(); err != nil {
					return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, err)
				} else {
					dataRows = append(dataRows, r)
				}
			}
			if len(dataRows) > 0 {
				iamAuthRule = iamAuthorizationRule[0]
			}
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(storageFilesMModel, nil, false, false, columns)
	if err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoStorageFilesFindOneForDeletionByID))
		return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, fmt.Errorf("more than one %s found", intdoment.StorageFilesRepository().RepositoryName))
	}

	storageFile := new(intdoment.StorageFiles)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing storageFile", "storageFile", string(jsonData), "function", intlib.FunctionName(n.RepoStorageFilesFindOneForDeletionByID))
		if err := json.Unmarshal(jsonData, storageFile); err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoStorageFilesFindOneForDeletionByID, err)
		}
	}

	return storageFile, iamAuthRule, nil
}

func (n *PostrgresRepository) RepoStorageFilesUpdateOne(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	datum *intdoment.StorageFiles,
) error {
	query := ""
	where := ""

	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		cteName := fmt.Sprintf("%s_%s", intdoment.StorageFilesRepository().RepositoryName, RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME)
		query = RecursiveDirectoryGroupsSubGroupsCte(authContextDirectoryGroupID, cteName)
		where = fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.StorageFilesRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName)
	}
	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		whereQuery := fmt.Sprintf("%s = '%s'", intdoment.StorageFilesRepository().DirectoryGroupsID, authContextDirectoryGroupID.String())
		if len(where) > 0 {
			where += " OR " + whereQuery
		} else {
			where = whereQuery
		}
	}
	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		whereQuery := ""
		if len(iamCredential.DirectoryID) > 0 {
			whereQuery = fmt.Sprintf("%s = '%s'", intdoment.StorageFilesRepository().DirectoryID, iamCredential.DirectoryID[0].String())
		} else {
			whereQuery = fmt.Sprintf("%s = TRUE", intdoment.StorageFilesRepository().EditUnauthorized)
		}

		if len(where) > 0 {
			where += " OR " + whereQuery
		} else {
			where = whereQuery
		}
	}
	if len(where) == 0 {
		where = fmt.Sprintf("%s = TRUE", intdoment.StorageFilesRepository().EditUnauthorized)
	}

	valuesToUpdate := make([]any, 0)
	columnsToUpdate := make([]string, 0)
	if v, c, err := n.RepoStorageFilesValidateAndGetColumnsAndData(datum, false); err != nil {
		return err
	} else if len(c) == 0 || len(v) == 0 {
		return intlib.NewError(http.StatusBadRequest, "no values to update")
	} else {
		valuesToUpdate = append(valuesToUpdate, v...)
		columnsToUpdate = append(columnsToUpdate, c...)
	}

	nextPlaceholder := 1
	query += fmt.Sprintf(
		" UPDATE %[1]s SET %[2]s WHERE %[3]s = %[4]s AND %[5]s IS NULL AND (%[6]s);",
		intdoment.StorageFilesRepository().RepositoryName,      //1
		GetUpdateSetColumns(columnsToUpdate, &nextPlaceholder), //2
		intdoment.StorageFilesRepository().ID,                  //3
		GetandUpdateNextPlaceholder(&nextPlaceholder),          //4
		intdoment.StorageFilesRepository().DeactivatedOn,       //5
		where, //6
	)
	query = strings.TrimLeft(query, " \n")
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesUpdateOne))
	valuesToUpdate = append(valuesToUpdate, datum.ID[0])

	if _, err := n.db.Exec(ctx, query, valuesToUpdate...); err != nil {
		return intlib.FunctionNameAndError(n.RepoStorageFilesUpdateOne, fmt.Errorf("update %s failed, err: %v", intdoment.StorageFilesRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoStorageFilesInsertOne(
	ctx context.Context,
	iamAuthRule *intdoment.IamAuthorizationRule,
	fileService intdomint.FileService,
	datum *intdoment.StorageFiles,
	directoryID uuid.UUID,
	file io.Reader,
	columns []string,
) (*intdoment.StorageFiles, error) {
	storageFilesMetadataModel, err := intlib.MetadataModelGet(intdoment.StorageFilesRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(storageFilesMetadataModel, intdoment.StorageFilesRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.StorageFilesRepository().ID) {
		columns = append(columns, intdoment.StorageFilesRepository().ID)
	}

	if !slices.Contains(columns, intdoment.StorageFilesRepository().DirectoryGroupsID) {
		columns = append(columns, intdoment.StorageFilesRepository().DirectoryGroupsID)
	}

	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, fmt.Errorf("start transaction to create %s failed, error: %v", intdoment.StorageFilesRepository().RepositoryName, err))
	}

	valuesToInsert := []any{datum.DirectoryGroupsID[0], directoryID}
	columnsToInsert := []string{intdoment.StorageFilesRepository().DirectoryGroupsID, intdoment.StorageFilesRepository().DirectoryID}
	if v, c, err := n.RepoStorageFilesValidateAndGetColumnsAndData(datum, true); err != nil {
		return nil, err
	} else if len(c) == 0 || len(v) == 0 {
		return nil, intlib.NewError(http.StatusBadRequest, "no values to insert")
	} else {
		valuesToInsert = append(valuesToInsert, v...)
		columnsToInsert = append(columnsToInsert, c...)
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s) VALUES (%[3]s) RETURNING %[4]s;",
		intdoment.StorageFilesRepository().RepositoryName,            //1
		strings.Join(columnsToInsert, " , "),                         //2
		GetQueryPlaceholderString(len(valuesToInsert), &[]int{1}[0]), //3
		strings.Join(columns, " , "),                                 //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesInsertOne))

	rows, err := transaction.Query(ctx, query, valuesToInsert...)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.StorageFilesRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(storageFilesMetadataModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		transaction.Rollback(ctx)
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.StorageFilesRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		transaction.Rollback(ctx)
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoStorageFilesInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, fmt.Errorf("more than one %s found", intdoment.StorageFilesRepository().RepositoryName))
	}

	storageFile := new(intdoment.StorageFiles)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing storageFile", "storageFile", string(jsonData), "function", intlib.FunctionName(n.RepoStorageFilesInsertOne))
		if err := json.Unmarshal(jsonData, storageFile); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, err)
		}
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.StorageFilesAuthorizationIDsRepository().RepositoryName,                   //1
		intdoment.StorageFilesAuthorizationIDsRepository().ID,                               //2
		intdoment.StorageFilesAuthorizationIDsRepository().CreationIamGroupAuthorizationsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesInsertOne))

	if _, err := transaction.Exec(ctx, query, storageFile.ID[0], iamAuthRule.ID); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.StorageFilesAuthorizationIDsRepository().RepositoryName, err))
	}

	if err := fileService.Create(ctx, storageFile, file); err != nil {
		transaction.Rollback(ctx)
		return nil, err
	}

	if err := transaction.Commit(ctx); err != nil {
		if err := fileService.Delete(ctx, storageFile); err != nil {
			n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("Delete file failed: error %v", err), "function", intlib.FunctionName(n.RepoStorageFilesInsertOne))
		}
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesInsertOne, fmt.Errorf("commit transaction to create %s failed, error: %v", intdoment.StorageFilesRepository().RepositoryName, err))
	}

	return storageFile, nil
}

func (n *PostrgresRepository) RepoStorageFilesValidateAndGetColumnsAndData(datum *intdoment.StorageFiles, insert bool) ([]any, []string, error) {
	values := make([]any, 0)
	columns := make([]string, 0)

	if insert {
		if len(datum.StorageFileMimeType) == 0 || len(datum.StorageFileMimeType[0]) == 0 {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.StorageFilesRepository().StorageFileMimeType)
		} else {
			values = append(values, datum.StorageFileMimeType[0])
			columns = append(columns, intdoment.StorageFilesRepository().StorageFileMimeType)
		}

		if len(datum.OriginalName) == 0 || len(datum.OriginalName[0]) == 0 {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.StorageFilesRepository().OriginalName)
		} else {
			values = append(values, datum.OriginalName[0])
			columns = append(columns, intdoment.StorageFilesRepository().OriginalName)
		}
	}

	if len(datum.EditAuthorized) > 0 {
		values = append(values, datum.EditAuthorized[0])
		columns = append(columns, intdoment.StorageFilesRepository().EditAuthorized)
	}

	if len(datum.EditUnauthorized) > 0 {
		values = append(values, datum.EditUnauthorized[0])
		columns = append(columns, intdoment.StorageFilesRepository().EditUnauthorized)
	}

	if len(datum.ViewAuthorized) > 0 {
		values = append(values, datum.ViewAuthorized[0])
		columns = append(columns, intdoment.StorageFilesRepository().ViewAuthorized)
	}

	if len(datum.ViewUnauthorized) > 0 {
		values = append(values, datum.ViewUnauthorized[0])
		columns = append(columns, intdoment.StorageFilesRepository().ViewUnauthorized)
	}

	if insert {
		if len(datum.Tags) > 0 {
			values = append(values, datum.Tags)
			columns = append(columns, intdoment.StorageFilesRepository().Tags)
		}
	} else {
		if datum.Tags != nil && len(datum.Tags) >= 0 {
			values = append(values, datum.Tags)
			columns = append(columns, intdoment.StorageFilesRepository().Tags)
		}
	}

	if insert {
		if len(datum.SizeInBytes) > 0 {

			values = append(values, datum.SizeInBytes[0])
			columns = append(columns, intdoment.StorageFilesRepository().SizeInBytes)
		}
	}

	return values, columns, nil
}

func (n *PostrgresRepository) RepoStorageFilesSearch(
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
	selectQuery, err := pSelectQuery.StorageFilesGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoStorageFilesSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.StorageFilesRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoStorageFilesSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoStorageFilesSearch, err)
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

func (n *PostgresSelectQuery) StorageFilesGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
	quoteColumns := true
	if len(metadataModelParentPath) == 0 {
		metadataModelParentPath = "$"
		quoteColumns = false
	}
	if !n.whereAfterJoin {
		quoteColumns = false
	}

	selectQuery := SelectQuery{
		TableName: intdoment.StorageFilesRepository().RepositoryName,
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

	iamWhereOr := make([]string, 0)
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
			},
		},
		n.iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		iamWhereOr = append(iamWhereOr, fmt.Sprintf("%s.%s = TRUE", selectQuery.TableUid, intdoment.StorageFilesRepository().ViewAuthorized))
	}
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
			},
		},
		n.iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		if len(n.iamCredential.DirectoryID) > 0 {
			iamWhereOr = append(iamWhereOr, fmt.Sprintf("%s.%s = '%s'", selectQuery.TableUid, intdoment.StorageFilesRepository().DirectoryID, n.iamCredential.DirectoryID[0].String()))
		} else {
			iamWhereOr = append(iamWhereOr, fmt.Sprintf("%s.%s = TRUE", selectQuery.TableUid, intdoment.StorageFilesRepository().ViewUnauthorized))
		}
	}
	if len(iamWhereOr) == 0 {
		iamWhereOr = append(iamWhereOr, fmt.Sprintf("%s.%s = TRUE", selectQuery.TableUid, intdoment.StorageFilesRepository().ViewUnauthorized))
	}
	selectQuery.WhereAnd = append(selectQuery.WhereAnd, fmt.Sprintf("(%s)", strings.Join(iamWhereOr, " OR ")))

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
					RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
				},
			},
			n.iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			selectQuery.DirectoryGroupsSubGroupsCTEName = cteName
			selectQuery.DirectoryGroupsSubGroupsCTE = RecursiveDirectoryGroupsSubGroupsCte(n.startSearchDirectoryGroupID, cteName)
			cteWhere = append(cteWhere, fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.StorageFilesRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName))
		}

		if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			n.iamCredential,
			n.authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE,
					RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
				},
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
					RuleGroup: intdoment.AUTH_RULE_GROUP_STORAGE_FILES,
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
			cteWhere = append(cteWhere, fmt.Sprintf("%s = '%s'", intdoment.StorageFilesRepository().DirectoryGroupsID, n.startSearchDirectoryGroupID.String()))
		}

		if iamWhereOr != nil {
			cteWhere = append(cteWhere, fmt.Sprintf("%s = TRUE", intdoment.StorageFilesRepository().ViewAuthorized))
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

	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().ID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().ID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().ID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().DirectoryGroupsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().DirectoryGroupsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().DirectoryGroupsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().DirectoryID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().DirectoryID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().DirectoryID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().StorageFileMimeType][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().StorageFileMimeType, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().StorageFileMimeType] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().OriginalName][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().OriginalName, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().OriginalName] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().EditAuthorized][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().EditAuthorized, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().EditAuthorized] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().EditUnauthorized][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().EditUnauthorized, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().EditUnauthorized] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().ViewAuthorized][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().ViewAuthorized, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().ViewAuthorized] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().ViewUnauthorized][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().ViewUnauthorized, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().ViewUnauthorized] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().Tags][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().Tags, "", PROCESS_QUERY_CONDITION_AS_ARRAY, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().Tags] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().SizeInBytes][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().SizeInBytes, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().SizeInBytes] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().LastUpdatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().LastUpdatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().LastUpdatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.StorageFilesRepository().DeactivatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.StorageFilesRepository().DeactivatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.StorageFilesRepository().DeactivatedOn] = value
		}
	}
	if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, selectQuery.TableName, "", "", "", intdoment.StorageFilesRepository().FullTextSearch); len(value) > 0 {
		selectQuery.Where[intdoment.StorageFilesRepository().RepositoryName] = value
	}

	directoryGroupIDJoinDirectoryGroups := intlib.MetadataModelGenJoinKey(intdoment.StorageFilesRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, directoryGroupIDJoinDirectoryGroups); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", directoryGroupIDJoinDirectoryGroups, err))
	} else {
		if sq, err := n.DirectoryGroupsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", directoryGroupIDJoinDirectoryGroups, err))
		} else {
			if len(sq.Where) == 0 {
				sq.JoinType = JOIN_LEFT
			} else {
				sq.JoinType = JOIN_INNER
			}
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryRepository().ID, true),                             //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.StorageFilesRepository().DirectoryGroupsID, false), //2
			)

			selectQuery.Join[directoryGroupIDJoinDirectoryGroups] = sq
		}
	}

	directoryIDJoinDirectory := intlib.MetadataModelGenJoinKey(intdoment.StorageFilesRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)
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
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryRepository().ID, true),                       //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.StorageFilesRepository().DirectoryID, false), //2
			)

			selectQuery.Join[directoryIDJoinDirectory] = sq
		}
	}

	storageFilesJoinStorageFilesAuthorizationIDs := intlib.MetadataModelGenJoinKey(intdoment.StorageFilesRepository().RepositoryName, intdoment.StorageFilesAuthorizationIDsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, storageFilesJoinStorageFilesAuthorizationIDs); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", storageFilesJoinStorageFilesAuthorizationIDs, err))
	} else {
		if sq, err := n.AuthorizationIDsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
			intdoment.StorageFilesAuthorizationIDsRepository().RepositoryName,
			[]AuthIDsSelectQueryPKey{{Name: intdoment.StorageFilesAuthorizationIDsRepository().ID, ProcessAs: PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE}},
			intdoment.StorageFilesAuthorizationIDsRepository().CreationIamGroupAuthorizationsID,
			intdoment.StorageFilesAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", storageFilesJoinStorageFilesAuthorizationIDs, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.StorageFilesAuthorizationIDsRepository().ID, true), //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.StorageFilesRepository().ID, false),       //2
			)

			selectQuery.Join[storageFilesJoinStorageFilesAuthorizationIDs] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}
