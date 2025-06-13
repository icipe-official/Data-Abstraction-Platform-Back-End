package lib

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
)

func IamAuthenticationMiddleware(logger intdomint.Logger, env *EnvVariables, openId intdomint.OpenID, iamCookie http.Cookie, repo intdomint.IamRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			openIDToken := new(intdoment.OpenIDToken)

			tokenFromCookie := false
			if accessToken := r.Header.Get(IamCookieGetAccessTokenName(iamCookie.Name)); len(accessToken) > 0 {
				token := ""
				if env.Get(ENV_IAM_ENCRYPT_TOKENS) != "false" {
					if decrypted, err := DecryptData(env.Get(ENV_IAM_ENCRYPTION_KEY), accessToken); err != nil {
						logger.Log(r.Context(), slog.LevelWarn, fmt.Sprintf("header decrypt access token failed, error: %v", err))
					} else {
						token = decrypted
					}
				} else {
					token = accessToken
				}

				if len(token) > 0 {
					if tokenBytes, err := base64.StdEncoding.DecodeString(token); err != nil {
						logger.Log(r.Context(), slog.LevelWarn, fmt.Sprintf("header decode access token failed, error: %v", err))
						next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))))
						return
					} else {
						openIDToken.AccessToken = string(tokenBytes)
					}
				}
			} else {
				if encryptedCookie, err := r.Cookie(IamCookieGetAccessTokenName(iamCookie.Name)); err == nil {
					token := ""
					if env.Get(ENV_IAM_ENCRYPT_TOKENS) != "false" {
						if decrypted, err := DecryptData(env.Get(ENV_IAM_ENCRYPTION_KEY), encryptedCookie.Value); err != nil {
							logger.Log(r.Context(), slog.LevelWarn, fmt.Sprintf("cookie decrypt access token failed, error: %v", err))
						} else {
							token = decrypted
						}
					} else {
						token = encryptedCookie.Value
					}

					if len(token) > 0 {
						if tokenBytes, err := base64.StdEncoding.DecodeString(token); err != nil {
							logger.Log(r.Context(), slog.LevelWarn, fmt.Sprintf("cookie decode access token failed, error: %v", err))
							next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))))
							return
						} else {
							openIDToken.AccessToken = string(tokenBytes)
						}
					}

					tokenFromCookie = true
				}
			}

			if refreshToken := r.Header.Get(IamCookieGetRefreshTokenName(iamCookie.Name)); len(refreshToken) > 0 {
				token := refreshToken
				if env.Get(ENV_IAM_ENCRYPT_TOKENS) != "false" {
					if decrypted, err := DecryptData(env.Get(ENV_IAM_ENCRYPTION_KEY), token); err != nil {
						logger.Log(r.Context(), slog.LevelWarn, fmt.Sprintf("header decrypt refresh token failed, error: %v", err))
						next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))))
						return
					} else {
						token = decrypted
					}
				}
				if tokenBytes, err := base64.StdEncoding.DecodeString(token); err != nil {
					logger.Log(r.Context(), slog.LevelWarn, fmt.Sprintf("header decode refresh token failed, error: %v", err))
					next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))))
					return
				} else {
					openIDToken.RefreshToken = string(tokenBytes)
				}
			} else {
				if encryptedCookie, err := r.Cookie(IamCookieGetRefreshTokenName(iamCookie.Name)); err == nil {
					token := encryptedCookie.Value
					if env.Get(ENV_IAM_ENCRYPT_TOKENS) != "false" {
						if decrypted, err := DecryptData(env.Get(ENV_IAM_ENCRYPTION_KEY), token); err != nil {
							logger.Log(r.Context(), slog.LevelWarn, fmt.Sprintf("cookie decrypt refresh token failed, error: %v", err))
							next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))))
							return
						} else {
							token = decrypted
						}
					}
					if tokenBytes, err := base64.StdEncoding.DecodeString(token); err != nil {
						logger.Log(r.Context(), slog.LevelWarn, fmt.Sprintf("cookie decode refresh token failed, error: %v", err))
						next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))))
						return
					} else {
						openIDToken.RefreshToken = string(tokenBytes)
					}
				}
			}

			var newAccessRefreshToken *IamAccessRefreshToken
			openIDTokenIntrospect := new(intdoment.OpenIDTokenIntrospect)
			if value, err := openId.OpenIDIntrospectToken(openIDToken); err != nil {
				logger.Log(r.Context(), slog.LevelWarn+1, fmt.Sprintf("introspect open id token failed, error: %v", err))
				if strings.HasPrefix(r.URL.Path, fmt.Sprintf("%siam/sign-out", env.Get(ENV_WEB_SERVICE_BASE_PATH))) {
					next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)))))
					return
				}

				if strings.HasPrefix(r.URL.Path, fmt.Sprintf("%siam/refresh-token", env.Get(ENV_WEB_SERVICE_BASE_PATH))) || tokenFromCookie {
					if newToken, err := openId.OpenIDRefreshToken(openIDToken); err != nil {
						logger.Log(r.Context(), slog.LevelError, fmt.Sprintf("refresh open id token failed, error: %v", err))
						next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))))
						return
					} else {
						openIDToken = newToken
						if nvalue, err := openId.OpenIDIntrospectToken(openIDToken); err != nil {
							logger.Log(r.Context(), slog.LevelError, fmt.Sprintf("introspect open id token after refresh failed, error: %v", err))
							next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))))
							return
						} else {
							if token, err := IamPrepOpenIDTokenForClient(env, openIDToken); err != nil {
								logger.Log(r.Context(), slog.LevelError, fmt.Sprintf("Prepare access refresh token for client failed, error: %v", err))
								if err := openId.OpenIDRevokeToken(openIDToken); err != nil {
									logger.Log(r.Context(), slog.LevelError, fmt.Sprintf("Revoke access refresh token for client failed, error: %v", err))
								}
								next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))))
								return
							} else {
								if tokenFromCookie {
									IamSetCookieInResponse(w, iamCookie, token, int(openIDToken.ExpiresIn), int(openIDToken.RefreshExpiresIn))
								} else {
									newAccessRefreshToken = token
								}
							}
							openIDTokenIntrospect = nvalue
						}
					}

				}
			} else {
				openIDTokenIntrospect = value
			}

			iamCredentials, err := repo.RepoIamCredentialsFindOneByID(r.Context(), intdoment.IamCredentialsRepository().OpenidSub, openIDTokenIntrospect.Sub, nil)
			if err != nil {
				logger.Log(r.Context(), slog.LevelError, fmt.Sprintf("introspect open id token after refresh failed, error: %v", err))
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)))))
				return
			}
			if iamCredentials == nil {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))))
				return
			}
			if len(iamCredentials.DeactivatedOn) == 1 && !iamCredentials.DeactivatedOn[0].IsZero() {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ERROR_CODE_CTX_KEY, NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden)))))
				return
			}

			logger.Log(r.Context(), slog.LevelDebug, fmt.Sprintf("authed iamCredentials: %+v", iamCredentials))

			if strings.HasPrefix(r.URL.Path, fmt.Sprintf("%siam/sign-out", env.Get(ENV_WEB_SERVICE_BASE_PATH))) {
				if err := openId.OpenIDRevokeToken(openIDToken); err != nil {
					logger.Log(r.Context(), slog.LevelWarn, fmt.Sprintf("revoke open id token for logout failed, error: %v", err))
				}

				if tokenFromCookie {
					IamSetCookieInResponse(w, iamCookie, new(IamAccessRefreshToken), 0, 0)
				}
			}

			newCtx := context.WithValue(r.Context(), IAM_CREDENTIAL_ID_CTX_KEY, *iamCredentials)
			if newAccessRefreshToken != nil {
				newCtx = context.WithValue(newCtx, IAM_NEW_ACCESS_REFRESH_TOKEN, *newAccessRefreshToken)
			}
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}

