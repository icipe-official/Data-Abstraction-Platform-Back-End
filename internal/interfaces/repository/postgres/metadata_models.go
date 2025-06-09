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

func (n *PostrgresRepository) RepoMetadataModelsDeleteOne(
	ctx context.Context,
	iamAuthRule *intdoment.IamAuthorizationRule,
	datum *intdoment.MetadataModels,
) error {
	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDeleteOne, fmt.Errorf("start transaction to delete %s failed, error: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s = $1;",
		intdoment.MetadataModelsAuthorizationIDsRepository().RepositoryName, //1
		intdoment.MetadataModelsAuthorizationIDsRepository().ID,             //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDeleteOne))

	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		query := fmt.Sprintf(
			"DELETE FROM %[1]s WHERE %[2]s = $1;",
			intdoment.MetadataModelsRepository().RepositoryName, //1
			intdoment.MetadataModelsRepository().ID,             //2
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDeleteOne))
		if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
			if err := transaction.Commit(ctx); err != nil {
				return intlib.FunctionNameAndError(n.RepoMetadataModelsDeleteOne, fmt.Errorf("commit transaction to delete %s failed, error: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
			}
			return nil
		} else {
			transaction.Rollback(ctx)
		}
	}

	transaction, err = n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDeleteOne, fmt.Errorf("start transaction to deactivate %s failed, error: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s = $2;",
		intdoment.MetadataModelsAuthorizationIDsRepository().RepositoryName,                       //1
		intdoment.MetadataModelsAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID, //2
		intdoment.MetadataModelsAuthorizationIDsRepository().ID,                                   //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDeleteOne))
	if _, err := transaction.Exec(ctx, query, iamAuthRule.ID, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.MetadataModelsAuthorizationIDsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = NOW() WHERE %[3]s = $1;",
		intdoment.MetadataModelsRepository().RepositoryName, //1
		intdoment.MetadataModelsRepository().DeactivatedOn,  //2
		intdoment.MetadataModelsRepository().ID,             //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return intlib.FunctionNameAndError(n.RepoMetadataModelsDeleteOne, fmt.Errorf("commit transaction to update deactivation of %s failed, error: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoMetadataModelsFindOneForDeletionByID(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	datum *intdoment.MetadataModels,
	columns []string,
) (*intdoment.MetadataModels, *intdoment.IamAuthorizationRule, error) {
	metadataModelsMModel, err := intlib.MetadataModelGet(intdoment.MetadataModelsRepository().RepositoryName)
	if err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(metadataModelsMModel, intdoment.MetadataModelsRepository().RepositoryName, false, false); err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.MetadataModelsRepository().ID) {
		columns = append(columns, intdoment.MetadataModelsRepository().ID)
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
				RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		if len(iamCredential.DirectoryID) == 0 {
			return nil, nil, intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		}
		query := fmt.Sprintf(
			"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s = $2;",
			strings.Join(columns, " , "),                        //1
			intdoment.MetadataModelsRepository().RepositoryName, //2
			intdoment.MetadataModelsRepository().ID,             //3
			intdoment.MetadataModelsRepository().DirectoryID,    //4
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsFindOneForDeletionByID))

		rows, err := n.db.Query(ctx, query, datum.ID[0], iamCredential.DirectoryID[0])
		if err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
		}
		defer rows.Close()
		for rows.Next() {
			if r, err := rows.Values(); err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, err)
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
					RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
				},
			},
			iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			query := fmt.Sprintf(
				"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s = $2;",
				strings.Join(columns, " , "),                           //1
				intdoment.MetadataModelsRepository().RepositoryName,    //2
				intdoment.MetadataModelsRepository().ID,                //3
				intdoment.MetadataModelsRepository().DirectoryGroupsID, //4
			)
			n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsFindOneForDeletionByID))

			rows, err := n.db.Query(ctx, query, datum.ID[0], authContextDirectoryGroupID)
			if err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
			}
			defer rows.Close()
			for rows.Next() {
				if r, err := rows.Values(); err != nil {
					return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, err)
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
					RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
				},
			},
			iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			cteName := fmt.Sprintf("%s_%s", intdoment.MetadataModelsRepository().RepositoryName, RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME)
			query := fmt.Sprintf(
				"%[1]s SELECT %[2]s FROM %[3]s WHERE %[4]s = $1 AND %[5]s;",
				RecursiveDirectoryGroupsSubGroupsCte(authContextDirectoryGroupID, cteName), //1
				strings.Join(columns, " , "),                                               //2
				intdoment.MetadataModelsRepository().RepositoryName,                        //3
				intdoment.MetadataModelsRepository().ID,                                    //4
				fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.MetadataModelsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName), //5
			)
			n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsFindOneForDeletionByID))

			rows, err := n.db.Query(ctx, query, datum.ID[0])
			if err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
			}
			defer rows.Close()
			for rows.Next() {
				if r, err := rows.Values(); err != nil {
					return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, err)
				} else {
					dataRows = append(dataRows, r)
				}
			}
			if len(dataRows) > 0 {
				iamAuthRule = iamAuthorizationRule[0]
			}
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(metadataModelsMModel, nil, false, false, columns)
	if err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoMetadataModelsFindOneForDeletionByID))
		return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, fmt.Errorf("more than one %s found", intdoment.MetadataModelsRepository().RepositoryName))
	}

	metadataModel := new(intdoment.MetadataModels)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing metadataModel", "metadataModel", string(jsonData), "function", intlib.FunctionName(n.RepoMetadataModelsFindOneForDeletionByID))
		if err := json.Unmarshal(jsonData, metadataModel); err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoMetadataModelsFindOneForDeletionByID, err)
		}
	}

	return metadataModel, iamAuthRule, nil
}

