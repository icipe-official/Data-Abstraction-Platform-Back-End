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

func (n *PostrgresRepository) RepoAbstractionsUpdateDirectory(
	ctx context.Context,
	authContextDirectoryGroupID uuid.UUID,
	data *intdoment.AbstractionsUpdateDirectory,
	columns []string,
) ([]*intdoment.Abstractions, error) {
	if data.NewDirectoryID == nil || data.NewDirectoryID.String() == uuid.Nil.String() {
		return nil, errors.New("NewDirectoryID empty")
	}

	abstractionsMetadataModel, err := intlib.MetadataModelGet(intdoment.AbstractionsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(abstractionsMetadataModel, intdoment.AbstractionsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.AbstractionsRepository().ID) {
		columns = append(columns, intdoment.AbstractionsRepository().ID)
	}

	selectColumns := make([]string, len(columns))
	for cIndex, cValue := range columns {
		selectColumns[cIndex] = intdoment.AbstractionsRepository().RepositoryName + "." + cValue
	}

	updateWhere := make([]string, 0)
	if len(data.AbstractionsID) > 0 {
		datumString := make([]string, len(data.AbstractionsID))
		for dIndex, d := range data.AbstractionsID {
			datumString[dIndex] = fmt.Sprintf("'%s'", d.String())
		}

		updateWhere = append(updateWhere,
			fmt.Sprintf(
				"%[1]s IN (%[2]s)",
				intdoment.AbstractionsRepository().ID, //1
				strings.Join(datumString, " , "),      //2
			),
		)
	}

	if len(updateWhere) == 0 {
		if len(data.DirectoryGroupID) > 0 {
			datumString := make([]string, len(data.DirectoryGroupID))
			for dIndex, d := range data.DirectoryGroupID {
				datumString[dIndex] = fmt.Sprintf("'%s'", d.String())
			}

			updateWhere = append(updateWhere,
				fmt.Sprintf(
					"%[1]s IN (%[2]s)",
					intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, //1
					strings.Join(datumString, " , "),                                 //2
				),
			)
		}

		if len(data.DirectoryID) > 0 {
			datumString := make([]string, len(data.DirectoryID))
			for dIndex, d := range data.DirectoryID {
				datumString[dIndex] = fmt.Sprintf("'%s'", d.String())
			}

			updateWhere = append(updateWhere,
				fmt.Sprintf(
					"%[1]s IN (%[2]s)",
					intdoment.AbstractionsRepository().DirectoryID, //1
					strings.Join(datumString, " , "),               //2
				),
			)
		}

		if len(data.StorageFilesFullTextSearch) > 0 {
			datumString := make([]string, len(data.StorageFilesFullTextSearch))
			for dIndex, d := range data.StorageFilesFullTextSearch {
				datumString[dIndex] = fmt.Sprintf("'%s'", d)
			}

			updateWhere = append(updateWhere,
				fmt.Sprintf(
					"%[1]s IN (SELECT %[2]s FROM %[3]s WHERE %[4]s @@ to_tsquery(ARRAY_TO_STRING(ARRAY[%[5]s], ' | ')))",
					intdoment.AbstractionsRepository().StorageFilesID, //1
					intdoment.StorageFilesRepository().ID,             //2
					intdoment.StorageFilesRepository().RepositoryName, //3
					intdoment.StorageFilesRepository().FullTextSearch, //4
					strings.Join(datumString, " , "),                  //5
				),
			)
		}
	}

	if data.Completed != nil {
		updateWhere = append(updateWhere, fmt.Sprintf("%s = %v", intdoment.AbstractionsRepository().Completed, *data.Completed))
	}

	if data.ReviewPass != nil {
		updateWhere = append(updateWhere, fmt.Sprintf("%s = %v", intdoment.AbstractionsRepository().ReviewPass, *data.ReviewPass))
	}

	if len(updateWhere) == 0 {
		return nil, errors.New("no condition set")
	}

	query := fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s RETURNING %[4]s;",
		intdoment.AbstractionsRepository().RepositoryName, //1
		intdoment.AbstractionsRepository().DirectoryID,    //2
		strings.Join(updateWhere, " AND "),                //3
		strings.Join(columns, " , "),                      //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsUpdateDirectory))

	rows, err := n.db.Query(ctx, query, *data.NewDirectoryID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsUpdateDirectory, fmt.Errorf("retrieve %s failed, err: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsUpdateDirectory, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(abstractionsMetadataModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsUpdateDirectory, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsUpdateDirectory, err)
	}

	abstractions := make([]*intdoment.Abstractions, 0)
	if jsonData, err := json.Marshal(array2DToObject.Objects()); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsUpdateDirectory, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing abstractions", "abstractions", string(jsonData), "function", intlib.FunctionName(n.RepoAbstractionsUpdateDirectory))
		if err := json.Unmarshal(jsonData, &abstractions); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsUpdateDirectory, err)
		}
	}

	return abstractions, nil
}

