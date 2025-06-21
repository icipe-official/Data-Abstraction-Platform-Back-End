package iam

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intwebservice "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/web_service"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *service) ServiceGetIamAuthenticationHeaders(ctx context.Context, iamCookie http.Cookie) *intdoment.IamAuthenticationHeaders {
	iamAuthenticationHeaders := new(intdoment.IamAuthenticationHeaders)

	iamAuthenticationHeaders.AccessTokenHeader = intlib.IamCookieGetAccessTokenName(iamCookie.Name)
	iamAuthenticationHeaders.RefreshTokenHeader = intlib.IamCookieGetRefreshTokenName(iamCookie.Name)
	iamAuthenticationHeaders.CookieDomain = iamCookie.Domain
	iamAuthenticationHeaders.CookieHttpOnly = iamCookie.HttpOnly
	iamAuthenticationHeaders.CookieSameSite = iamCookie.SameSite
	iamAuthenticationHeaders.CookieSecure = iamCookie.Secure

	return iamAuthenticationHeaders
}
func (n *service) ServiceGetIamOpenIDEndpoints(ctx context.Context, openid intdomint.OpenID) *intdoment.IamOpenIDEndpoints {
	iamOpenidEndpoints := new(intdoment.IamOpenIDEndpoints)

	if value := openid.OpenIDGetLoginEndpoint(); len(value) > 0 {
		iamOpenidEndpoints.LoginEndpoint = value
	}

	if value, err := openid.OpenIDGetRegistrationEndpoint(); err == nil {
		iamOpenidEndpoints.RegistrationEndpoint = value
	}

	if value, err := openid.OpenIDGetAccountManagementEndpoint(); err == nil {
		iamOpenidEndpoints.AccountManagementEndpoint = value
	}

	return iamOpenidEndpoints
}

func (n *service) ServiceGetIamSession(ctx context.Context, openid intdomint.OpenID, iamAuthInfo *intdoment.IamCredentials) (*intdoment.IamSession, error) {
	iamSession := new(intdoment.IamSession)

	iamSession.IamCredential = iamAuthInfo

	if iamAuthInfo != nil && len(iamAuthInfo.ID) > 0 {
		if value, err := n.repo.RepoDirectoryFindOneByIamCredentialID(ctx, iamAuthInfo.ID[0], nil); err == nil {
			iamSession.Directory = value
		}
	}

	return iamSession, nil
}

func (n *service) ServiceOpenIDRevokeToken(ctx context.Context, openid intdomint.OpenID, token *intdoment.OpenIDToken) error {
	if err := openid.OpenIDRevokeToken(token); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("revoke open id token failed, error: %v", err), intlib.FunctionName(n.ServiceOpenIDRevokeToken))
		return intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	return nil
}

func (n *service) ServiceGetIamCredentialsByOpenIDSub(ctx context.Context, openIDTokenIntrospect *intdoment.OpenIDTokenIntrospect) (*intdoment.IamCredentials, error) {
	iamCredentials, err := n.repo.RepoIamCredentialsFindOneByID(ctx, intdoment.IamCredentialsRepository().OpenidSub, openIDTokenIntrospect.Sub, []string{intdoment.IamCredentialsRepository().ID, intdoment.IamCredentialsRepository().DeactivatedOn})
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

	//TODO: Create new iamCredential?
	return nil, intlib.NewError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func (n *service) ServiceOpenIDIntrospectToken(ctx context.Context, openid intdomint.OpenID, token *intdoment.OpenIDToken) (*intdoment.OpenIDTokenIntrospect, error) {
	if value, err := openid.OpenIDIntrospectToken(token); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("revoke open id token failed, error: %v", err), intlib.FunctionName(n.ServiceOpenIDIntrospectToken))
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		return value, nil
	}
}

type service struct {
	repo   intdomint.RouteIamRepository
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
