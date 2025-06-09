package entities

const (
	OPENID_GRANT_TYPE_AUTHORIZATION_CODE string = "authorization_code"
	OPENID_GRANT_TYPE_REFRESH_TOKEN      string = "refresh_token"
	OPENID_GRANT_TYPE_PASSWORD           string = "password"

	OPENID_RESPONSE_TYPE_CODE string = "code"

	OPENID_HEADER_ACCESS_TOKEN             string = "OpenID-Access-Token"
	OPENID_HEADER_ACCESS_TOKEN_EXPIRES_IN  string = "OpenID-Access-Token-Expires-In"
	OPENID_HEADER_REFRESH_TOKEN            string = "OpenID-Refresh-Token"
	OPENID_HEADER_REFRESH_TOKEN_EXPIRES_IN string = "OpenID-Refresh-Token-Expires-In"
)

type OpenIDRedirectParams struct {
	SessionState string
	Iss          string
	Code         string
}

type OpenIDConfiguration struct {
	Issuer                     string   `json:"issuer,omitempty"`
	AuthorizationEndpoint      string   `json:"authorization_endpoint,omitempty"`
	TokenEndpoint              string   `json:"token_endpoint,omitempty"`
	TokenIntrospectionEndpoint string   `json:"introspection_endpoint,omitempty"`
	UserinfoEndpoint           string   `json:"userinfo_endpoint,omitempty"`
	RevocationEndpoint         string   `json:"revocation_endpoint,omitempty"`
	GrantTypesSupported        []string `json:"grant_types_supported,omitempty"`
	ResponseTypesSupported     []string `json:"response_types_supported,omitempty"`
	LoginEndpoint              string   `json:"login_endpoint,omitempty"`
}

type OpenIDToken struct {
	AccessToken      string `json:"access_token,omitempty"`
	ExpiresIn        int64  `json:"expires_in,omitempty"`
	RefreshExpiresIn int64  `json:"refresh_expires_in,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	IDToken          string `json:"id_token,omitempty"`
	NotBeforePolicy  int    `json:"not-before-policy,omitempty"`
	SessionState     string `json:"session_state,omitempty"`
	Scope            string `json:"scope,omitempty"`
}

type OpenIDUserInfo struct {
	Sub               string `json:"sub,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	Email             string `json:"email,omitempty"`
}

func (n *OpenIDUserInfo) IsFamilyNameValid() bool {
	return len(n.FamilyName) > 0
}

func (n *OpenIDUserInfo) IsGivenNameValid() bool {
	return len(n.GivenName) > 0
}

func (n *OpenIDUserInfo) IsSubValid() bool {
	return len(n.Sub) > 0
}

func (n *OpenIDUserInfo) IsPreferredUsernameValid() bool {
	return len(n.PreferredUsername) > 0
}

func (n *OpenIDUserInfo) IsEmailValid() bool {
	return len(n.Email) > 0
}

type OpenIDTokenIntrospect struct {
	Exp            int64    `json:"exp,omitempty"`
	Iat            int64    `json:"iat,omitempty"`
	Jti            string   `json:"jti,omitempty"`
	Iss            string   `json:"iss,omitempty"`
	Aud            any      `json:"aud,omitempty"`
	Sub            string   `json:"sub,omitempty"`
	Type           string   `json:"type,omitempty"`
	Azp            string   `json:"azp,omitempty"`
	Sid            string   `json:"sid,omitempty"`
	Acr            string   `json:"acr,omitempty"`
	AllowedOrigins []string `json:"allowed_origns,omitempty"`
	RealmAccess    struct {
		Roles []string `json:"roles,omitempty"`
	} `json:"realm_access,omitempty"`
	ResourceAccess struct {
		RealmManagement struct {
			Roles []string `json:"roles,omitempty"`
		} `json:"realm-management,omitempty"`
		Broker struct {
			Roles []string `json:"roles,omitempty"`
		} `json:"broker,omitempty"`
		Account struct {
			Roles []string `json:"roles,omitempty"`
		} `json:"account,omitempty"`
	} `json:"resource_access,omitempty"`
	Scope             string `json:"scope,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	Email             string `json:"email,omitempty"`
	ClientID          string `json:"client_id,omitempty"`
	Username          string `json:"username,omitempty"`
	TokenType         string `json:"token_type,omitempty"`
	Active            bool   `json:"active,omitempty"`
}