func (n *PostrgresRepository) RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups(
	ctx context.Context,
	id uuid.UUID,
	abstractionsDirectoryGroupsID uuid.UUID,
	columns []string,
) (*intdoment.Abstractions, error) {
	abstractionsMetadataModel, err := intlib.MetadataModelGet(intdoment.AbstractionsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(abstractionsMetadataModel, intdoment.AbstractionsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.AbstractionsRepository().ID) {
		columns = append(columns, intdoment.AbstractionsRepository().ID)
	}

	selectColumns := make([]string, len(columns))
	for cIndex, cValue := range columns {
		selectColumns[cIndex] = intdoment.AbstractionsRepository().RepositoryName + "." + cValue
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s INNER JOIN %[3]s ON %[2]s.%[4]s = $1 AND %[2]s.%[5]s = $2 AND %[2]s.%[5]s = %[3]s.%[6]s AND %[2]s.%[7]s IS NULL AND %[3]s.%[8]s IS NULL;",
		strings.Join(selectColumns, ","),                                    //1
		intdoment.AbstractionsRepository().RepositoryName,                   //2
		intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //3
		intdoment.AbstractionsRepository().ID,                               //4
		intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID,    //5
		intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //6
		intdoment.AbstractionsRepository().DeactivatedOn,                    //7
		intdoment.AbstractionsDirectoryGroupsRepository().DeactivatedOn,     //8
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups))

	rows, err := n.db.Query(ctx, query, id, abstractionsDirectoryGroupsID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, fmt.Errorf("retrieve %s failed, err: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(abstractionsMetadataModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.AbstractionsRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups))
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, fmt.Errorf("more than one %s found", intdoment.AbstractionsRepository().RepositoryName))
	}

	abstractions := new(intdoment.Abstractions)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing abstractions", "abstractions", string(jsonData), "function", intlib.FunctionName(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups))
		if err := json.Unmarshal(jsonData, abstractions); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups, err)
		}
	}

	return abstractions, nil
}

