package groups

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intwebservice "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/web_service"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *service) ServiceDirectoryGroupsDeleteMany(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	verboseResponse bool,
	data []*intdoment.DirectoryGroups,
) (int, *intdoment.MetadataModelVerbRes, error) {
	verbres := new(intdoment.MetadataModelVerbRes)
	verbres.MetadataModelVerboseResponse = new(intdoment.MetadataModelVerboseResponse)
	if verboseResponse {
		if d, err := intlib.MetadataModelMiscGet(intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsDeleteMany, err).Error())
			return 0, nil, intlib.NewError(http.StatusInternalServerError, fmt.Sprintf("Get %v metadata-model failed", intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE))
		} else {
			verbres.MetadataModelVerboseResponse.MetadataModel = d
		}
	}
	verbres.MetadataModelVerboseResponse.Data = make([]*intdoment.MetadataModelVerboseResponseData, 0)

	successful := 0
	failed := 0
	for _, datum := range data {
		verbRes := new(intdoment.MetadataModelVerboseResponseData)

		if len(datum.ID) > 0 {
			iamAuthorizationRule := new(intdoment.IamAuthorizationRule)
			if iar, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
				ctx,
				iamCredential,
				authContextDirectoryGroupID,
				[]*intdoment.IamGroupAuthorizationRule{
					{
						ID:        intdoment.AUTH_RULE_DELETE,
						RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY_GROUPS,
					},
				},
				iamAuthorizationRules,
			); err != nil {
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), "get iam auth rule failed", err.Error()}
				failed += 1
				goto appendNewVerboseResponse
			} else if iar == nil {
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusForbidden}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusForbidden)}
				failed += 1
				goto appendNewVerboseResponse
			} else {
				iamAuthorizationRule = iar[0]
			}

			if datum.ID[0].String() != authContextDirectoryGroupID.String() {
				if value, err := n.repo.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID(ctx, authContextDirectoryGroupID, datum.ID[0]); err != nil {
					n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsDeleteMany, err).Error())
					verbRes.Data = []any{datum}
					verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
					verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
					verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), fmt.Sprintf("validate %s failed", intdoment.DirectoryGroupsSubGroupsRepository().RepositoryName), err.Error()}
					failed += 1
					goto appendNewVerboseResponse
				} else {
					if value == nil {
						verbRes.Data = []any{datum}
						verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
						verbRes.Status[0].StatusCode = []int{http.StatusForbidden}
						verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusForbidden)}
						failed += 1
						goto appendNewVerboseResponse
					}
				}
			}

			if value, err := n.repo.RepoDirectoryGroupsCheckIfSystemGroup(ctx, datum.ID[0]); err != nil {
				n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsDeleteMany, err).Error())
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), fmt.Sprintf("validate non-system %s failed", intdoment.DirectoryGroupsRepository().RepositoryName), err.Error()}
				failed += 1
				goto appendNewVerboseResponse
			} else {
				if value {
					verbRes.Data = []any{datum}
					verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
					verbRes.Status[0].StatusCode = []int{http.StatusBadRequest}
					verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusBadRequest), fmt.Sprintf("%s is a system group", intdoment.DirectoryGroupsRepository().RepositoryName)}
					failed += 1
					goto appendNewVerboseResponse
				}
			}

			if err := n.repo.RepoDirectoryGroupsDeleteOne(ctx, iamAuthorizationRule, datum); err != nil {
				n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsDeleteMany, err).Error())
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), "delete/deactivate failed", err.Error()}
				failed += 1
				goto appendNewVerboseResponse
			} else {
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusOK}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusOK), "delete/deactivate successful"}
				successful += 1
				goto appendNewVerboseResponse
			}
		} else {
			verbRes.Data = []any{datum}
			verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
			verbRes.Status[0].StatusCode = []int{http.StatusBadRequest}
			verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusBadRequest), "data is not valid"}
			failed += 1
		}
	appendNewVerboseResponse:
		verbres.MetadataModelVerboseResponse.Data = append(verbres.MetadataModelVerboseResponse.Data, verbRes)
	}

	verbres.Message = fmt.Sprintf("Delete/Deactivate %[1]s: %[2]d/%[4]d successful and %[3]d/%[4]d failed", intdoment.DirectoryGroupsRepository().RepositoryName, successful, failed, len(data))
	verbres.Successful = successful
	verbres.Failed = failed

	return http.StatusOK, verbres, nil
}

