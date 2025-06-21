package metadatamodelretrieve

import (
	"context"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *MetadataModelRetrieve) StorageFilesTemporaryGetMetadataModel(ctx context.Context, currentJoinDepth int, targetJoinDepth int, skipJoin map[string]bool) (map[string]any, error) {
	parentMetadataModel, err := n.GetMetadataModel(intdoment.StorageFilesTemporaryRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.StorageFilesGetMetadataModel, err)
	}

	parentMetadataModel, err = n.SetTableCollectionUidAndJoinDepthForMetadataModel(parentMetadataModel, intdoment.StorageFilesRepository().RepositoryName, currentJoinDepth)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.StorageFilesGetMetadataModel, err)
	}

	// if currentJoinDepth < targetJoinDepth || targetJoinDepth < 0 {
	// 	if skipJoin == nil {
	// 		skipJoin = make(map[string]bool)
	// 	}

	// 	if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.StorageFilesRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)]; !ok || !skipMMJoin {
	// 		newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.StorageFilesRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)
	// 		if childMetadataModel, err := n.DirectoryGroupsGetMetadataModel(
	// 			ctx,
	// 			currentJoinDepth+1,
	// 			targetJoinDepth,
	// 			nil,
	// 		); err != nil {
	// 			n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.StorageFilesGetMetadataModel))
	// 		} else {
	// 			parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
	// 				parentMetadataModel,
	// 				childMetadataModel,
	// 				intdoment.StorageFilesRepository().DirectoryGroupsID,
	// 				false,
	// 				newChildMetadataModelfgSuffix,
	// 				[]string{intdoment.StorageFilesRepository().DirectoryGroupsID, intdoment.StorageFilesRepository().DirectoryGroupsID},
	// 			)
	// 			if err != nil {
	// 				return nil, intlib.FunctionNameAndError(n.StorageFilesGetMetadataModel, err)
	// 			}
	// 		}
	// 	}

	// 	if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.StorageFilesRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)]; !ok || !skipMMJoin {
	// 		newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.StorageFilesRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)
	// 		if childMetadataModel, err := n.DirectoryGetMetadataModel(
	// 			ctx,
	// 			currentJoinDepth+1,
	// 			targetJoinDepth,
	// 			nil,
	// 		); err != nil {
	// 			n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.StorageFilesGetMetadataModel))
	// 		} else {
	// 			parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
	// 				parentMetadataModel,
	// 				childMetadataModel,
	// 				intdoment.StorageFilesRepository().DirectoryID,
	// 				false,
	// 				newChildMetadataModelfgSuffix,
	// 				[]string{intdoment.StorageFilesRepository().DirectoryID},
	// 			)
	// 			if err != nil {
	// 				return nil, intlib.FunctionNameAndError(n.StorageFilesGetMetadataModel, err)
	// 			}
	// 		}
	// 	}

	// 	if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.StorageFilesRepository().RepositoryName, intdoment.StorageFilesAuthorizationIDsRepository().RepositoryName)]; !ok || !skipMMJoin {
	// 		newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.StorageFilesRepository().RepositoryName, intdoment.StorageFilesAuthorizationIDsRepository().RepositoryName)
	// 		if childMetadataModel, err := n.DefaultAuthorizationIDsGetMetadataModel(
	// 			ctx,
	// 			intdoment.StorageFilesAuthorizationIDsRepository().RepositoryName,
	// 			currentJoinDepth+1,
	// 			targetJoinDepth,
	// 			nil,
	// 			intdoment.StorageFilesAuthorizationIDsRepository().CreationIamGroupAuthorizationsID,
	// 			intdoment.StorageFilesAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID,
	// 		); err != nil {
	// 			n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.StorageFilesGetMetadataModel))
	// 		} else {
	// 			parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(parentMetadataModel, childMetadataModel, "", false, newChildMetadataModelfgSuffix, nil)
	// 			if err != nil {
	// 				return nil, intlib.FunctionNameAndError(n.StorageFilesGetMetadataModel, err)
	// 			}
	// 		}
	// 	}
	// }

	return parentMetadataModel, nil
}