func (n *PostrgresRepository) RepoAbstractionsDeleteOne(
	ctx context.Context,
	iamAuthRule *intdoment.IamAuthorizationRule,
	datum *intdoment.Abstractions,
) error {
	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoAbstractionsDeleteOne, fmt.Errorf("start transaction to delete %s failed, error: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s = $1;",
		intdoment.AbstractionsAuthorizationIDsRepository().RepositoryName, //1
		intdoment.AbstractionsAuthorizationIDsRepository().ID,             //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDeleteOne))

	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		query := fmt.Sprintf(
			"DELETE FROM %[1]s WHERE %[2]s = $1;",
			intdoment.AbstractionsRepository().RepositoryName, //1
			intdoment.AbstractionsRepository().ID,             //2
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDeleteOne))
		if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
			if err := transaction.Commit(ctx); err != nil {
				return intlib.FunctionNameAndError(n.RepoAbstractionsDeleteOne, fmt.Errorf("commit transaction to delete %s failed, error: %v", intdoment.AbstractionsRepository().RepositoryName, err))
			}
			return nil
		} else {
			transaction.Rollback(ctx)
		}
	}

	transaction, err = n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoAbstractionsDeleteOne, fmt.Errorf("start transaction to deactivate %s failed, error: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s = $2;",
		intdoment.AbstractionsAuthorizationIDsRepository().RepositoryName,                       //1
		intdoment.AbstractionsAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID, //2
		intdoment.AbstractionsAuthorizationIDsRepository().ID,                                   //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDeleteOne))
	if _, err := transaction.Exec(ctx, query, iamAuthRule.ID, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoAbstractionsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.AbstractionsAuthorizationIDsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = NOW() WHERE %[3]s = $1;",
		intdoment.AbstractionsRepository().RepositoryName, //1
		intdoment.AbstractionsRepository().DeactivatedOn,  //2
		intdoment.AbstractionsRepository().ID,             //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoAbstractionsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return intlib.FunctionNameAndError(n.RepoAbstractionsDeleteOne, fmt.Errorf("commit transaction to update deactivation of %s failed, error: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoAbstractionsFindOneForDeletionByID(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	datum *intdoment.Abstractions,
	columns []string,
) (*intdoment.Abstractions, *intdoment.IamAuthorizationRule, error) {
	abstractionsMModel, err := intlib.MetadataModelGet(intdoment.AbstractionsRepository().RepositoryName)
	if err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(abstractionsMModel, intdoment.AbstractionsRepository().RepositoryName, false, false); err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.AbstractionsRepository().ID) {
		columns = append(columns, intdoment.AbstractionsRepository().ID)
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
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
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
			intdoment.AbstractionsRepository().RepositoryName, //2
			intdoment.AbstractionsRepository().ID,             //3
			intdoment.AbstractionsRepository().DirectoryID,    //4
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsFindOneForDeletionByID))

		rows, err := n.db.Query(ctx, query, datum.ID[0], iamCredential.DirectoryID[0])
		if err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.AbstractionsRepository().RepositoryName, err))
		}
		defer rows.Close()
		for rows.Next() {
			if r, err := rows.Values(); err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, err)
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
					RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
				},
			},
			iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			query := fmt.Sprintf(
				"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s = $2;",
				strings.Join(columns, " , "),                                     //1
				intdoment.AbstractionsRepository().RepositoryName,                //2
				intdoment.AbstractionsRepository().ID,                            //3
				intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, //4
			)
			n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsFindOneForDeletionByID))

			rows, err := n.db.Query(ctx, query, datum.ID[0], authContextDirectoryGroupID)
			if err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.AbstractionsRepository().RepositoryName, err))
			}
			defer rows.Close()
			for rows.Next() {
				if r, err := rows.Values(); err != nil {
					return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, err)
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
					RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
				},
			},
			iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			cteName := fmt.Sprintf("%s_%s", intdoment.AbstractionsRepository().RepositoryName, RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME)
			query := fmt.Sprintf(
				"%[1]s SELECT %[2]s FROM %[3]s WHERE %[4]s = $1 AND %[5]s;",
				RecursiveDirectoryGroupsSubGroupsCte(authContextDirectoryGroupID, cteName), //1
				strings.Join(columns, " , "),                                               //2
				intdoment.AbstractionsRepository().RepositoryName,                          //3
				intdoment.AbstractionsRepository().ID,                                      //4
				fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName), //5
			)
			n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsFindOneForDeletionByID))

			rows, err := n.db.Query(ctx, query, datum.ID[0])
			if err != nil {
				return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.AbstractionsRepository().RepositoryName, err))
			}
			defer rows.Close()
			for rows.Next() {
				if r, err := rows.Values(); err != nil {
					return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, err)
				} else {
					dataRows = append(dataRows, r)
				}
			}
			if len(dataRows) > 0 {
				iamAuthRule = iamAuthorizationRule[0]
			}
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(abstractionsMModel, nil, false, false, columns)
	if err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoAbstractionsFindOneForDeletionByID))
		return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, fmt.Errorf("more than one %s found", intdoment.AbstractionsRepository().RepositoryName))
	}

	abstraction := new(intdoment.Abstractions)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing abstraction", "abstraction", string(jsonData), "function", intlib.FunctionName(n.RepoAbstractionsFindOneForDeletionByID))
		if err := json.Unmarshal(jsonData, abstraction); err != nil {
			return nil, nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindOneForDeletionByID, err)
		}
	}

	return abstraction, iamAuthRule, nil
}