func (n *service) ServiceDirectoryGroupsUpdateMany(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	fieldAnyMetadataModelGet intdomint.FieldAnyMetadataModel,
	verboseResponse bool,
	data []*intdoment.DirectoryGroups,
) (int, *intdoment.MetadataModelVerbRes, error) {
	verbres := new(intdoment.MetadataModelVerbRes)
	verbres.MetadataModelVerboseResponse = new(intdoment.MetadataModelVerboseResponse)
	if verboseResponse {
		if d, err := intlib.MetadataModelMiscGet(intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsUpdateMany, err).Error())
			return 0, nil, intlib.NewError(http.StatusInternalServerError, fmt.Sprintf("Get %v metadata-model failed", intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE))
		} else {
			verbres.MetadataModelVerboseResponse.MetadataModel = d
		}
	}
	verbres.MetadataModelVerboseResponse.Data = make([]*intdoment.MetadataModelVerboseResponseData, 0)

	successful := 0
	failed := 0
	for _, datum := range data {
		verbRes := new(intdoment.MetadataModelVerboseResponseData)

		if len(datum.ID) > 0 {
			if datum.ID[0].String() != authContextDirectoryGroupID.String() {
				if iar, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
					ctx,
					iamCredential,
					authContextDirectoryGroupID,
					[]*intdoment.IamGroupAuthorizationRule{
						{
							ID:        intdoment.AUTH_RULE_UPDATE_OTHERS,
							RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY_GROUPS,
						},
					},
					iamAuthorizationRules,
				); err != nil {
					verbRes.Data = []any{datum}
					verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
					verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
					verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), "get iam auth rule failed", err.Error()}
					failed += 1
					goto appendNewVerboseResponse
				} else {
					if iar == nil {
						verbRes.Data = []any{datum}
						verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
						verbRes.Status[0].StatusCode = []int{http.StatusForbidden}
						verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusForbidden)}
						failed += 1
						goto appendNewVerboseResponse
					}
				}

				if value, err := n.repo.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID(ctx, authContextDirectoryGroupID, datum.ID[0]); err != nil {
					n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsUpdateMany, err).Error())
					verbRes.Data = []any{datum}
					verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
					verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
					verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), fmt.Sprintf("validate %s failed", intdoment.DirectoryGroupsSubGroupsRepository().RepositoryName), err.Error()}
					failed += 1
					goto appendNewVerboseResponse
				} else {
					if value == nil {
						verbRes.Data = []any{datum}
						verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
						verbRes.Status[0].StatusCode = []int{http.StatusForbidden}
						verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusForbidden)}
						failed += 1
						goto appendNewVerboseResponse
					}
				}
			} else {
				if iar, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
					ctx,
					iamCredential,
					authContextDirectoryGroupID,
					[]*intdoment.IamGroupAuthorizationRule{
						{
							ID:        intdoment.AUTH_RULE_UPDATE,
							RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY_GROUPS,
						},
					},
					iamAuthorizationRules,
				); err != nil {
					verbRes.Data = []any{datum}
					verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
					verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
					verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), "get iam auth rule failed", err.Error()}
					failed += 1
					goto appendNewVerboseResponse
				} else {
					if iar == nil {
						verbRes.Data = []any{datum}
						verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
						verbRes.Status[0].StatusCode = []int{http.StatusForbidden}
						verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusForbidden)}
						failed += 1
						goto appendNewVerboseResponse
					}
				}
			}

			if err := n.repo.RepoDirectoryGroupsUpdateOne(ctx, datum, fieldAnyMetadataModelGet); err != nil {
				n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsInsertMany, err).Error())
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), "update failed", err.Error()}
				failed += 1
				goto appendNewVerboseResponse
			} else {
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusOK}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusOK), "update successful"}
				successful += 1
				goto appendNewVerboseResponse
			}
		} else {
			verbRes.Data = []any{datum}
			verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
			verbRes.Status[0].StatusCode = []int{http.StatusBadRequest}
			verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusBadRequest), "data is not valid"}
			failed += 1
		}
	appendNewVerboseResponse:
		verbres.MetadataModelVerboseResponse.Data = append(verbres.MetadataModelVerboseResponse.Data, verbRes)
	}

	verbres.Message = fmt.Sprintf("Update %[1]s: %[2]d/%[4]d successful and %[3]d/%[4]d failed", intdoment.DirectoryGroupsRepository().RepositoryName, successful, failed, len(data))
	verbres.Successful = successful
	verbres.Failed = failed

	return http.StatusOK, verbres, nil
}