func (n *PostrgresRepository) RepoMetadataModelsUpdateOne(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	datum *intdoment.MetadataModels,
) error {
	query := ""
	authWhere := ""

	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		cteName := fmt.Sprintf("%s_%s", intdoment.MetadataModelsRepository().RepositoryName, RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME)
		query = RecursiveDirectoryGroupsSubGroupsCte(authContextDirectoryGroupID, cteName)
		authWhere = fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.MetadataModelsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName)
	}
	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		whereQuery := fmt.Sprintf("%s = '%s'", intdoment.MetadataModelsRepository().DirectoryGroupsID, authContextDirectoryGroupID.String())
		if len(authWhere) > 0 {
			authWhere += " OR " + whereQuery
		} else {
			authWhere = whereQuery
		}
	}
	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		whereQuery := ""
		if len(iamCredential.DirectoryID) > 0 {
			whereQuery = fmt.Sprintf("%s = '%s'", intdoment.MetadataModelsRepository().DirectoryID, iamCredential.DirectoryID[0].String())
		} else {
			whereQuery = fmt.Sprintf("%s = TRUE", intdoment.MetadataModelsRepository().EditUnauthorized)
		}

		if len(authWhere) > 0 {
			authWhere += " OR " + whereQuery
		} else {
			authWhere = whereQuery
		}
	}
	if len(authWhere) == 0 {
		authWhere = fmt.Sprintf("%s = TRUE", intdoment.MetadataModelsRepository().EditUnauthorized)
	}

	valuesToUpdate := make([]any, 0)
	columnsToUpdate := make([]string, 0)
	if v, c, err := n.RepoMetadataModelsValidateAndGetColumnsAndData(datum, false); err != nil {
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
		intdoment.MetadataModelsRepository().RepositoryName,    //1
		GetUpdateSetColumns(columnsToUpdate, &nextPlaceholder), //2
		intdoment.MetadataModelsRepository().ID,                //3
		GetandUpdateNextPlaceholder(&nextPlaceholder),          //4
		intdoment.MetadataModelsRepository().DeactivatedOn,     //5
		authWhere, //6
	)
	query = strings.TrimLeft(query, " \n")
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsUpdateOne))
	valuesToUpdate = append(valuesToUpdate, datum.ID[0])

	if _, err := n.db.Exec(ctx, query, valuesToUpdate...); err != nil {
		return intlib.FunctionNameAndError(n.RepoMetadataModelsUpdateOne, fmt.Errorf("update %s failed, err: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoMetadataModelsInsertOne(
	ctx context.Context,
	iamAuthRule *intdoment.IamAuthorizationRule,
	directoryID uuid.UUID,
	directoryGroupID uuid.UUID,
	datum *intdoment.MetadataModels,
	columns []string,
) (*intdoment.MetadataModels, error) {
	metadataModelsMModel, err := intlib.MetadataModelGet(intdoment.MetadataModelsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(metadataModelsMModel, intdoment.MetadataModelsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.MetadataModelsRepository().ID) {
		columns = append(columns, intdoment.MetadataModelsRepository().ID)
	}

	valuesToInsert := []any{directoryGroupID, directoryID}
	columnsToInsert := []string{intdoment.MetadataModelsRepository().DirectoryGroupsID, intdoment.MetadataModelsRepository().DirectoryID}
	if v, c, err := n.RepoMetadataModelsValidateAndGetColumnsAndData(datum, true); err != nil {
		return nil, err
	} else if len(c) == 0 || len(v) == 0 {
		return nil, intlib.NewError(http.StatusBadRequest, "no values to insert")
	} else {
		valuesToInsert = append(valuesToInsert, v...)
		columnsToInsert = append(columnsToInsert, c...)
	}

	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, fmt.Errorf("start transaction to create %s failed, error: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s) VALUES (%[3]s) RETURNING %[4]s;",
		intdoment.MetadataModelsRepository().RepositoryName,          //1
		strings.Join(columnsToInsert, " , "),                         //2
		GetQueryPlaceholderString(len(valuesToInsert), &[]int{1}[0]), //3
		strings.Join(columns, " , "),                                 //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsInsertOne))

	rows, err := transaction.Query(ctx, query, valuesToInsert...)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(metadataModelsMModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		transaction.Rollback(ctx)
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.MetadataModelsRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		transaction.Rollback(ctx)
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoMetadataModelsInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, fmt.Errorf("more than one %s found", intdoment.MetadataModelsRepository().RepositoryName))
	}

	metadataModel := new(intdoment.MetadataModels)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing metadataModel", "metadataModel", string(jsonData), "function", intlib.FunctionName(n.RepoMetadataModelsInsertOne))
		if err := json.Unmarshal(jsonData, metadataModel); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, err)
		}
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.MetadataModelsAuthorizationIDsRepository().RepositoryName,                   //1
		intdoment.MetadataModelsAuthorizationIDsRepository().ID,                               //2
		intdoment.MetadataModelsAuthorizationIDsRepository().CreationIamGroupAuthorizationsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsInsertOne))

	if _, err := transaction.Exec(ctx, query, metadataModel.ID[0], iamAuthRule.ID); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.MetadataModelsAuthorizationIDsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsInsertOne, fmt.Errorf("commit transaction to create %s failed, error: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
	}

	return metadataModel, nil
}

