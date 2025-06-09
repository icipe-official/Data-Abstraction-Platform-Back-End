package keycloak

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
)

const (
	ENV_OPENID_CLIENT_ID                   string = "OPENID_CLIENT_ID"
	ENV_OPENID_CLIENT_SECRET               string = "OPENID_CLIENT_SECRET"
	ENV_OPENID_USER_REGISTRATION_ENDPOINT  string = "OPENID_USER_REGISTRATION_ENDPOINT"
	ENV_OPENID_LOGIN_ENDPOINT              string = "OPENID_LOGIN_ENDPOINT"
	ENV_OPENID_LOGIN_REDIRECT_URL          string = "OPENID_LOGIN_REDIRECT_URL"
	ENV_OPENID_ACCOUNT_MANAGEMENT_ENDPOINT string = "OPENID_ACCOUNT_MANAGEMENT_ENDPOINT"
)

type KeycloakOpenID struct {
	openIDConfig                    *intdoment.OpenIDConfiguration
	logger                          intdomint.Logger
	openIDClientID                  string
	openIDClientSecret              string
	openIDUserRegistrationEndpoint  string
	openIDRedirectUrl               string
	openIDLoginEndpoint             string
	openIDAccountManagementEndpoint string
}

func NewKeycloakOpenID(logger intdomint.Logger, baseUrl string, basePath string) (*KeycloakOpenID, error) {
	n := new(KeycloakOpenID)

	request, err := http.NewRequest(http.MethodGet, os.Getenv("OPENID_CONFIGURATION_ENDPOINT"), nil)
	if err != nil {
		return nil, fmt.Errorf("create new http request failed, error: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute http request failed, error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request not OK. http status code: %d; http status: %s", response.StatusCode, response.Status)
	}

	n.openIDConfig = new(intdoment.OpenIDConfiguration)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed, error: %v", err)
	}
	if err := json.Unmarshal(body, n.openIDConfig); err != nil {
		return nil, fmt.Errorf("read body in json failed, error: %v", err)
	}

	if len(n.openIDConfig.Issuer) == 0 {
		return nil, errors.New("OpenIDConfig.Issuer is nil")
	}

	if len(n.openIDConfig.AuthorizationEndpoint) == 0 {
		return nil, errors.New("OpenIDConfig.AuthorizationEndpoint is nil")
	}

	if len(n.openIDConfig.TokenEndpoint) == 0 {
		return nil, errors.New("OpenIDConfig.TokenEndpoint is nil")
	}

	if len(n.openIDConfig.TokenIntrospectionEndpoint) == 0 {
		return nil, errors.New("OpenIDConfig.TokenIntrospectionEndpoint is nil")
	}

	if len(n.openIDConfig.UserinfoEndpoint) == 0 {
		return nil, errors.New("OpenIDConfig.UserinfoEndpoint is nil")
	}

	if len(n.openIDConfig.RevocationEndpoint) == 0 {
		return nil, errors.New("OpenIDConfig.UserinfoEndpoint is nil")
	}

	if len(n.openIDConfig.GrantTypesSupported) < 2 {
		return nil, errors.New("not enough OpenIDConfig.GrantTypesSupported supported")
	} else {
		if !slices.Contains(n.openIDConfig.GrantTypesSupported, intdoment.OPENID_GRANT_TYPE_AUTHORIZATION_CODE) {
			return nil, fmt.Errorf("OpenIDConfig.GrantTypesSupported type '%s' not supported", intdoment.OPENID_GRANT_TYPE_AUTHORIZATION_CODE)
		}

		if !slices.Contains(n.openIDConfig.GrantTypesSupported, intdoment.OPENID_GRANT_TYPE_REFRESH_TOKEN) {
			return nil, fmt.Errorf("OpenIDConfig.GrantTypesSupported type '%s' not supported", intdoment.OPENID_GRANT_TYPE_REFRESH_TOKEN)
		}

		if !slices.Contains(n.openIDConfig.GrantTypesSupported, intdoment.OPENID_GRANT_TYPE_REFRESH_TOKEN) {
			n.logger.Log(context.TODO(), slog.LevelWarn, fmt.Sprintf("OpenIDConfig.GrantTypesSupported type '%s' not supported, Direct token retrieval from openid server not available", intdoment.OPENID_GRANT_TYPE_PASSWORD))
		}
	}

	if len(n.openIDConfig.ResponseTypesSupported) == 0 {
		return nil, errors.New("not enough OpenIDConfig.ResponseTypesSupported supported")
	} else {
		if !slices.Contains(n.openIDConfig.ResponseTypesSupported, intdoment.OPENID_RESPONSE_TYPE_CODE) {
			return nil, fmt.Errorf("OpenIDConfig.ResponseTypesSupported type '%s' not supported", intdoment.OPENID_RESPONSE_TYPE_CODE)
		}
	}

	n.openIDClientID = os.Getenv(ENV_OPENID_CLIENT_ID)
	n.openIDClientSecret = os.Getenv(ENV_OPENID_CLIENT_SECRET)
	n.openIDUserRegistrationEndpoint = os.Getenv(ENV_OPENID_USER_REGISTRATION_ENDPOINT)
	n.openIDAccountManagementEndpoint = os.Getenv(ENV_OPENID_ACCOUNT_MANAGEMENT_ENDPOINT)

	loginRedirectUrl := new(url.URL)
	if url, err := url.Parse(baseUrl); err != nil {
		return nil, fmt.Errorf("parse baseUrl failed, error: %v", err)
	} else {
		loginRedirectUrl = url
	}
	if len(basePath) > 0 {
		loginRedirectUrl.Path += basePath
	}
	loginRedirectUrl.Path += "redirect"

	n.openIDRedirectUrl = loginRedirectUrl.String()

	loginEndpointUrl := new(url.URL)
	if url, err := url.Parse(n.openIDConfig.AuthorizationEndpoint); err != nil {
		return nil, err
	} else {
		loginEndpointUrl = url
	}

	params := url.Values{}
	params.Add("scope", "openid")
	params.Add("response_type", "code")
	params.Add("client_id", n.openIDClientID)
	params.Add("redirect_uri", loginRedirectUrl.String())
	loginEndpointUrl.RawQuery = params.Encode()

	n.openIDLoginEndpoint = loginEndpointUrl.String()

	return n, nil
}

