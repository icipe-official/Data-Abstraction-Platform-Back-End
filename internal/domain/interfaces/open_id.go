package interfaces

import intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"

type OpenID interface {
	OpenIDGetLoginEndpoint() string
	OpenIDGetRegistrationEndpoint() (string, error)
	OpenIDGetAccountManagementEndpoint() (string, error)
	OpenIDGetConfig() intdoment.OpenIDConfiguration
	OpenIDRevokeToken(token *intdoment.OpenIDToken) error
	OpenIDRefreshToken(token *intdoment.OpenIDToken) (*intdoment.OpenIDToken, error)
	OpenIDGetUserinfo(token *intdoment.OpenIDToken) (*intdoment.OpenIDUserInfo, error)
	OpenIDIntrospectToken(token *intdoment.OpenIDToken) (*intdoment.OpenIDTokenIntrospect, error)
	OpenIDGetTokenFromRedirect(redirectParams *intdoment.OpenIDRedirectParams) (*intdoment.OpenIDToken, error)
}
