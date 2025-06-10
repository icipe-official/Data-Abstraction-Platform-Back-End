package abstractions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	inthttp "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *service) ServiceAbstractionsUpdateDirectory(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	authContextDirectoryGroupID uuid.UUID,
	verboseResponse bool,
	data *intdoment.AbstractionsUpdateDirectory,
) (int, *intdoment.MetadataModelVerbRes, error) {
	if iar, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
		ctx,
		iamCredential,
		authContextDirectoryGroupID,
		[]*intdoment.IamGroupAuthorizationRule{
			{
				ID:        intdoment.AUTH_RULE_UPDATE_DIRECTORY,
				RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS,
			},
		},
		nil,
	); err != nil {
		return http.StatusInternalServerError, nil, errors.New("get iam auth rule failed " + err.Error())
	} else {
		if iar == nil {
			return http.StatusForbidden, nil, errors.New(http.StatusText(http.StatusForbidden))
		}
	}

	verbres := new(intdoment.MetadataModelVerbRes)
	verbres.MetadataModelVerboseResponse = new(intdoment.MetadataModelVerboseResponse)
	if verboseResponse {
		if d, err := intlib.MetadataModelMiscGet(intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsUpdateMany, err).Error())
			return 0, nil, intlib.NewError(http.StatusInternalServerError, fmt.Sprintf("Get %v metadata-model failed", intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE))
		} else {
			verbres.MetadataModelVerboseResponse.MetadataModel = d
		}
	}
	verbres.MetadataModelVerboseResponse.Data = make([]*intdoment.MetadataModelVerboseResponseData, 0)

	if res, err := n.repo.RepoAbstractionsUpdateDirectory(ctx, authContextDirectoryGroupID, data, nil); err != nil {
		n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsUpdateDirectory, err).Error())
		return http.StatusInternalServerError, nil, fmt.Errorf("update %s %s failed, error: %v", intdoment.AbstractionsRepository().RepositoryName, intdoment.AbstractionsRepository().DirectoryID, err)
	} else {
		verbres.Message = fmt.Sprintf("update %s %s executed", intdoment.AbstractionsRepository().RepositoryName, intdoment.AbstractionsRepository().DirectoryID)
		verbres.Successful = len(res)
		for _, value := range res {
			verbres.MetadataModelVerboseResponse.Data = append(verbres.MetadataModelVerboseResponse.Data, &intdoment.MetadataModelVerboseResponseData{
				Status: []intdoment.MetadataModelVerboseResponseStatus{{
					StatusCode:    []int{http.StatusOK},
					StatusMessage: []string{http.StatusText(http.StatusOK)},
				}},
				Data: []any{value},
			})
		}
	}

	return http.StatusOK, verbres, nil
}

func (n *service) ServiceAbstractionsDeleteMany(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	verboseResponse bool,
	data []*intdoment.Abstractions,
) (int, *intdoment.MetadataModelVerbRes, error) {
	verbres := new(intdoment.MetadataModelVerbRes)
	verbres.MetadataModelVerboseResponse = new(intdoment.MetadataModelVerboseResponse)
	if verboseResponse {
		if d, err := intlib.MetadataModelMiscGet(intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsDeleteMany, err).Error())
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
			mm, iamAuthorizationRule, err := n.repo.RepoAbstractionsFindOneForDeletionByID(ctx, iamCredential, iamAuthorizationRules, authContextDirectoryGroupID, datum, nil)
			if err != nil {
				n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsDeleteMany, err).Error())
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), fmt.Sprintf("get %s failed", intdoment.AbstractionsRepository().RepositoryName), err.Error()}
				failed += 1
				goto appendNewVerboseResponse
			}

			if mm == nil || iamAuthorizationRule == nil {
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusNotFound}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusNotFound), "Content not found or not authorized to delete content"}
				failed += 1
				goto appendNewVerboseResponse
			}

			if err := n.repo.RepoAbstractionsDeleteOne(ctx, iamAuthorizationRule, mm); err != nil {
				n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsDeleteMany, err).Error())
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

	verbres.Message = fmt.Sprintf("Delete/Deactivate %[1]s: %[2]d/%[4]d successful and %[3]d/%[4]d failed", intdoment.AbstractionsRepository().RepositoryName, successful, failed, len(data))
	verbres.Successful = successful
	verbres.Failed = failed

	return http.StatusOK, verbres, nil
}

func (n *service) ServiceAbstractionsUpdateMany(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	verboseResponse bool,
	data []*intdoment.Abstractions,
) (int, *intdoment.MetadataModelVerbRes, error) {
	verbres := new(intdoment.MetadataModelVerbRes)
	verbres.MetadataModelVerboseResponse = new(intdoment.MetadataModelVerboseResponse)
	if verboseResponse {
		if d, err := intlib.MetadataModelMiscGet(intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsUpdateMany, err).Error())
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
			if err := n.repo.RepoAbstractionsUpdateOne(ctx, iamCredential, iamAuthorizationRules, authContextDirectoryGroupID, datum); err != nil {
				n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsUpdateMany, err).Error())
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

	verbres.Message = fmt.Sprintf("Update %[1]s: %[2]d/%[4]d successful and %[3]d/%[4]d failed", intdoment.AbstractionsRepository().RepositoryName, successful, failed, len(data))
	verbres.Successful = successful
	verbres.Failed = failed

	return http.StatusOK, verbres, nil
}