func (n *service) ServiceDirectoryGroupsInsertMany(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	fieldAnyMetadataModelGet intdomint.FieldAnyMetadataModel,
	verboseResponse bool,
	data []*intdoment.DirectoryGroups,
) (int, *intdoment.MetadataModelVerbRes, error) {
	verbres := new(intdoment.MetadataModelVerbRes)
	verbres.MetadataModelVerboseResponse = new(intdoment.MetadataModelVerboseResponse)
	if verboseResponse {
		if d, err := intlib.MetadataModelMiscGet(intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsInsertMany, err).Error())
			return 0, nil, intlib.NewError(http.StatusInternalServerError, fmt.Sprintf("Get %v metadata-model failed", intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE))
		} else {
			verbres.MetadataModelVerboseResponse.MetadataModel = d
		}
	}
	verbres.MetadataModelVerboseResponse.Data = make([]*intdoment.MetadataModelVerboseResponseData, 0)

	successful := 0
	failed := 0
	for _, datum := range data {
		verbRes := new(intdoment.MetadataModelVerboseResponseData)

		iamAuthorizationRule := new(intdoment.IamAuthorizationRule)
		if iar, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
			ctx,
			iamCredential,
			authContextDirectoryGroupID,
			[]*intdoment.IamGroupAuthorizationRule{
				{
					ID:        intdoment.AUTH_RULE_CREATE,
					RuleGroup: intdoment.AUTH_RULE_GROUP_DIRECTORY_GROUPS,
				},
			},
			iamAuthorizationRules,
		); err != nil {
			verbRes.Data = []any{datum}
			verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
			verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
			verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), "get iam auth rule failed", err.Error()}
			failed += 1
			goto appendNewVerboseResponse
		} else if iar == nil {
			verbRes.Data = []any{datum}
			verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
			verbRes.Status[0].StatusCode = []int{http.StatusForbidden}
			verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusForbidden)}
			failed += 1
			goto appendNewVerboseResponse
		} else {
			iamAuthorizationRule = iar[0]
		}

		if value, err := n.repo.RepoDirectoryGroupsInsertOne(ctx, datum, authContextDirectoryGroupID, iamAuthorizationRule, fieldAnyMetadataModelGet, nil); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsInsertMany, err).Error())
			verbRes.Data = []any{datum}
			verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
			verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
			verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), "insert failed", err.Error()}
			failed += 1
			goto appendNewVerboseResponse
		} else {
			verbRes.Data = []any{value}
			verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
			verbRes.Status[0].StatusCode = []int{http.StatusOK}
			verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusOK), "creation successful"}
			successful += 1
			goto appendNewVerboseResponse
		}
	appendNewVerboseResponse:
		verbres.MetadataModelVerboseResponse.Data = append(verbres.MetadataModelVerboseResponse.Data, verbRes)
	}

	verbres.Message = fmt.Sprintf("Create %[1]s: %[2]d/%[4]d successful and %[3]d/%[4]d failed", intdoment.DirectoryGroupsRepository().RepositoryName, successful, failed, len(data))
	verbres.Successful = successful
	verbres.Failed = failed

	return http.StatusOK, verbres, nil
}

func (n *service) ServiceIamGroupAuthorizationsGetAuthorized(
	ctx context.Context,
	iamAuthInfo *intdoment.IamCredentials,
	authContextDirectoryGroupID uuid.UUID,
	groupAuthorizationRules []*intdoment.IamGroupAuthorizationRule,
	currentIamAuthorizationRules *intdoment.IamAuthorizationRules,
) ([]*intdoment.IamAuthorizationRule, error) {
	return n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamAuthInfo,
		authContextDirectoryGroupID,
		groupAuthorizationRules,
		currentIamAuthorizationRules,
	)
}

func (n *service) ServiceDirectoryGroupsSearch(
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
	if value, err := n.repo.RepoDirectoryGroupsSearch(
		ctx,
		mmsearch,
		repo,
		iamCredential,
		iamAuthorizationRules,
		startSearchDirectoryGroupID,
		authContextDirectoryGroupID,
		skipIfFGDisabled,
		skipIfDataExtraction,
		whereAfterJoin,
	); err != nil {
		n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsSearch, err).Error())
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		return value, nil
	}
}

func (n *service) ServiceDirectoryGroupsGetMetadataModel(ctx context.Context, metadataModelRetrieve intdomint.MetadataModelRetrieve, targetJoinDepth int) (map[string]any, error) {
	if value, err := metadataModelRetrieve.DirectoryGroupsGetMetadataModel(ctx, 0, targetJoinDepth, nil); err != nil {
		n.logger.Log(ctx, slog.LevelWarn+1, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsGetMetadataModel, err).Error())
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		return value, nil
	}
}

func (n *service) ServiceDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID) (*intdoment.DirectoryGroups, error) {
	var directoryGroup *intdoment.DirectoryGroups
	if value, err := n.repo.RepoDirectoryGroupsFindOneByIamCredentialID(ctx, iamCredentialID, nil); err != nil {
		n.logger.Log(ctx, slog.LevelWarn+1, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsFindOneByIamCredentialID, err).Error())
	} else {
		directoryGroup = value
	}

	if directoryGroup == nil {
		if dg, err := n.repo.RepoDirectoryGroupsFindSystemGroup(ctx, nil); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceDirectoryGroupsFindOneByIamCredentialID, err).Error())
			return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		} else {
			directoryGroup = dg
		}
	}

	return directoryGroup, nil
}

type service struct {
	repo   intdomint.RouteDirectoryGroupsRepository
	logger intdomint.Logger
}

func NewService(webService *intwebservice.WebService) (*service, error) {
	n := new(service)

	n.repo = webService.PostgresRepository
	n.logger = webService.Logger

	if n.logger == nil {
		return n, errors.New("webService.Logger is empty")
	}

	if n.repo == nil {
		return n, errors.New("webService.PostgresRepository is empty")
	}

	return n, nil
}
