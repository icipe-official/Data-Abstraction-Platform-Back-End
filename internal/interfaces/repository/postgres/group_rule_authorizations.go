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

func (n *PostrgresRepository) RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID(ctx context.Context, iamGroupAuthorizationID uuid.UUID, columns []string) (*intdoment.GroupRuleAuthorizations, error) {
	groupRuleAuthorizationsMModel, err := intlib.MetadataModelGet(intdoment.GroupRuleAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(groupRuleAuthorizationsMModel, intdoment.GroupRuleAuthorizationsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.GroupRuleAuthorizationsRepository().ID) {
		columns = append(columns, intdoment.GroupRuleAuthorizationsRepository().ID)
	}

	selectColumns := make([]string, len(columns))
	for cIndex, cValue := range columns {
		selectColumns[cIndex] = intdoment.GroupRuleAuthorizationsRepository().RepositoryName + "." + cValue
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s INNER JOIN %[3]s ON %[3]s.%[4]s = $1 AND %[3]s.%[5]s = %[2]s.%[6]s;",
		strings.Join(selectColumns, "  , "),                                    //1
		intdoment.GroupRuleAuthorizationsRepository().RepositoryName,           //2
		intdoment.IamGroupAuthorizationsRepository().RepositoryName,            //3
		intdoment.IamGroupAuthorizationsRepository().ID,                        //4
		intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID, //5
		intdoment.GroupRuleAuthorizationsRepository().ID,                       //6
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID))

	rows, err := n.db.Query(ctx, query, iamGroupAuthorizationID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(groupRuleAuthorizationsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID))
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID, fmt.Errorf("more than one %s found", intdoment.GroupRuleAuthorizationsRepository().RepositoryName))
	}

	groupRuleAuthorization := new(intdoment.GroupRuleAuthorizations)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing groupRuleAuthorization", "groupRuleAuthorization", string(jsonData), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID))
		if err := json.Unmarshal(jsonData, groupRuleAuthorization); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID, err)
		}
	}

	return groupRuleAuthorization, nil
}

func (n *PostrgresRepository) RepoGroupRuleAuthorizationsFindActiveOneByID(ctx context.Context, groupRuleAuthorizationID uuid.UUID, columns []string) (*intdoment.GroupRuleAuthorizations, error) {
	groupRuleAuthorizationsMModel, err := intlib.MetadataModelGet(intdoment.GroupRuleAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindActiveOneByID, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(groupRuleAuthorizationsMModel, intdoment.GroupRuleAuthorizationsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindActiveOneByID, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.GroupRuleAuthorizationsRepository().ID) {
		columns = append(columns, intdoment.GroupRuleAuthorizationsRepository().ID)
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s IS NULL;",
		strings.Join(columns, "  , "),                                //1
		intdoment.GroupRuleAuthorizationsRepository().RepositoryName, //2
		intdoment.GroupRuleAuthorizationsRepository().ID,             //3
		intdoment.GroupRuleAuthorizationsRepository().DeactivatedOn,  //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindActiveOneByID))

	rows, err := n.db.Query(ctx, query, groupRuleAuthorizationID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindActiveOneByID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindActiveOneByID, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(groupRuleAuthorizationsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindActiveOneByID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindActiveOneByID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindActiveOneByID))
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindActiveOneByID, fmt.Errorf("more than one %s found", intdoment.GroupRuleAuthorizationsRepository().RepositoryName))
	}

	groupRuleAuthorization := new(intdoment.GroupRuleAuthorizations)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindActiveOneByID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing groupRuleAuthorization", "groupRuleAuthorization", string(jsonData), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindActiveOneByID))
		if err := json.Unmarshal(jsonData, groupRuleAuthorization); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindActiveOneByID, err)
		}
	}

	return groupRuleAuthorization, nil
}