func (n *PostrgresRepository) RepoMetadataModelsValidateAndGetColumnsAndData(datum *intdoment.MetadataModels, insert bool) ([]any, []string, error) {
	values := make([]any, 0)
	columns := make([]string, 0)

	if len(datum.Name) == 0 || len(datum.Name[0]) < 4 {
		if insert {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.MetadataModelsRepository().Name)
		}
	} else {
		values = append(values, datum.Name[0])
		columns = append(columns, intdoment.MetadataModelsRepository().Name)
	}

	if len(datum.Description) == 0 || len(datum.Description[0]) < 4 {
		if insert {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.MetadataModelsRepository().Description)
		}
	} else {
		values = append(values, datum.Description[0])
		columns = append(columns, intdoment.MetadataModelsRepository().Description)
	}

	if len(datum.Data) == 0 {
		if insert {
			return nil, nil, fmt.Errorf("%s is not valid", intdoment.MetadataModelsRepository().Data)
		}
	} else {
		values = append(values, datum.Data[0])
		columns = append(columns, intdoment.MetadataModelsRepository().Data)
	}

	if len(datum.EditAuthorized) > 0 {
		values = append(values, datum.EditAuthorized[0])
		columns = append(columns, intdoment.MetadataModelsRepository().EditAuthorized)
	}

	if len(datum.EditUnauthorized) > 0 {
		values = append(values, datum.EditUnauthorized[0])
		columns = append(columns, intdoment.MetadataModelsRepository().EditUnauthorized)
	}

	if len(datum.ViewAuthorized) > 0 {
		values = append(values, datum.ViewAuthorized[0])
		columns = append(columns, intdoment.MetadataModelsRepository().ViewAuthorized)
	}

	if len(datum.ViewUnauthorized) > 0 {
		values = append(values, datum.ViewUnauthorized[0])
		columns = append(columns, intdoment.MetadataModelsRepository().ViewUnauthorized)
	}

	if insert {
		if len(datum.Tags) > 0 {
			values = append(values, datum.Tags)
			columns = append(columns, intdoment.MetadataModelsRepository().Tags)
		}
	} else {
		if datum.Tags != nil && len(datum.Tags) >= 0 {
			values = append(values, datum.Tags)
			columns = append(columns, intdoment.MetadataModelsRepository().Tags)
		}
	}

	return values, columns, nil
}

func (n *PostrgresRepository) RepoMetadataModelsSearch(
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
	selectQuery, err := pSelectQuery.MetadataModelsGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelsSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.MetadataModelsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoMetadataModelsSearch, err)
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