func (n *KeycloakOpenID) OpenIDGetLoginEndpoint() string {
	return n.openIDLoginEndpoint
}

func (n *KeycloakOpenID) OpenIDGetRegistrationEndpoint() (string, error) {
	if len(n.openIDUserRegistrationEndpoint) > 0 {
		return n.openIDUserRegistrationEndpoint, nil
	}
	return "", errors.New("openIDUserRegistrationEndpoint not set")
}

func (n *KeycloakOpenID) OpenIDGetAccountManagementEndpoint() (string, error) {
	if len(n.openIDAccountManagementEndpoint) > 0 {
		return n.openIDAccountManagementEndpoint, nil
	}
	return "", errors.New("openIDAccountManagementEndpoint not set")
}

func (n *KeycloakOpenID) OpenIDGetConfig() intdoment.OpenIDConfiguration {
	return *n.openIDConfig
}

func (n *KeycloakOpenID) OpenIDGetTokenFromRedirect(redirectParams *intdoment.OpenIDRedirectParams) (*intdoment.OpenIDToken, error) {
	token := new(intdoment.OpenIDToken)

	data := url.Values{}
	data.Set("grant_type", intdoment.OPENID_GRANT_TYPE_AUTHORIZATION_CODE)
	data.Set("client_id", n.openIDClientID)
	data.Set("client_secret", n.openIDClientSecret)
	data.Set("code", redirectParams.Code)
	data.Set("redirect_uri", n.openIDRedirectUrl)

	request, err := http.NewRequest(http.MethodPost, n.openIDConfig.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create new http request failed, error: %v", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute http request failed, error: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed, error: %v", err)
	}

	if response.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, token); err != nil {
			return nil, fmt.Errorf("read body in json failed, error: %v", err)
		}
	} else {
		return nil, fmt.Errorf("request error: StatusCode: %d; Status: %s; Body: %s", response.StatusCode, response.Status, string(body))
	}

	return token, nil
}