func (n *service) ServiceAbstractionsInsertMany(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	verboseResponse bool,
	data []*intdoment.Abstractions,
	doNotSkipIFAbstractionWithFileIDExists bool,
) (int, *intdoment.MetadataModelVerbRes, error) {
	verbres := new(intdoment.MetadataModelVerbRes)
	verbres.MetadataModelVerboseResponse = new(intdoment.MetadataModelVerboseResponse)
	if verboseResponse {
		if d, err := intlib.MetadataModelMiscGet(intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsInsertMany, err).Error())
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

		if len(datum.DirectoryID) > 0 && len(datum.StorageFilesID) > 0 {
			iamAuthorizationRule := new(intdoment.IamAuthorizationRule)
			if len(iamCredential.DirectoryID) > 0 && iamCredential.DirectoryID[0].String() == datum.DirectoryID[0].String() {
				if iar, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
					ctx,
					iamCredential,
					authContextDirectoryGroupID,
					[]*intdoment.IamGroupAuthorizationRule{
						{
							ID:        intdoment.AUTH_RULE_CREATE,
							RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS_DIRECTORY_GROUPS,
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
			} else {
				if iar, err := n.repo.RepoIamGroupAuthorizationsGetAuthorized(
					ctx,
					iamCredential,
					authContextDirectoryGroupID,
					[]*intdoment.IamGroupAuthorizationRule{
						{
							ID:        intdoment.AUTH_RULE_CREATE_OTHERS,
							RuleGroup: intdoment.AUTH_RULE_GROUP_ABSTRACTIONS_DIRECTORY_GROUPS,
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
			}

			if !doNotSkipIFAbstractionWithFileIDExists {
				if value, err := n.repo.RepoAbstractionsFindManyByAbstractionsDirectoryGroupsIDAndStorageFilesID(ctx, authContextDirectoryGroupID, datum.StorageFilesID[0], nil); err != nil {
					n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsInsertMany, err).Error())
					verbRes.Data = []any{datum}
					verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
					verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
					verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), fmt.Sprintf("validate existance %s failed", intdoment.AbstractionsRepository().RepositoryName), err.Error()}
					failed += 1
					goto appendNewVerboseResponse
				} else {
					if value != nil {
						verbRes.Data = []any{datum}
						verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
						verbRes.Status[0].StatusCode = []int{http.StatusBadRequest}
						verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusBadRequest), fmt.Sprintf("%s already exists", intdoment.AbstractionsRepository().RepositoryName)}
						failed += 1
						goto appendNewVerboseResponse
					}
				}
			}

			if value, err := n.repo.RepoAbstractionsInsertOne(ctx, iamAuthorizationRule, authContextDirectoryGroupID, datum, nil); err != nil {
				n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsInsertMany, err).Error())
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

	verbres.Message = fmt.Sprintf("Create %[1]s: %[2]d/%[4]d successful and %[3]d/%[4]d failed", intdoment.AbstractionsRepository().RepositoryName, successful, failed, len(data))
	verbres.Successful = successful
	verbres.Failed = failed

	return http.StatusOK, verbres, nil
}

func (n *service) ServiceAbstractionsSearch(
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
	if value, err := n.repo.RepoAbstractionsSearch(
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
		n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceAbstractionsSearch, err).Error())
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		return value, nil
	}
}

func (n *service) ServiceAbstractionsGetMetadataModel(ctx context.Context, metadataModelRetrieve intdomint.MetadataModelRetrieve, targetJoinDepth int) (map[string]any, error) {
	if value, err := metadataModelRetrieve.AbstractionsGetMetadataModel(ctx, 0, targetJoinDepth, nil); err != nil {
		n.logger.Log(ctx, slog.LevelWarn+1, intlib.FunctionNameAndError(n.ServiceAbstractionsGetMetadataModel, err).Error())
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		return value, nil
	}
}

func (n *service) ServiceAbstractionsMetadataModelGet(ctx context.Context, directoryGroupID uuid.UUID) (map[string]any, error) {
	if value, err := n.repo.RepoMetadataModelFindOneByAbstractionsDirectoryGroupsID(ctx, directoryGroupID); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("Get metadata-model failed, error: %v", err), intlib.FunctionName(n.ServiceAbstractionsMetadataModelGet))
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		if value == nil {
			return nil, intlib.NewError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}
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
	repo   intdomint.RouteAbstractionsRepository
	logger intdomint.Logger
}

func NewService(webService *inthttp.WebService) (*service, error) {
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

const param_DO_NOT_SKIP_IF_ABSTRACTION_WITH_FILE_ID_EXISTS string = "do_not_skip_if_abstraction_with_file_id_exists"