const IAM_NEW_ACCESS_REFRESH_TOKEN = CtxKey("iam_new_access_refresh_token")

func IamHttpRequestCtxGetNewAccessRefreshToken(r *http.Request) (*IamAccessRefreshToken, error) {
	if value, ok := r.Context().Value(IAM_NEW_ACCESS_REFRESH_TOKEN).(IamAccessRefreshToken); ok {
		return &value, nil
	}
	return nil, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
}

func IamHttpRequestCtxGetAuthedIamCredential(r *http.Request) (*intdoment.IamCredentials, error) {
	if value, ok := r.Context().Value(IAM_CREDENTIAL_ID_CTX_KEY).(intdoment.IamCredentials); ok {
		return &value, nil
	}
	return nil, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
}

const IAM_CREDENTIAL_ID_CTX_KEY = CtxKey("iam_credential_id")

func IamGetCurrentUserIamCredentialID(r *http.Request) (*intdoment.IamCredentials, error) {
	if iamCredID, ok := r.Context().Value(IAM_CREDENTIAL_ID_CTX_KEY).(intdoment.IamCredentials); ok {
		return &iamCredID, nil
	}
	if err, ok := r.Context().Value(ERROR_CODE_CTX_KEY).(error); ok {
		return nil, err
	}
	return nil, NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
}

