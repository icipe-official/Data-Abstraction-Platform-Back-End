package metadatamodelretrieve

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *MetadataModelRetrieve) GroupRuleAuthorizationsGetMetadataModel(ctx context.Context, currentJoinDepth int, targetJoinDepth int, skipJoin map[string]bool) (map[string]any, error) {
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

	parentMetadataModel, err := n.GetMetadataModel(intdoment.GroupRuleAuthorizationsRepository().RepositoryName)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.GroupRuleAuthorizationsGetMetadataModel, err)
	}

	parentMetadataModel, err = n.SetTableCollectionUidAndJoinDepthForMetadataModel(parentMetadataModel, intdoment.GroupRuleAuthorizationsRepository().RepositoryName, currentJoinDepth)
	if err != nil {
		return nil, intlib.FunctionNameAndError(n.GroupRuleAuthorizationsGetMetadataModel, err)
	}

	if currentJoinDepth < targetJoinDepth || targetJoinDepth < 0 {
		if skipJoin == nil {
			skipJoin = make(map[string]bool)
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID, intdoment.DirectoryGroupsRepository().RepositoryName)
			if childMetadataModel, err := n.DirectoryGroupsGetMetadataModel(
				ctx,
				currentJoinDepth+1,
				targetJoinDepth,
				nil,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.GroupRuleAuthorizationsGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
					parentMetadataModel,
					childMetadataModel,
					intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.GroupRuleAuthorizationsRepository().DirectoryGroupsID},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.GroupRuleAuthorizationsGetMetadataModel, err)
				}
			}
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupAuthorizationRulesRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupAuthorizationRulesRepository().RepositoryName)
			if childMetadataModel, err := n.GroupAuthorizationRulesGetMetadataModel(ctx, currentJoinDepth+1); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.GroupRuleAuthorizationsGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(
					parentMetadataModel,
					childMetadataModel,
					intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup,
					false,
					newChildMetadataModelfgSuffix,
					[]string{intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleID, intdoment.GroupRuleAuthorizationsRepository().GroupAuthorizationsRuleGroup},
				)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.GroupRuleAuthorizationsGetMetadataModel, err)
				}
			}
		}

		if skipMMJoin, ok := skipJoin[intlib.MetadataModelGenJoinKey(intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupRuleAuthorizationsIDsRepository().RepositoryName)]; !ok || !skipMMJoin {
			newChildMetadataModelfgSuffix := intlib.MetadataModelGenJoinKey(intdoment.GroupRuleAuthorizationsRepository().RepositoryName, intdoment.GroupRuleAuthorizationsIDsRepository().RepositoryName)
			if childMetadataModel, err := n.DefaultAuthorizationIDsGetMetadataModel(
				ctx,
				intdoment.GroupRuleAuthorizationsIDsRepository().RepositoryName,
				currentJoinDepth+1,
				targetJoinDepth,
				nil,
				intdoment.GroupRuleAuthorizationsIDsRepository().CreationIamGroupAuthorizationsID,
				intdoment.GroupRuleAuthorizationsIDsRepository().DeactivationIamGroupAuthorizationsID,
			); err != nil {
				n.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("setup %s failed, err: %v", newChildMetadataModelfgSuffix, err), "function", intlib.FunctionName(n.GroupRuleAuthorizationsGetMetadataModel))
			} else {
				parentMetadataModel, err = n.MetadataModelInsertChildIntoParent(parentMetadataModel, childMetadataModel, "", false, newChildMetadataModelfgSuffix, nil)
				if err != nil {
					return nil, intlib.FunctionNameAndError(n.GroupRuleAuthorizationsGetMetadataModel, err)
				}
			}
		}
	}

	return parentMetadataModel, nil
}