func (n *PostrgresRepository) RepoAbstractionsUpdateOne(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	datum *intdoment.Abstractions,
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
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		where = fmt.Sprintf(
			"%[1]s IN (SELECT %[2]s FROM %[3]s WHERE %[4]s = '%[5]s')",
			intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID,    //1
			intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //2
			intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //3
			intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //4
			authContextDirectoryGroupID.String(),                                //5
		)
	}
	if iamAuthorizationRule, err := n.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
			},
		},
		iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		whereQuery := fmt.Sprintf("%s = '%s'", intdoment.AbstractionsRepository().DirectoryID, iamCredential.DirectoryID[0].String())
		if len(where) > 0 {
			where += " OR " + whereQuery
		} else {
			where = whereQuery
		}
	}

	valuesToUpdate := make([]any, 0)
	valueToUpdateQuery := make([]string, 0)
	columnsToUpdate := make([]string, 0)
	nextPlaceholder := 1
	if v, vQ, c := n.RepoAbstractionsValidateAndGetColumnsAndData(ctx, authContextDirectoryGroupID, &nextPlaceholder, datum, false); len(c) == 0 || len(v) == 0 || len(vQ) == 0 {
		return intlib.NewError(http.StatusBadRequest, "no values to update")
	} else {
		valuesToUpdate = append(valuesToUpdate, v...)
		columnsToUpdate = append(columnsToUpdate, c...)
		valueToUpdateQuery = append(valueToUpdateQuery, vQ...)
	}

	query += fmt.Sprintf(
		" UPDATE %[1]s SET %[2]s WHERE %[3]s = %[4]s AND %[5]s IS NULL AND (%[6]s) AND (%[7]s) IN (SELECT %[8]s FROM %[9]s WHERE %[10]s IS NULL);",
		intdoment.AbstractionsRepository().RepositoryName,                  //1
		GetUpdateSetColumnsWithVQuery(columnsToUpdate, valueToUpdateQuery), //2
		intdoment.AbstractionsRepository().ID,                              //3
		GetandUpdateNextPlaceholder(&nextPlaceholder),                      //4
		intdoment.AbstractionsRepository().DeactivatedOn,                   //5
		where, //6
		intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID,    //7
		intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //8
		intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //9
		intdoment.AbstractionsDirectoryGroupsRepository().DeactivatedOn,     //10
	)
	query = strings.TrimLeft(query, " \n")
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsUpdateOne))

	valuesToUpdate = append(valuesToUpdate, datum.ID[0])
	if _, err := n.db.Exec(ctx, query, valuesToUpdate...); err != nil {
		return intlib.FunctionNameAndError(n.RepoAbstractionsUpdateOne, fmt.Errorf("update %s failed, err: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	return nil
}

func (n *PostrgresRepository) RepoAbstractionsInsertOne(
	ctx context.Context,
	iamAuthRule *intdoment.IamAuthorizationRule,
	directoryGroupID uuid.UUID,
	datum *intdoment.Abstractions,
	columns []string,
) (*intdoment.Abstractions, error) {
	abstractionsMetadataModel, err := intlib.MetadataModelGet(intdoment.AbstractionsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(abstractionsMetadataModel, intdoment.AbstractionsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.AbstractionsRepository().ID) {
		columns = append(columns, intdoment.AbstractionsRepository().ID)
	}

	nextPlaceholder := 1
	columnsToInsert := []string{intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, intdoment.AbstractionsRepository().DirectoryID, intdoment.AbstractionsRepository().StorageFilesID}
	valuesToInsert := []any{directoryGroupID, datum.DirectoryID[0], datum.StorageFilesID[0]}
	valueToInsertQuery := []string{GetandUpdateNextPlaceholder(&nextPlaceholder), GetandUpdateNextPlaceholder(&nextPlaceholder), GetandUpdateNextPlaceholder(&nextPlaceholder)}
	if v, vQ, c := n.RepoAbstractionsValidateAndGetColumnsAndData(ctx, directoryGroupID, &nextPlaceholder, datum, true); len(c) == 0 || len(v) == 0 || len(vQ) == 0 {
		return nil, intlib.NewError(http.StatusBadRequest, "no values to insert")
	} else {
		valuesToInsert = append(valuesToInsert, v...)
		columnsToInsert = append(columnsToInsert, c...)
		valueToInsertQuery = append(valueToInsertQuery, vQ...)
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s) VALUES(%[3]s) RETURNING %[4]s;",
		intdoment.AbstractionsRepository().RepositoryName, //1
		strings.Join(columnsToInsert, " , "),              //2
		strings.Join(valueToInsertQuery, " , "),           //3
		strings.Join(columns, " , "),                      //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsInsertOne))

	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsInsertOne, fmt.Errorf("start transaction to create %s failed, error: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	rows, err := transaction.Query(ctx, query, valuesToInsert...)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(abstractionsMetadataModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		transaction.Rollback(ctx)
		return nil, fmt.Errorf("insert %s did not return any row", intdoment.AbstractionsRepository().RepositoryName)
	}

	if len(array2DToObject.Objects()) > 1 {
		transaction.Rollback(ctx)
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoAbstractionsInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsInsertOne, fmt.Errorf("more than one %s found", intdoment.AbstractionsRepository().RepositoryName))
	}

	abstractions := new(intdoment.Abstractions)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing abstractions", "abstractions", string(jsonData), "function", intlib.FunctionName(n.RepoAbstractionsInsertOne))
		if err := json.Unmarshal(jsonData, abstractions); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsInsertOne, err)
		}
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.AbstractionsAuthorizationIDsRepository().RepositoryName,                   //1
		intdoment.AbstractionsAuthorizationIDsRepository().ID,                               //2
		intdoment.AbstractionsAuthorizationIDsRepository().CreationIamGroupAuthorizationsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsInsertOne))

	if _, err := transaction.Exec(ctx, query, abstractions.ID[0], iamAuthRule.ID); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.AbstractionsAuthorizationIDsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsInsertOne, fmt.Errorf("commit transaction to create %s failed, error: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	return abstractions, nil
}

