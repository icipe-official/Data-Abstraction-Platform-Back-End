package metadatamodelretrieve

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *MetadataModelRetrieve) IamGroupAuthorizationsGetMetadataModel(ctx context.Context, currentJoinDepth int, targetJoinDepth int, skipJoin map[string]bool) (map[string]any, error) {
	if iamAuthorizationRule, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		n.iamCredential,
		n.authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_SELF,
				RuleGroup: intdoment.AUTH_RULE_GROUP_IAM_GROUP_AUTHORIZATIONS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE,
				RuleGroup: intdoment.AUTH_RULE_GROUP_IAM_GROUP_AUTHORIZATIONS,
			},
			{
				ID:        intdoment.AUTH_RULE_RETRIEVE_OTHERS,
				RuleGroup: intdoment.AUTH_RULE_GROUP_IAM_GROUP_AUTHORIZATIONS,
			},
		},
		n.iamAuthorizationRules,
	); err != nil || iamAuthorizationRule == nil {
		return nil, intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}

	parentMetadataModel, err := n.GetMetadataModel(intdoment.IamGroupAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.IamGroupAuthorizationsGetMetadataModel, err)
	}

	parentMetadataModel, err = n.SetTableCollectionUidAndJoinDepthForMetadataModel(parentMetadataModel, intdoment.IamGroupAuthorizationsRepository().RepositoryName, currentJoinDepth)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.IamGroupAuthorizationsGetMetadataModel, err)
	}

	if currentJoinDepth < targetJoinDepth || targetJoinDepth < 0 {
		if skipJoin == nil {
			skipJoin = make(map[string]bool)
		}

		newTargetJoinDepth := 0
		if targetJoinDepth-currentJoinDepth+1 > 2 {
			newTargetJoinDepth = targetJoinDepth + 1
		} else if targetJoinDepth < 0 {
			newTargetJoinDepth = currentJoinDepth + 3
		} else {
			newTargetJoinDepth = targetJoinDepth
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.IamGroupAuthorizationsRepository().IamCredentialsID, intdoment.IamCredentialsRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.IamGroupAuthorizationsRepository().IamCredentialsID, intdoment.IamCredentialsRepository().RepositoryName)
			if childMetadataModel, err := n.IamCredentialsGetMetadataModel(
				ctx,
				currentJoinDepth+1,
				newTargetJoinDepth,
				nil,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.IamGroupAuthorizationsGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
					parentMetadataModel,
					childMetadataModel,
					intdoment.IamGroupAuthorizationsRepository().IamCredentialsID,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.IamGroupAuthorizationsRepository().IamCredentialsID},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.IamGroupAuthorizationsGetMetadataModel, err)
				}
			}
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID, intdoment.GroupRuleAuthorizationsRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID, intdoment.GroupRuleAuthorizationsRepository().RepositoryName)
			if childMetadataModel, err := n.GroupRuleAuthorizationsGetMetadataModel(
				ctx,
				currentJoinDepth+1,
				newTargetJoinDepth,
				nil,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.IamGroupAuthorizationsGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
					parentMetadataModel,
					childMetadataModel,
					intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.IamGroupAuthorizationsRepository().GroupRuleAuthorizationsID},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.IamGroupAuthorizationsGetMetadataModel, err)
				}
			}
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.IamGroupAuthorizationsRepository().RepositoryName, intdoment.IamGroupAuthorizationsIDsRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.IamGroupAuthorizationsRepository().RepositoryName, intdoment.IamGroupAuthorizationsIDsRepository().RepositoryName)
			if childMetadataModel, err := n.DefaultAuthorizationIDsGetMetadataModel(
				ctx,
				intdoment.IamGroupAuthorizationsIDsRepository().RepositoryName,
				currentJoinDepth+1,
				newTargetJoinDepth,
				nil,
				intdoment.IamGroupAuthorizationsIDsRepository().CreationIamGroupAuthorizationsID,
				intdoment.IamGroupAuthorizationsIDsRepository().DeactivationIamGroupAuthorizationsID,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.IamGroupAuthorizationsGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(parentMetadataModel, childMetadataModel, "", false, newChildMetadataModelfgSuffix, nil)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.IamGroupAuthorizationsGetMetadataModel, err)
				}
			}
		}
	}

	return parentMetadataModel, nil
}