func (n *PostrgresRepository) RepoGroupRuleAuthorizationsDeleteOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.GroupRuleAuthorizations) error {
	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsDeleteOne, fmt.Errorf("start transaction to delete %s failed, error: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"DELETE FROM %[1]s WHERE %[2]s = $1;",
		intdoment.GroupRuleAuthorizationsIDsRepository().RepositoryName, //1
		intdoment.GroupRuleAuthorizationsIDsRepository().ID,             //2
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsDeleteOne))

	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		query := fmt.Sprintf(
			"DELETE FROM %[1]s WHERE %[2]s = $1;",
			intdoment.GroupRuleAuthorizationsRepository().RepositoryName, //1
			intdoment.GroupRuleAuthorizationsRepository().ID,             //2
		)
		n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsDeleteOne))
		if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
			if err := transaction.Commit(ctx); err != nil {
				return intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsDeleteOne, fmt.Errorf("commit transaction to delete %s failed, error: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
			}
			return nil
		} else {
			transaction.Rollback(ctx)
		}
	}

	transaction, err = n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsDeleteOne, fmt.Errorf("start transaction to deactivate %s failed, error: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = $1 WHERE %[3]s = $2;",
		intdoment.GroupRuleAuthorizationsIDsRepository().RepositoryName,                       //1
		intdoment.GroupRuleAuthorizationsIDsRepository().DeactivationIamGroupAuthorizationsID, //2
		intdoment.GroupRuleAuthorizationsIDsRepository().ID,                                   //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsDeleteOne))
	if _, err := transaction.Exec(ctx, query, iamAuthRule.ID, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.GroupRuleAuthorizationsIDsRepository().RepositoryName, err))
	}

	query = fmt.Sprintf(
		"UPDATE %[1]s SET %[2]s = NOW() WHERE %[3]s = $1;",
		intdoment.GroupRuleAuthorizationsRepository().RepositoryName, //1
		intdoment.GroupRuleAuthorizationsRepository().DeactivatedOn,  //2
		intdoment.GroupRuleAuthorizationsRepository().ID,             //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsDeleteOne))
	if _, err := transaction.Exec(ctx, query, datum.ID[0]); err == nil {
		transaction.Rollback(ctx)
		return intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsDeleteOne, fmt.Errorf("update %s failed, err: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsDeleteOne, fmt.Errorf("commit transaction to update deactivation of %s failed, error: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}

	//TODO: Add deactivate IamGroupAuthorizations

	return nil
}

func (n *PostrgresRepository) RepoGroupRuleAuthorizationsFindOneInactiveRule(ctx context.Context, groupRuleAuthorizationID uuid.UUID, columns []string) (*intdoment.GroupRuleAuthorizations, error) {
	groupRuleAuthorizationsMModel, err := intlib.MetadataModelGet(intdoment.GroupRuleAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneInactiveRule, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(groupRuleAuthorizationsMModel, intdoment.GroupRuleAuthorizationsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneInactiveRule, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.GroupRuleAuthorizationsRepository().ID) {
		columns = append(columns, intdoment.GroupRuleAuthorizationsRepository().ID)
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s IS NOT NULL;",
		strings.Join(columns, "  , "),                                //1
		intdoment.GroupRuleAuthorizationsRepository().RepositoryName, //2
		intdoment.GroupRuleAuthorizationsRepository().ID,             //3
		intdoment.GroupRuleAuthorizationsRepository().DeactivatedOn,  //4
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindOneInactiveRule))

	rows, err := n.db.Query(ctx, query, groupRuleAuthorizationID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneInactiveRule, fmt.Errorf("retrieve %s failed, err: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneInactiveRule, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(groupRuleAuthorizationsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneInactiveRule, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneInactiveRule, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindOneInactiveRule))
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneInactiveRule, fmt.Errorf("more than one %s found", intdoment.GroupRuleAuthorizationsRepository().RepositoryName))
	}

	groupRuleAuthorization := new(intdoment.GroupRuleAuthorizations)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneInactiveRule, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing groupRuleAuthorization", "groupRuleAuthorization", string(jsonData), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindOneInactiveRule))
		if err := json.Unmarshal(jsonData, groupRuleAuthorization); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneInactiveRule, err)
		}
	}

	return groupRuleAuthorization, nil
}

