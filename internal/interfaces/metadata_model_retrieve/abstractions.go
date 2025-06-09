package metadatamodelretrieve

import (
	"context"
	"fmt"
	"log/slog"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *MetadataModelRetrieve) AbstractionsGetMetadataModel(ctx context.Context, currentJoinDepth int, targetJoinDepth int, skipJoin map[string]bool) (map[string]any, error) {
	parentMetadataModel, err := n.GetMetadataModel(intdoment.AbstractionsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.AbstractionsGetMetadataModel, err)
	}

	parentMetadataModel, err = n.SetTableCollectionUidAndJoinDepthForMetadataModel(parentMetadataModel, intdoment.AbstractionsRepository().RepositoryName, currentJoinDepth)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.AbstractionsGetMetadataModel, err)
	}

	if currentJoinDepth < targetJoinDepth || targetJoinDepth < 0 {
		if skipJoin == nil {
			skipJoin = make(map[string]bool)
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID, intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName)
			if childMetadataModel, err := n.AbstractionsDirectoryGroupsGetMetadataModel(
				ctx,
				currentJoinDepth+1,
				targetJoinDepth,
				nil,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.AbstractionsGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
					parentMetadataModel,
					childMetadataModel,
					intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.AbstractionsRepository().AbstractionsDirectoryGroupsID},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.AbstractionsGetMetadataModel, err)
				}
			}
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)
			if childMetadataModel, err := n.DirectoryGetMetadataModel(
				ctx,
				currentJoinDepth+1,
				targetJoinDepth,
				nil,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.AbstractionsGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
					parentMetadataModel,
					childMetadataModel,
					intdoment.AbstractionsRepository().DirectoryID,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.AbstractionsRepository().DirectoryID},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.AbstractionsGetMetadataModel, err)
				}
			}
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().StorageFilesID, intdoment.StorageFilesRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().StorageFilesID, intdoment.StorageFilesRepository().RepositoryName)
			if childMetadataModel, err := n.StorageFilesGetMetadataModel(
				ctx,
				currentJoinDepth+1,
				targetJoinDepth,
				nil,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.AbstractionsGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
					parentMetadataModel,
					childMetadataModel,
					intdoment.AbstractionsRepository().StorageFilesID,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.AbstractionsRepository().StorageFilesID},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.AbstractionsGetMetadataModel, err)
				}
			}
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().RepositoryName, intdoment.AbstractionsAuthorizationIDsRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsRepository().RepositoryName, intdoment.AbstractionsAuthorizationIDsRepository().RepositoryName)
			if childMetadataModel, err := n.DefaultAuthorizationIDsGetMetadataModel(
				ctx,
				intdoment.AbstractionsAuthorizationIDsRepository().RepositoryName,
				currentJoinDepth+1,
				targetJoinDepth,
				nil,
				intdoment.AbstractionsAuthorizationIDsRepository().CreationIamGroupAuthorizationsID,
				intdoment.AbstractionsAuthorizationIDsRepository().DeactivationIamGroupAuthorizationsID,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.AbstractionsGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(parentMetadataModel, childMetadataModel, "", false, newChildMetadataModelfgSuffix, nil)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.AbstractionsGetMetadataModel, err)
				}
			}
		}
	}

	return parentMetadataModel, nil
}