func GetTokenCookie(iamCookie http.Cookie, token string, refreshTokenAge int) *http.Cookie {
	return &http.Cookie{
		Name:     iamCookie.Name,
		Value:    token,
		MaxAge:   refreshTokenAge,
		HttpOnly: iamCookie.HttpOnly,
		Secure:   iamCookie.Secure,
		SameSite: iamCookie.SameSite,
		Domain:   iamCookie.Domain,
		Path:     iamCookie.Path,
	}
}

func IamSetCookieInResponse(w http.ResponseWriter, iamCookie http.Cookie, token *IamAccessRefreshToken, accessExpiresIn int, refreshExpiresIn int) {
	cookiePrefix := iamCookie.Name
	iamCookie.Name = IamCookieGetAccessTokenName(cookiePrefix)
	http.SetCookie(w, GetTokenCookie(iamCookie, token.AccessToken, accessExpiresIn))
	iamCookie.Name = IamCookieGetRefreshTokenName(cookiePrefix)
	http.SetCookie(w, GetTokenCookie(iamCookie, token.RefreshToken, refreshExpiresIn))
}

func IamPrepOpenIDTokenForClient(env *EnvVariables, token *intdoment.OpenIDToken) (*IamAccessRefreshToken, error) {
	tk := new(IamAccessRefreshToken)
	tk.AccessToken = base64.StdEncoding.EncodeToString([]byte(token.AccessToken))
	tk.RefreshToken = base64.StdEncoding.EncodeToString([]byte(token.RefreshToken))

	if env.Get(ENV_IAM_ENCRYPT_TOKENS) != "false" {
		if encrypted, err := EncryptData(env.Get(ENV_IAM_ENCRYPTION_KEY), []byte(tk.AccessToken)); err != nil {
			return nil, fmt.Errorf("encrypt minAccessRefreshToken failed, error: %v", err)
		} else {
			tk.AccessToken = encrypted
			tk.AccessTokenExpiresIn = token.ExpiresIn
		}
		if encrypted, err := EncryptData(env.Get(ENV_IAM_ENCRYPTION_KEY), []byte(tk.RefreshToken)); err != nil {
			return nil, fmt.Errorf("encrypt minAccessRefreshToken failed, error: %v", err)
		} else {
			tk.RefreshToken = encrypted
			tk.RefreshTokenExpiresIn = token.RefreshExpiresIn
		}
	}

	return tk, nil
}

type IamAccessRefreshToken struct {
	AccessToken           string `json:"access_token,omitempty"`
	AccessTokenExpiresIn  int64  `json:"access_token_expires_in,omitempty"`
	RefreshToken          string `json:"refresh_token,omitempty"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in,omitempty"`
}

func IamInitCookie(env *EnvVariables) http.Cookie {
	iamCookie := http.Cookie{}
	if cookieName := os.Getenv("IAM_COOKIE_NAME"); len(cookieName) > 0 {
		iamCookie.Name = cookieName
	} else {
		iamCookie.Name = WebServiceAppPrefix()
	}
	if os.Getenv("IAM_COOKIE_HTTP_ONLY") == "true" {
		iamCookie.HttpOnly = true
	} else {
		iamCookie.HttpOnly = false
	}
	if os.Getenv("IAM_COOKIE_SECURE") == "true" {
		iamCookie.Secure = true
	} else {
		iamCookie.Secure = false
	}
	if sameSite, err := strconv.Atoi(os.Getenv("IAM_COOKIE_SAME_SITE")); err != nil {
		iamCookie.SameSite = http.SameSite(http.SameSiteStrictMode)
	} else {
		iamCookie.SameSite = http.SameSite(sameSite)
	}
	iamCookie.Domain = os.Getenv("IAM_COOKIE_DOMAIN")
	iamCookie.Path = env.Get(ENV_WEB_SERVICE_BASE_PATH)
	return iamCookie
}

func IamCookieGetAccessTokenName(iamCookieName string) string {
	return iamCookieName + "_z"
}

func IamCookieGetRefreshTokenName(iamCookieName string) string {
	return iamCookieName + "_y"
}