func (n *PostrgresRepository) RepoGroupRuleAuthorizationsInsertOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.GroupRuleAuthorizations, columns []string) (*intdoment.GroupRuleAuthorizations, error) {
	groupRuleAuthorizationsMModel, err := intlib.MetadataModelGet(intdoment.GroupRuleAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(groupRuleAuthorizationsMModel, intdoment.GroupRuleAuthorizationsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.GroupRuleAuthorizationsRepository().ID) {
		columns = append(columns, intdoment.GroupRuleAuthorizationsRepository().ID)
	}

	transaction, err := n.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, fmt.Errorf("start transaction to create %s failed, error: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}

	query := fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s, %[4]s) VALUES ($1, $2, $3) RETURNING %[5]s;",
		intdoment.GroupRuleAuthorizationsRepository().RepositoryName,               //1
		intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID,            //2
		intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID,    //3
		intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup, //4
		strings.Join(columns, "  , "),                                              //5
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsInsertOne))

	rows, err := transaction.Query(ctx, query, datum.DirectoryGroupsID[0], datum.GroupAuthorizationRuleID[0].GroupAuthorizationRulesID[0], datum.GroupAuthorizationRuleID[0].GroupAuthorizationRulesGroup[0])
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(groupRuleAuthorizationsMModel, nil, false, false, columns)
	if err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		transaction.Rollback(ctx)
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, fmt.Errorf("no %s found", intdoment.GroupRuleAuthorizationsRepository().RepositoryName))
	}

	if len(array2DToObject.Objects()) > 1 {
		transaction.Rollback(ctx)
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsInsertOne))
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, fmt.Errorf("more than one %s found", intdoment.GroupRuleAuthorizationsRepository().RepositoryName))
	}

	groupRuleAuthorization := new(intdoment.GroupRuleAuthorizations)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing groupRuleAuthorization", "groupRuleAuthorization", string(jsonData), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsInsertOne))
		if err := json.Unmarshal(jsonData, groupRuleAuthorization); err != nil {
			transaction.Rollback(ctx)
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, err)
		}
	}

	query = fmt.Sprintf(
		"INSERT INTO %[1]s (%[2]s, %[3]s) VALUES ($1, $2);",
		intdoment.GroupRuleAuthorizationsIDsRepository().RepositoryName,                   //1
		intdoment.GroupRuleAuthorizationsIDsRepository().ID,                               //2
		intdoment.GroupRuleAuthorizationsIDsRepository().CreationIamGroupAuthorizationsID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsInsertOne))

	if _, err := transaction.Exec(ctx, query, groupRuleAuthorization.ID[0], iamAuthRule.ID); err != nil {
		transaction.Rollback(ctx)
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, fmt.Errorf("insert %s failed, err: %v", intdoment.GroupRuleAuthorizationsIDsRepository().RepositoryName, err))
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsInsertOne, fmt.Errorf("commit transaction to create %s failed, error: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}

	return groupRuleAuthorization, nil
}

func (n *PostrgresRepository) RepoGroupRuleAuthorizationsFindOneActiveRule(ctx context.Context, directoryGroupID uuid.UUID, groupAuthorizationRuleID string, groupAuthorizationRuleGroup string, columns []string) (*intdoment.GroupRuleAuthorizations, error) {
	groupRuleAuthorizationsMModel, err := intlib.MetadataModelGet(intdoment.GroupRuleAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneActiveRule, err)
	}

	if len(columns) == 0 {
		if dbColumnFields, err := intlibmmodel.DatabaseGetColumnFields(groupRuleAuthorizationsMModel, intdoment.GroupRuleAuthorizationsRepository().RepositoryName, false, false); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneActiveRule, err)
		} else {
			columns = dbColumnFields.ColumnFieldsReadOrder
		}
	}

	if !slices.Contains(columns, intdoment.GroupRuleAuthorizationsRepository().ID) {
		columns = append(columns, intdoment.GroupRuleAuthorizationsRepository().ID)
	}

	query := fmt.Sprintf(
		"SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s = $2 AND %[5]s = $3 AND %[6]s IS NULL;",
		strings.Join(columns, "  , "),                                              //1
		intdoment.GroupRuleAuthorizationsRepository().RepositoryName,               //2
		intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID,            //3
		intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID,    //4
		intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup, //5
		intdoment.GroupRuleAuthorizationsRepository().DeactivatedOn,                //6
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindOneActiveRule))

	rows, err := n.db.Query(ctx, query, directoryGroupID, groupAuthorizationRuleID, groupAuthorizationRuleGroup)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneActiveRule, fmt.Errorf("retrieve %s failed, err: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneActiveRule, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(groupRuleAuthorizationsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneActiveRule, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneActiveRule, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindOneActiveRule))
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneActiveRule, fmt.Errorf("more than one %s found", intdoment.GroupRuleAuthorizationsRepository().RepositoryName))
	}

	groupRuleAuthorization := new(intdoment.GroupRuleAuthorizations)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneActiveRule, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing groupRuleAuthorization", "groupRuleAuthorization", string(jsonData), "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsFindOneActiveRule))
		if err := json.Unmarshal(jsonData, groupRuleAuthorization); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsFindOneActiveRule, err)
		}
	}

	return groupRuleAuthorization, nil
}