func (n *PostgresSelectQuery) MetadataModelsGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
	quoteColumns := true
	if len(metadataModelParentPath) == 0 {
		metadataModelParentPath = "$"
		quoteColumns = false
	}
	if !n.whereAfterJoin {
		quoteColumns = false
	}

	selectQuery := SelectQuery{
		TableName: intdoment.MetadataModelsRepository().RepositoryName,
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
				RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
			},
		},
		n.iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		iamWhereOr = append(iamWhereOr, fmt.Sprintf("%s.%s = TRUE", selectQuery.TableUid, intdoment.MetadataModelsRepository().ViewAuthorized))
	}
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
			},
		},
		n.iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		if len(n.iamCredential.DirectoryID) > 0 {
			iamWhereOr = append(iamWhereOr, fmt.Sprintf("%s.%s = '%s'", selectQuery.TableUid, intdoment.MetadataModelsRepository().DirectoryID, n.iamCredential.DirectoryID[0].String()))
		} else {
			iamWhereOr = append(iamWhereOr, fmt.Sprintf("%s.%s = TRUE", selectQuery.TableUid, intdoment.MetadataModelsRepository().ViewUnauthorized))
		}
	}
	if len(iamWhereOr) == 0 {
		iamWhereOr = append(iamWhereOr, fmt.Sprintf("%s.%s = TRUE", selectQuery.TableUid, intdoment.MetadataModelsRepository().ViewUnauthorized))
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
					RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
				},
			},
			n.iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			selectQuery.DirectoryGroupsSubGroupsCTEName = cteName
			selectQuery.DirectoryGroupsSubGroupsCTE = RecursiveDirectoryGroupsSubGroupsCte(n.startSearchDirectoryGroupID, cteName)
			cteWhere = append(cteWhere, fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.MetadataModelsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName))
		}

		if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			n.iamCredential,
			n.authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE,
					RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
				},
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
					RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS,
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
			cteWhere = append(cteWhere, fmt.Sprintf("%s = '%s'", intdoment.MetadataModelsRepository().DirectoryGroupsID, n.startSearchDirectoryGroupID.String()))
		}

		if iamWhereOr != nil {
			cteWhere = append(cteWhere, fmt.Sprintf("%s = TRUE", intdoment.MetadataModelsRepository().ViewAuthorized))
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

	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().ID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().ID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().ID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().DirectoryGroupsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().DirectoryGroupsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().DirectoryGroupsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().Name][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().Name, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().Name] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().Description][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().Description, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().Description] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().EditAuthorized][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().EditAuthorized, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().EditAuthorized] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().EditUnauthorized][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().EditUnauthorized, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().EditUnauthorized] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().ViewAuthorized][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().ViewAuthorized, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().ViewAuthorized] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().ViewUnauthorized][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().ViewUnauthorized, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().ViewUnauthorized] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().Tags][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().Tags, "", PROCESS_QUERY_CONDITION_AS_ARRAY, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().Tags] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().LastUpdatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().LastUpdatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().LastUpdatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.MetadataModelsRepository().DeactivatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.MetadataModelsRepository().DeactivatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.MetadataModelsRepository().DeactivatedOn] = value
		}
	}
	if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, selectQuery.TableName, "", "", "", intdoment.MetadataModelsRepository().FullTextSearch); len(value) > 0 {
		selectQuery.Where[intdoment.MetadataModelsRepository().RepositoryName] = value
	}

	directoryGroupsIDJoinDirectoryGroups := intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)
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
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryGroupsRepository().ID, true),                         //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.MetadataModelsRepository().DirectoryGroupsID, false), //2
			)

			selectQuery.Join[directoryGroupsIDJoinDirectoryGroups] = sq
		}
	}

	directoryIDJoinDirectory := intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)
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
				GetJoinColumnName(selectQuery.TableUid, intdoment.MetadataModelsRepository().DirectoryID, false), //2
			)

			selectQuery.Join[directoryIDJoinDirectory] = sq
		}
	}

	metadatamodelsJoinMetadatamodelsAuthorizationIDs := intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsRepository().RepositoryName, intdoment.MetadataModelsAuthorizationIDsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, metadatamodelsJoinMetadatamodelsAuthorizationIDs); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", metadatamodelsJoinMetadatamodelsAuthorizationIDs, err))
	} else {
		if sq, err := n.AuthorizationIDsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
			intdoment.MetadataModelsAuthorizationIDsRepository().RepositoryName,
			[]AuthIDsSelectQueryPKey{{Name: intdoment.MetadataModelsAuthorizationIDsRepository().ID, ProcessAs: PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE}},
			intdoment.MetadataModelsAuthorizationIDsRepository().CreationIamGroupAuthorizationsID,
			intdoment.MetadataModelsAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", metadatamodelsJoinMetadatamodelsAuthorizationIDs, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.MetadataModelsAuthorizationIDsRepository().ID, true), //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.MetadataModelsRepository().ID, false),       //2
			)

			selectQuery.Join[metadatamodelsJoinMetadatamodelsAuthorizationIDs] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}