func (n *PostrgresRepository) RepoAbstractionsValidateAndGetColumnsAndData(ctx context.Context, directoryGroupID uuid.UUID, nextPlaceholder *int, datum *intdoment.Abstractions, insert bool) ([]any, []string, []string) {
	values := make([]any, 0)
	valuesQuery := make([]string, 0)
	columns := make([]string, 0)

	fullTextSearchValue := make([]string, 0)
	if len(datum.Data) > 0 {
		if dMap, ok := datum.Data[0].(map[string]any); ok {
			columns = append(columns, intdoment.AbstractionsRepository().Data)
			values = append(values, dMap)
			valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
		} else {
			if insert {
				datum.Data[0] = map[string]any{}
				columns = append(columns, intdoment.AbstractionsRepository().Data)
				values = append(values, datum.Data[0])
				valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
			}
		}
	} else {
		if insert {
			datum.Data = []any{map[string]any{}}
			columns = append(columns, intdoment.AbstractionsRepository().Data)
			values = append(values, datum.Data[0])
			valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
		}
	}

	if slices.Contains(columns, intdoment.AbstractionsRepository().Data) {
		if mm, err := n.RepoMetadataModelFindOneByAbstractionsDirectoryGroupsID(ctx, directoryGroupID); err == nil {
			if value := MetadataModelExtractFullTextSearchValue(mm, datum.Data[0]); len(value) > 0 {
				fullTextSearchValue = append(fullTextSearchValue, value...)
			}
		}
	}

	if insert {
		if len(datum.Tags) > 0 {
			values = append(values, datum.Tags)
			columns = append(columns, intdoment.AbstractionsRepository().Tags)
			valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
			fullTextSearchValue = append(fullTextSearchValue, datum.Tags...)
		}
	} else {
		if datum.Tags != nil && len(datum.Tags) >= 0 {
			values = append(values, datum.Tags)
			columns = append(columns, intdoment.AbstractionsRepository().Tags)
			valuesQuery = append(valuesQuery, GetandUpdateNextPlaceholder(nextPlaceholder))
			if len(datum.Tags) > 0 {
				fullTextSearchValue = append(fullTextSearchValue, datum.Tags...)
			}
		}
	}

	if len(fullTextSearchValue) > 0 {
		columns = append(columns, intdoment.AbstractionsRepository().FullTextSearch)
		values = append(values, strings.Join(fullTextSearchValue, " "))
		valuesQuery = append(valuesQuery, fmt.Sprintf("to_tsvector(%s)", GetandUpdateNextPlaceholder(nextPlaceholder)))
	}

	return values, valuesQuery, columns
}

