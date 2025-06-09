package metadatamodelretrieve

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *MetadataModelRetrieve) AbstractionsReviewsCommentsGetMetadataModel(ctx context.Context, currentJoinDepth int, targetJoinDepth int, skipJoin map[string]bool) (map[string]any, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        "",
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS_REVIEWS,
			},
		},
		n.iamAuthorizationRules,
	); err != nil || iamAuthorizationRule == nil {
		return nil, intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}

	parentMetadataModel, err := n.GetMetadataModel(intdoment.AbstractionsReviewsCommentsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.AbstractionsReviewsCommentsGetMetadataModel, err)
	}

	parentMetadataModel, err = n.SetTableCollectionUidAndJoinDepthForMetadataModel(parentMetadataModel, intdoment.AbstractionsReviewsCommentsRepository().RepositoryName, currentJoinDepth)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.AbstractionsReviewsCommentsGetMetadataModel, err)
	}

	if currentJoinDepth < targetJoinDepth || targetJoinDepth < 0 {
		if skipJoin == nil {
			skipJoin = make(map[string]bool)
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID, intdoment.AbstractionsRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID, intdoment.AbstractionsRepository().RepositoryName)
			if childMetadataModel, err := n.AbstractionsGetMetadataModel(
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
					intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.AbstractionsReviewsCommentsRepository().AbstractionsID},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.AbstractionsReviewsCommentsGetMetadataModel, err)
				}
			}
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.AbstractionsReviewsCommentsRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.AbstractionsReviewsCommentsRepository().DirectoryID, intdoment.DirectoryRepository().RepositoryName)
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
					intdoment.AbstractionsReviewsCommentsRepository().DirectoryID,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.AbstractionsReviewsCommentsRepository().DirectoryID},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.AbstractionsReviewsCommentsGetMetadataModel, err)
				}
			}
		}
	}

	return parentMetadataModel, nil
}
