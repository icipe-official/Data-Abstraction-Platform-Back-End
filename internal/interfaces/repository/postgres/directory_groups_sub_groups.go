package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
	intlibmmodel "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib/metadata_model"
)

func (n *PostrgresRepository) RepoDirectoryGroupsSubGroupsFindOneBySubGroupID(ctx context.Context, parentGroupID uuid.UUID, subGroupID uuid.UUID) (*intdoment.DirectoryGroupsSubGroups, error) {
	directoryGroupsSubGroupsMModel, err := intlib.MetadataModelGet(intdoment.DirectoryGroupsSubGroupsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID, err)
	}

	columns := []string{intdoment.DirectoryGroupsSubGroupsRepository().ParentGroupID, intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID}
	query := RecursiveDirectoryGroupsSubGroupsCte(parentGroupID, RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME)
	query += fmt.Sprintf(
		" SELECT %[1]s FROM %[2]s WHERE %[3]s = $1;",
		strings.Join(columns, "  , "),                             //1
		RECURSIVE_DIRECTORY_GROUPS_DEFAULT_CTE_NAME,               //2
		intdoment.DirectoryGroupsSubGroupsRepository().SubGroupID, //3
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID))

	rows, err := n.db.Query(ctx, query, subGroupID)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID, fmt.Errorf("retrieve %s failed, err: %v", intdoment.DirectoryGroupsSubGroupsRepository().RepositoryName, err))
	}
	defer rows.Close()
	dataRows := make([]any, 0)
	for rows.Next() {
		if r, err := rows.Values(); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID, err)
		} else {
			dataRows = append(dataRows, r)
		}
	}

	array2DToObject, err := intlibmmodel.NewConvert2DArrayToObjects(directoryGroupsSubGroupsMModel, nil, false, false, columns)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID, err)
	}
	if err := array2DToObject.Convert(dataRows); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID, err)
	}

	if len(array2DToObject.Objects()) == 0 {
		return nil, nil
	}

	if len(array2DToObject.Objects()) > 1 {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("length of array2DToObject.Objects(): %v", len(array2DToObject.Objects())), "function", intlib.FunctionName(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID))
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID, fmt.Errorf("more than one %s found", intdoment.DirectoryGroupsSubGroupsRepository().RepositoryName))
	}

	directoryGroupSubGroup := new(intdoment.DirectoryGroupsSubGroups)
	if jsonData, err := json.Marshal(array2DToObject.Objects()[0]); err != nil {
		return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID, err)
	} else {
		n.logger.Log(ctx, slog.LevelDebug, "json parsing directoryGroupSubGroup", "directoryGroupSubGroup", string(jsonData), "function", intlib.FunctionName(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID))
		if err := json.Unmarshal(jsonData, directoryGroupSubGroup); err != nil {
			return nil, intlib.FunctionNameAndError(n.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID, err)
		}
	}

	return directoryGroupSubGroup, nil
}