func (n *PostrgresRepository) RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID(ctx context.Context, abstractionsDirectoryGroupsID uuid.UUID, storageFilesID uuid.UUID, columns []string) ([]*intdoment.Abstractions, error) {
	abstractionsMetadataModel, err := intlib.MetadataModelGet(intdoment.AbstractionsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(abstractionsMetadataModel, intdoment.AbstractionsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.AbstractionsRepository().ID) {
		columns = append(columns, intdoment.AbstractionsRepository().ID)
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s = $2;",
		strings.Join(columns, " , "),                                     //1
		intdoment.AbstractionsRepository().RepositoryName,                //2
		intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, //3
		intdoment.AbstractionsRepository().StorageFilesID,                //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID))

	rows, err := n.db.Query(ctx, query, abstractionsDirectoryGroupsID, storageFilesID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}

	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(abstractionsMetadataModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	abstractions := make([]*intdoment.Abstractions, 0)
	if jsonData, err := json.Marshal(array2DToObject.Objects()); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing abstractions", "abstractions", string(jsonData), "function", intlib.FunctionName(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID))
		if err := json.Unmarshal(jsonData, &abstractions); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID, err)
		}
	}

	return abstractions, nil
}

func (n *PostrgresRepository) RepoAbstractionsSearch(
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
	selectQuery, err := pSelectQuery.AbstractionsGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoAbstractionsSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.AbstractionsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoAbstractionsSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoAbstractionsSearch, err)
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

