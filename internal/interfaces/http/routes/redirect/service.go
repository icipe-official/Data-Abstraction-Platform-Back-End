package home

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

type service struct {
	repo   intdomint.RouteRedirectRepository
	logger intdomint.Logger
}

const (
	redirect_PARAM_SESSION_STATE string = "session_state"
	redirect_PARAM_ISS           string = "iss"
	redirect_PARAM_CODE          string = "code"
)

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

func (n *service) ServiceGetIamCredentialsByOpenIDSub(ctx context.Context, openIDUserInfo *intdoment.OpenIDUserInfo) (*intdoment.IamCredentials, error) {
	iamCredentials, err := n.repo.RepoIamCredentialsFindOneByID(ctx, intdoment.IamCredentialsRepository().OpenidSub, openIDUserInfo.Sub, []string{intdoment.IamCredentialsRepository().ID, intdoment.IamCredentialsRepository().DeactivatedOn})
	if err != nil {
		n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceGetIamCredentialsByOpenIDSub, err).Error())
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if iamCredentials != nil {
		if len(iamCredentials.DeactivatedOn) == 1 && !iamCredentials.DeactivatedOn[0].IsZero() {
			return nil, intlib.NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		}
		return iamCredentials, nil
	}

	iamCredentials, err = n.repo.RepoIamCredentialsInsertOpenIDUserInfo(ctx, openIDUserInfo, []string{intdoment.IamCredentialsRepository().ID})
	if err != nil {
		n.logger.Log(ctx, slog.LevelError, intlib.FunctionNameAndError(n.ServiceGetIamCredentialsByOpenIDSub, err).Error())
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	return iamCredentials, nil
}

func (n *service) ServiceOpenIDRevokeToken(ctx context.Context, openid intdomint.OpenID, token *intdoment.OpenIDToken) error {
	if err := openid.OpenIDRevokeToken(token); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("revoke open id token failed, error: %v", err), intlib.FunctionName(n.ServiceOpenIDRevokeToken))
		return intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	return nil
}

func (n *service) ServiceGetOpenIDUserInfo(ctx context.Context, openid intdomint.OpenID, token *intdoment.OpenIDToken) (*intdoment.OpenIDUserInfo, error) {
	if userInfo, err := openid.OpenIDGetUserinfo(token); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("get open id user info failed, error: %v", err), intlib.FunctionName(n.ServiceGetOpenIDUserInfo))
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		return userInfo, nil
	}
}

func (n *service) ServiceGetOpenIDToken(ctx context.Context, openid intdomint.OpenID, redirectParams *intdoment.OpenIDRedirectParams) (*intdoment.OpenIDToken, error) {
	if token, err := openid.OpenIDGetTokenFromRedirect(redirectParams); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("get open id token failed, error: %v", err), intlib.FunctionName(n.ServiceGetOpenIDToken))
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		return token, nil
	}
}