func (n *PostrgresRepository) RepoGroupRuleAuthorizationsSearch(
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
	selectQuery, err := pSelectQuery.GroupRuleAuthorizationsGetSelectQuery(ctx, mmsearch.MetadataModel, "")
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsSearch, err)
	}

	query, selectQueryExtract := GetSelectQuery(selectQuery, whereAfterJoin)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoGroupRuleAuthorizationsSearch))

	rows, err := n.db.Query(ctx, query)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsSearch, fmt.Errorf("retrieve %s failed, err: %v", intdoment.GroupRuleAuthorizationsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsSearch, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(mmsearch.MetadataModel, selectQueryExtract.Fields, false, false, nil)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsSearch, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoGroupRuleAuthorizationsSearch, err)
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

func (n *PostgresSelectQuery) GroupRuleAuthorizationsGetSelectQuery(ctx context.Context, metadataModel map[string]any, metadataModelParentPath string) (*SelectQuery, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_GROUP_RULE_AUTHORIZATIONS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_GROUP_RULE_AUTHORIZATIONS,
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
		TableName: intdoment.GroupRuleAuthorizationsRepository().RepositoryName,
		Query:     "",
		Where:     make(map[string]map[int][][]string),
		Join:      make(map[string]*SelectQuery),
		JoinQuery: make([]string, 0),
	}

	if tableUid, ok := metadataModel[intlibmmodel.FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID].(string); ok && len(tableUid) > 0 {
		selectQuery.TableUid = tableUid
	} else {
		return nil, intlib.FunctionNameAndError(n.GroupRuleAuthorizationsGetSelectQuery, errors.New("tableUid is empty"))
	}

	if value, err := intlibmmodel.DatabaseGetColumnFields(metadataModel, selectQuery.TableUid, false, false); err != nil {
		return nil, intlib.FunctionNameAndError(n.GroupRuleAuthorizationsGetSelectQuery, fmt.Errorf("extract database column fields failed, error: %v", err))
	} else {
		selectQuery.Columns = value
	}

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
					RuleGroup: intdoment.AUTH_RULE_GROUP_GROUP_RULE_AUTHORIZATIONS,
				},
			},
			n.iamAuthorizationRules,
		); err == nil && iamAuthorizationRule != nil {
			selectQuery.DirectoryGroupsSubGroupsCTEName = cteName
			selectQuery.DirectoryGroupsSubGroupsCTE = RecursiveDirectoryGroupsSubGroupsCte(n.startSearchDirectoryGroupID, cteName)
			cteWhere = append(cteWhere, fmt.Sprintf("(%s) IN (SELECT %s FROM %s)", intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, cteName))
		}

		if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			n.iamCredential,
			n.authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_RETRIEVE,
					RuleGroup: intdoment.AUTH_RULE_GROUP_GROUP_RULE_AUTHORIZATIONS,
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
			cteWhere = append(cteWhere, fmt.Sprintf("%s = '%s'", intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID, n.startSearchDirectoryGroupID.String()))
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

	if _, ok := selectQuery.Columns.Fields[intdoment.GroupRuleAuthorizationsRepository().ID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.GroupRuleAuthorizationsRepository().ID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.GroupRuleAuthorizationsRepository().ID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.GroupRuleAuthorizationsRepository().CreatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.GroupRuleAuthorizationsRepository().CreatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.GroupRuleAuthorizationsRepository().CreatedOn] = value
		}
	}
	if _, ok := selectQuery.Columns.Fields[intdoment.GroupRuleAuthorizationsRepository().DeactivatedOn][intlibmmodel.FIELD_GROUP_PROP_FIELD_GROUP_KEY].(string); ok {
		if value := n.getWhereCondition(quoteColumns, selectQuery.TableUid, "", intdoment.GroupRuleAuthorizationsRepository().DeactivatedOn, "", PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE, ""); len(value) > 0 {
			selectQuery.Where[intdoment.GroupRuleAuthorizationsRepository().DeactivatedOn] = value
		}
	}

	directoryGroupsIDJoinDirectoryGroups := intlib.MetadataModelGenJoinKey(intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)
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
				GetJoinColumnName(selectQuery.TableUid, intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID, false), //2
			)

			selectQuery.Join[directoryGroupsIDJoinDirectoryGroups] = sq
		}
	}

	groupRuleAuthorizationsJoinGroupAuthorizationRules := intlib.MetadataModelGenJoinKey(intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupAuthorizationRulesRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, groupRuleAuthorizationsJoinGroupAuthorizationRules); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", groupRuleAuthorizationsJoinGroupAuthorizationRules, err))
	} else {
		if sq, err := n.GroupAuthorizationRulesGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", groupRuleAuthorizationsJoinGroupAuthorizationRules, err))
		} else {
			sq.JoinType = JOIN_INNER
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s AND %[3]s = %[4]s",
				GetJoinColumnName(sq.TableUid, intdoment.GroupAuthorizationRulesRepository().ID, true),                                     //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID, false),    //2
				GetJoinColumnName(sq.TableUid, intdoment.GroupAuthorizationRulesRepository().RuleGroup, true),                              //3
				GetJoinColumnName(selectQuery.TableUid, intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup, false), //4
			)

			selectQuery.Join[groupRuleAuthorizationsJoinGroupAuthorizationRules] = sq
		}
	}

	groupRuleAuthorizationsJoinGroupRuleAuthorizationIDs := intlib.MetadataModelGenJoinKey(intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupRuleAuthorizationsIDsRepository().RepositoryName)
	if value, err := n.extractChildMetadataModel(metadataModel, groupRuleAuthorizationsJoinGroupRuleAuthorizationIDs); err != nil {
		n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("extract %s child metadata model failed, error: %v", groupRuleAuthorizationsJoinGroupRuleAuthorizationIDs, err))
	} else {
		if sq, err := n.AuthorizationIDsGetSelectQuery(
			ctx,
			value,
			metadataModelParentPath,
			intdoment.GroupRuleAuthorizationsIDsRepository().RepositoryName,
			[]AuthIDsSelectQueryPKey{{Name: intdoment.GroupRuleAuthorizationsIDsRepository().ID, ProcessAs: PROCESS_QUERY_CONDITION_AS_SINGLE_VALUE}},
			intdoment.GroupRuleAuthorizationsIDsRepository().CreationIamGroupAuthorizationsID,
			intdoment.GroupRuleAuthorizationsIDsRepository().DeactivationIamGroupAuthorizationsID,
		); err != nil {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("get child %s psql query failed, error: %v", groupRuleAuthorizationsJoinGroupRuleAuthorizationIDs, err))
		} else {
			if len(sq.Where) == 0 {
				sq.JoinType = JOIN_LEFT
			} else {
				sq.JoinType = JOIN_INNER
			}
			sq.JoinQuery = make([]string, 1)
			sq.JoinQuery[0] = fmt.Sprintf(
				"%[1]s = %[2]s",
				GetJoinColumnName(sq.TableUid, intdoment.GroupRuleAuthorizationsIDsRepository().ID, true),        //1
				GetJoinColumnName(selectQuery.TableUid, intdoment.GroupRuleAuthorizationsRepository().ID, false), //2
			)

			selectQuery.Join[groupRuleAuthorizationsJoinGroupRuleAuthorizationIDs] = sq
		}
	}

	selectQuery.appendSort()
	selectQuery.appendLimitOffset(metadataModel)
	selectQuery.appendDistinct()

	return &selectQuery, nil
}