func (n *PostgresSelectQuery) AbstractionsGetSelectQuery(ctx context.Context, metadataModel map[string]any, abstractionParentPath string) (*SelectQuery, error) {
	quoteColumns := true
	if len(abstractionParentPath) == 0 {
		abstractionParentPath = "$"
		quoteColumns = false
	}
	if !n.whereAfterJoin {
		quoteColumns = false
	}

	selectQuery := SelectQuery{
		TableName: intdoment.AbstractionsRepository().RepositoryName,
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
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
			},
		},
		n.iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		iamWhereOr = append(iamWhereOr, fmt.Sprintf(
			"%[1]s.%[2]s IN (SELECT %[3]s FROM %[4]s WHERE %[5]s = TRUE)",
			selectQuery.TableUid, //1
			intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID,    //2
			intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //3
			intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //4
			intdoment.AbstractionsDirectoryGroupsRepository().ViewAuthorized,    //5
		))
	}
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
			},
		},
		n.iamAuthorizationRules,
	); err == nil && iamAuthorizationRule != nil {
		if len(n.iamCredential.DirectoryID) > 0 {
			iamWhereOr = append(iamWhereOr, fmt.Sprintf("%s.%s = '%s'", selectQuery.TableUid, intdoment.AbstractionsRepository().DirectoryID, n.iamCredential.DirectoryID[0].String()))
		} else {
			iamWhereOr = append(iamWhereOr, fmt.Sprintf(
				"%[1]s.%[2]s IN (SELECT %[3]s FROM %[4]s WHERE %[5]s = TRUE)",
				selectQuery.TableUid, //1
				intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID,    //2
				intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //3
				intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //4
				intdoment.AbstractionsDirectoryGroupsRepository().ViewUnauthorized,  //5
			))
		}
	}
	if len(iamWhereOr) == 0 {
		iamWhereOr = append(iamWhereOr, fmt.Sprintf(
			"%[1]s.%[2]s IN (SELECT %[3]s FROM %[4]s WHERE %[5]s = TRUE)",
			selectQuery.TableUid, //1
			intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID,    //2
			intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //3
			intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //4
			intdoment.AbstractionsDirectoryGroupsRepository().ViewUnauthorized,  //5
		))
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
					RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
				},
			},
			n.iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			selectQuery.DirectoryGroupsSubGroupsCTEName = cteName
			selectQuery.DirectoryGroupsSubGroupsCTE = RecursiveDirectoryGroupsSubGroupsCte(n.startSearchDirectoryGroupID, cteName)
			cteWhere = append(cteWhere, fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName))
		}

		if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			n.iamCredential,
			n.authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE,
					RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
				},
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
					RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
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
			cteWhere = append(cteWhere, fmt.Sprintf("%s = '%s'", intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, n.startSearchDirectoryGroupID.String()))
		}

		if iamWhereOr != nil {
			cteWhere = append(cteWhere, fmt.Sprintf(
				"%[1]s.%[2]s IN (SELECT %[3]s FROM %[4]s WHERE %[5]s = TRUE)",
				selectQuery.TableUid, //1
				intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID,    //2
				intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //3
				intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //4
				intdoment.AbstractionsDirectoryGroupsRepository().ViewAuthorized,    //5
			))

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

	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().ID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().ID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().ID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().DirectoryID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().DirectoryID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().DirectoryID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().StorageFilesID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().StorageFilesID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().StorageFilesID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().Data][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().Data, "", PROCESS_QUERY_CONDITION_AS_JSONB, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().Data] = value
		}
	}
	if fgKeyString, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().Data][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().Data, fgKeyString, PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().Data] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().Completed][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().Completed, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().Completed] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().ReviewPass][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().ReviewPass, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().ReviewPass] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().Tags][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().Tags, "", PROCESS_QUERY_CONDITION_AS_ARRAY, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().Tags] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().LastUpdatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().LastUpdatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().LastUpdatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.AbstractionsRepository().DeactivatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.AbstractionsRepository().DeactivatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.AbstractionsRepository().DeactivatedOn] = value
		}
	}
	if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, selectQuery.TableName, "", "", "", intdoment.AbstractionsRepository().FullTextSearch); len(value) > 0 {
		selectQuery.Where[intdoment.AbstractionsRepository().RepositoryName] = value
	}

	abstractionsDirectoryGroupsIDJoinAbstractionsDirectoryGroups := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, abstractionsDirectoryGroupsIDJoinAbstractionsDirectoryGroups); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", abstractionsDirectoryGroupsIDJoinAbstractionsDirectoryGroups, err))
	} else {
		if sq, err := n.AbstractionsDirectoryGroupsGetSelectQuery(
			ctx,
			value,
			abstractionParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", abstractionsDirectoryGroupsIDJoinAbstractionsDirectoryGroups, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, true),        //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, false), //2
			)

			selectQuery.Join[abstractionsDirectoryGroupsIDJoinAbstractionsDirectoryGroups] = sq
		}
	}

	directoryIDJoinDirectory := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)
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
				GetJoinColumnName(sq.TableUid, intdoment.DirectoryRepository().ID, true),                       //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsRepository().DirectoryID, false), //2
			)

			selectQuery.Join[directoryIDJoinDirectory] = sq
		}
	}

	storageFilesIDJoinStorageFiles := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().StorageFilesID, intdoment.StorageFilesRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, storageFilesIDJoinStorageFiles); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", storageFilesIDJoinStorageFiles, err))
	} else {
		if sq, err := n.StorageFilesGetSelectQuery(
			ctx,
			value,
			abstractionParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", storageFilesIDJoinStorageFiles, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.StorageFilesRepository().ID, true),                       //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsRepository().StorageFilesID, false), //2
			)

			selectQuery.Join[storageFilesIDJoinStorageFiles] = sq
		}
	}

	abstractionsJoinAbstractionsAuthorizationIDs := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().RepositoryName, intdoment.AbstractionsAuthorizationIDsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, abstractionsJoinAbstractionsAuthorizationIDs); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", abstractionsJoinAbstractionsAuthorizationIDs, err))
	} else {
		if sq, err := n.AuthorizationIDsGetSelectQuery(
			ctx,
			value,
			abstractionParentPath,
			intdoment.AbstractionsAuthorizationIDsRepository().RepositoryName,
			[]AuthIDsSelectQueryPKey{{Name: intdoment.AbstractionsAuthorizationIDsRepository().ID, ProcessAs: PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE}},
			intdoment.AbstractionsAuthorizationIDsRepository().CreationIamGroupAuthorizationsID,
			intdoment.AbstractionsAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", abstractionsJoinAbstractionsAuthorizationIDs, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.AbstractionsAuthorizationIDsRepository().ID, true), //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.AbstractionsRepository().ID, false),       //2
			)

			selectQuery.Join[abstractionsJoinAbstractionsAuthorizationIDs] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}