func (n *KeycloakOpenID) OpenIDIntrospectToken(token *intdoment.OpenIDToken) (*intdoment.OpenIDTokenIntrospect, error) {
	tokenIntrospect := new(intdoment.OpenIDTokenIntrospect)

	if len(token.AccessToken) == 0 {
		return nil, errors.New("token.AccessToken is empty")
	}

	data := url.Values{}
	data.Set("token", token.AccessToken)
	data.Set("client_id", n.openIDClientID)
	data.Set("client_secret", n.openIDClientSecret)

	request, err := http.NewRequest(http.MethodPost, n.openIDConfig.TokenIntrospectionEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create new http request failed, error: %v", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute http request failed, error: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed, error: %v", err)
	}

	if response.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, tokenIntrospect); err != nil {
			return nil, fmt.Errorf("read body in json failed, error: %v", err)
		}
	} else {
		return nil, fmt.Errorf("request error: StatusCode: %d; Status: %s; Body: %s", response.StatusCode, response.Status, string(body))
	}

	return tokenIntrospect, nil
}

func (n *KeycloakOpenID) OpenIDGetUserinfo(token *intdoment.OpenIDToken) (*intdoment.OpenIDUserInfo, error) {
	userInfo := new(intdoment.OpenIDUserInfo)

	if len(token.AccessToken) == 0 {
		return nil, errors.New("token.AccessToken is empty")
	}

	request, err := http.NewRequest(http.MethodGet, n.openIDConfig.UserinfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create new http request failed, error: %v", err)
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute http request failed, error: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed, error: %v", err)
	}

	if response.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, userInfo); err != nil {
			return nil, fmt.Errorf("read body in json failed, error: %v", err)
		}
	} else {
		return nil, fmt.Errorf("request error: StatusCode: %d; Status: %s; Body: %s", response.StatusCode, response.Status, string(body))
	}

	return userInfo, nil
}

func (n *KeycloakOpenID) OpenIDRefreshToken(token *intdoment.OpenIDToken) (*intdoment.OpenIDToken, error) {
	newToken := new(intdoment.OpenIDToken)

	if len(token.RefreshToken) == 0 {
		return nil, errors.New("token.RefreshToken is empty")
	}

	data := url.Values{}
	data.Set("client_id", n.openIDClientID)
	data.Set("client_secret", n.openIDClientSecret)
	data.Set("grant_type", intdoment.OPENID_GRANT_TYPE_REFRESH_TOKEN)
	data.Set("refresh_token", token.RefreshToken)

	request, err := http.NewRequest(http.MethodPost, n.openIDConfig.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create new http request failed, error: %v", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute http request failed, error: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed, error: %v", err)
	}

	if response.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, newToken); err != nil {
			return nil, fmt.Errorf("read body in json failed, error: %v", err)
		}
	} else {
		return nil, fmt.Errorf("request error: StatusCode: %d; Status: %s; Body: %s", response.StatusCode, response.Status, string(body))
	}

	return newToken, nil
}

func (n *KeycloakOpenID) OpenIDRevokeToken(token *intdoment.OpenIDToken) error {
	if len(token.RefreshToken) == 0 || len(token.AccessToken) == 0 {
		return errors.New("token.RefreshToken and token.AccessToken is empty")
	}

	if len(token.AccessToken) > 0 {
		data := url.Values{}
		data.Set("token", token.AccessToken)
		data.Set("token_type_hint", "access_token")

		request, err := http.NewRequest(http.MethodPost, n.openIDConfig.RevocationEndpoint, strings.NewReader(data.Encode()))
		if err != nil {
			return fmt.Errorf("create new http request failed, error: %v", err)
		}
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(n.openIDClientID+":"+n.openIDClientSecret))))
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return fmt.Errorf("execute http request failed, error: %v", err)
		}

		if response.StatusCode != http.StatusOK {
			defer response.Body.Close()
			body, _ := io.ReadAll(response.Body)
			return fmt.Errorf("request error: StatusCode: %d; Status: %s; Body: %s", response.StatusCode, response.Status, string(body))
		}
	}

	if len(token.RefreshToken) > 0 {
		data := url.Values{}
		data.Set("token", token.RefreshToken)
		data.Set("token_type_hint", "refresh_token")

		request, err := http.NewRequest(http.MethodPost, n.openIDConfig.RevocationEndpoint, strings.NewReader(data.Encode()))
		if err != nil {
			return fmt.Errorf("create new http request failed, error: %v", err)
		}
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(n.openIDClientID+":"+n.openIDClientSecret))))
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return fmt.Errorf("execute http request failed, error: %v", err)
		}

		if response.StatusCode != http.StatusOK {
			defer response.Body.Close()
			body, _ := io.ReadAll(response.Body)
			return fmt.Errorf("request error: StatusCode: %d; Status: %s; Body: %s", response.StatusCode, response.Status, string(body))
		}
	}

	return nil
}
