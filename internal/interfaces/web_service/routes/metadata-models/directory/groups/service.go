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

func (n *service) ServiceMetadataModelsDirectoryGroupsDeleteMany(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	verboseResponse bool,
	data []*intdoment.MetadataModelsDirectoryGroups,
) (int, *intdoment.MetadataModelVerbRes, error) {
	verbres := new(intdoment.MetadataModelVerbRes)
	verbres.MetadataModelVerboseResponse = new(intdoment.MetadataModelVerboseResponse)
	if verboseResponse {
		if d, err := intlib.MetadataModelMiscGet(intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsInsertMany, err).Error())
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

		if len(datum.DirectoryGroupsID) > 0 && len(datum.MetadataModelsID) > 0 {
			if authContextDirectoryGroupID.String() != datum.DirectoryGroupsID[0].String() {
				if value, err := n.repo.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID(ctx, authContextDirectoryGroupID, datum.DirectoryGroupsID[0]); err != nil {
					n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsInsertMany, err).Error())
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

			if err := n.repo.RepoMetadataModelsDirectoryGroupsDeleteOne(ctx, datum); err != nil {
				n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsInsertMany, err).Error())
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusInternalServerError}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusInternalServerError), "delete failed", err.Error()}
				failed += 1
				goto appendNewVerboseResponse
			} else {
				verbRes.Data = []any{datum}
				verbRes.Status = make([]intdoment.MetadataModelVerboseResponseStatus, 1)
				verbRes.Status[0].StatusCode = []int{http.StatusOK}
				verbRes.Status[0].StatusMessage = []string{http.StatusText(http.StatusOK), "deletion successful"}
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

	verbres.Message = fmt.Sprintf("Delete %[1]s: %[2]d/%[4]d successful and %[3]d/%[4]d failed", intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, successful, failed, len(data))
	verbres.Successful = successful
	verbres.Failed = failed

	return http.StatusOK, verbres, nil
}

func (n *service) ServiceMetadataModelsDirectoryGroupsUpdateMany(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	verboseResponse bool,
	data []*intdoment.MetadataModelsDirectoryGroups,
) (int, *intdoment.MetadataModelVerbRes, error) {
	verbres := new(intdoment.MetadataModelVerbRes)
	verbres.MetadataModelVerboseResponse = new(intdoment.MetadataModelVerboseResponse)
	if verboseResponse {
		if d, err := intlib.MetadataModelMiscGet(intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsUpdateMany, err).Error())
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

		if len(datum.DirectoryGroupsID) > 0 && len(datum.MetadataModelsID) > 0 {
			if authContextDirectoryGroupID.String() != datum.DirectoryGroupsID[0].String() {
				if value, err := n.repo.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID(ctx, authContextDirectoryGroupID, datum.DirectoryGroupsID[0]); err != nil {
					n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsUpdateMany, err).Error())
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

			if err := n.repo.RepoMetadataModelsDirectoryGroupsUpdateOne(ctx, datum); err != nil {
				n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsUpdateMany, err).Error())
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

	verbres.Message = fmt.Sprintf("Update %[1]s: %[2]d/%[4]d successful and %[3]d/%[4]d failed", intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, successful, failed, len(data))
	verbres.Successful = successful
	verbres.Failed = failed

	return http.StatusOK, verbres, nil
}

func (n *service) ServiceMetadataModelsDirectoryGroupsInsertMany(
	ctx context.Context,
	iamCredential *intdoment.IamCredentials,
	iamAuthorizationRules *intdoment.IamAuthorizationRules,
	authContextDirectoryGroupID uuid.UUID,
	verboseResponse bool,
	data []*intdoment.MetadataModelsDirectoryGroups,
) (int, *intdoment.MetadataModelVerbRes, error) {
	verbres := new(intdoment.MetadataModelVerbRes)
	verbres.MetadataModelVerboseResponse = new(intdoment.MetadataModelVerboseResponse)
	if verboseResponse {
		if d, err := intlib.MetadataModelMiscGet(intlib.METADATA_MODELS_MISC_VERBOSE_RESPONSE); err != nil {
			n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsInsertMany, err).Error())
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

		if len(datum.DirectoryGroupsID) > 0 && len(datum.MetadataModelsID) > 0 {
			if authContextDirectoryGroupID.String() != datum.DirectoryGroupsID[0].String() {
				if value, err := n.repo.RepoDirectoryGroupsSubGroupsFindOneBySubGroupID(ctx, authContextDirectoryGroupID, datum.DirectoryGroupsID[0]); err != nil {
					n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsInsertMany, err).Error())
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

			if value, err := n.repo.RepoMetadataModelsDirectoryGroupsInsertOne(ctx, datum, nil); err != nil {
				n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsInsertMany, err).Error())
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

	verbres.Message = fmt.Sprintf("Create %[1]s: %[2]d/%[4]d successful and %[3]d/%[4]d failed", intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName, successful, failed, len(data))
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

func (n *service) ServiceMetadataModelsDirectoryGroupsSearch(
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
	if value, err := n.repo.RepoMetadataModelsDirectoryGroupsSearch(
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
		n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsSearch, err).Error())
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		return value, nil
	}
}

func (n *service) ServiceMetadataModelsDirectoryGroupsGetMetadataModel(ctx context.Context, metadataModelRetrieve intdomint.MetadataModelRetrieve, targetJoinDepth int) (map[string]any, error) {
	if value, err := metadataModelRetrieve.MetadataModelsDirectoryGroupsGetMetadataModel(ctx, 0, targetJoinDepth, nil); err != nil {
		n.logger.Log(ctx, slog.LevelWarn+1, intlib.FunctionNameAndError(n.ServiceMetadataModelsDirectoryGroupsGetMetadataModel, err).Error())
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
	repo   intdomint.RouteMetadataModelsDirectoryGroupsRepository
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