func (n *PostrgresRepository) RepoMetadataModelFindOneByAbstractionsDirectoryGroupsID(ctx context.Context, directoryGroupID uuid.UUID) (map[string]any, error) {
	query := fmt.Sprintf(
		"SELECT %[1]s.%[2]s FROM %[1]s INNER JOIN %[3]s ON %[3]s.%[4]s = $1 AND %[3]s.%[5]s = %[1]s.%[6]s;",
		intdoment.MetadataModelsRepository().RepositoryName,                 //1
		intdoment.MetadataModelsRepository().Data,                           //2
		intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName,    //3
		intdoment.AbstractionsDirectoryGroupsRepository().DirectoryGroupsID, //4
		intdoment.AbstractionsDirectoryGroupsRepository().MetadataModelsID,  //5
		intdoment.MetadataModelsRepository().ID,                             //6
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelFindOneByAbstractionsDirectoryGroupsID))
	value := make(map[string]any)
	if err := n.db.QueryRow(ctx, query, directoryGroupID).Scan(&value); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		errmsg := fmt.Errorf("get %s failed, error: %v", intdoment.MetadataModelsRepository().RepositoryName, err)
		n.logger.Log(ctx, slog.LevelDebug, errmsg.Error(), "function", intlib.FunctionName(n.RepoMetadataModelFindOneByAbstractionsDirectoryGroupsID))
		return nil, errmsg
	}

	return value, nil
}
