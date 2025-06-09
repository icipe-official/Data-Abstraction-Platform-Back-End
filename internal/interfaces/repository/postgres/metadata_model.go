package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
	"github.com/jackc/pgx/v5"
)

func (n *PostrgresRepository) RepoMetadataModelFindOneByDirectoryGroupID(
	ctx context.Context,
	metadataModelRepositoryName string,
	metadataMetadataModelIDColumnName string,
	metadataDirectoryGroupIDColumnName string,
	directoryGroupID uuid.UUID,
) (map[string]any, error) {
	query := fmt.Sprintf(
		"SELECT %[1]s.%[2]s FROM %[1]s INNER JOIN %[3]s ON %[3]s.%[4]s = $1 AND %[3]s.%[5]s = %[1]s.%[6]s;",
		intdoment.MetadataModelsRepository().RepositoryName, //1
		intdoment.MetadataModelsRepository().Data,           //2
		metadataModelRepositoryName,                         //3
		metadataDirectoryGroupIDColumnName,                  //4
		metadataMetadataModelIDColumnName,                   //5
		intdoment.MetadataModelsRepository().ID,             //6
	)
	n.logger.Log(ctx, slog.LevelDebug, query, "function", intlib.FunctionName(n.RepoMetadataModelFindOneByDirectoryGroupID))

	value := make(map[string]any)
	if err := n.db.QueryRow(ctx, query, directoryGroupID).Scan(&value); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		errmsg := fmt.Errorf("get %s failed, error: %v", metadataModelRepositoryName, err)
		n.logger.Log(ctx, slog.LevelDebug, errmsg.Error(), "function", intlib.FunctionName(n.RepoMetadataModelFindOneByDirectoryGroupID))
		return nil, errmsg
	}

	return value, nil
}
