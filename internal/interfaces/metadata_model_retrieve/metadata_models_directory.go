package metadatamodelretrieve

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *MetadataModelRetrieve) MetadataModelsDirectoryGetMetadataModel(ctx context.Context, currentJoinDepth int, targetJoinDepth int, skipJoin map[string]bool) (map[string]any, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_METADATA_MODELS_DIRECTORY,
			},
		},
		n.iamAuthorizationRules,
	); err != nil || iamAuthorizationRule == nil {
		return nil, intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}

	parentMetadataModel, err := n.GetMetadataModel(intdoment.MetadataModelsDirectoryRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.MetadataModelsDirectoryGetMetadataModel, err)
	}

	parentMetadataModel, err = n.SetTableCollectionUidAndJoinDepthForMetadataModel(parentMetadataModel, intdoment.MetadataModelsDirectoryRepository().RepositoryName, currentJoinDepth)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.MetadataModelsDirectoryGetMetadataModel, err)
	}

	if currentJoinDepth < targetJoinDepth || targetJoinDepth < 0 {
		if skipJoin == nil {
			skipJoin = make(map[string]bool)
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsDirectoryRepository().MetadataModelsID, intdoment.MetadataModelsRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsDirectoryRepository().MetadataModelsID, intdoment.MetadataModelsRepository().RepositoryName)
			if childMetadataModel, err := n.MetadataModelsGetMetadataModel(
				ctx,
				currentJoinDepth+1,
				targetJoinDepth,
				nil,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.MetadataModelsDirectoryGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
					parentMetadataModel,
					childMetadataModel,
					intdoment.MetadataModelsDirectoryRepository().MetadataModelsID,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.MetadataModelsDirectoryRepository().MetadataModelsID},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.MetadataModelsDirectoryGetMetadataModel, err)
				}
			}
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)
			if childMetadataModel, err := n.DirectoryGroupsGetMetadataModel(
				ctx,
				currentJoinDepth+1,
				targetJoinDepth,
				nil,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.MetadataModelsDirectoryGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
					parentMetadataModel,
					childMetadataModel,
					intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.MetadataModelsDirectoryGetMetadataModel, err)
				}
			}
		}
	}

	return parentMetadataModel, nil
}
