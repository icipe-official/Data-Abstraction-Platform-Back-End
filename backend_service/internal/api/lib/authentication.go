package lib

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"slices"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func IsUserAuthorized(isSystemUser bool, projectId uuid.UUID, ProjectRoles []string, user User, w http.ResponseWriter) bool {
	var isAuthorized bool
	if isSystemUser {
		if (user.SystemUserCreatedOn == time.Time{}) {
			if projectId == uuid.Nil {
				isAuthorized = false
			}
		} else {
			isAuthorized = true
		}
	}

	if projectId != uuid.Nil {
		for _, project := range user.ProjectsRoles {
			if project.ProjectID == projectId {
				if len(ProjectRoles) > 0 {
					for _, roles := range project.ProjectRoles {
						if slices.Contains(ProjectRoles, roles.ProjectRoleID) {
							isAuthorized = true
							break
						}
					}
				} else {
					isAuthorized = true
				}
			}
			if isAuthorized {
				break
			}
		}
	}

	if !isAuthorized && w != nil {
		SendErrorResponse(NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden)), w)
	}

	return isAuthorized
}

func AuthenticationMiddleware(next http.Handler) http.Handler {
	const currentSection string = "Identity and Access Management"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if currentUser, sessionId, expireTime, err := GetAuthenticatedUserInfo(r); err != nil {
			SendErrorResponse(err, w)
			return
		} else {
			if expireTime <= 60.0 && r.URL.Path != "/logout" && r.URL.Path != "/logout/" {
				newSessionId := uuid.New().String()
				if err = CacheSetSessionInfo(newSessionId, currentUser.DirectoryID.String()); err == nil {
					currentUser.SessionId = newSessionId
					if err = CacheDeleteSessionInfo(sessionId); err == nil {
						if err = CacheSetUserInfo(currentUser); err == nil {
							if arTokenEncrypt, err := GenerateAccessRefreshToken(newSessionId); err == nil {
								http.SetCookie(w, GetCookie("dap", arTokenEncrypt, ACCESS_REFRESH_TOKEN_AGE))
							} else {
								intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Generate new cookie for %v failed | reason: %v", currentUser.DirectoryID, err))
							}
						} else {
							intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Update session expiry time for %v failed | reason: %v", currentUser.DirectoryID, err))
						}
					} else {
						intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete old session for %v failed | reason: %v", currentUser.DirectoryID, err))
					}
				} else {
					intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Store new session for %v failed | reason: %v", currentUser.DirectoryID, err))
				}
			} else {
				currentUser.SessionId = sessionId
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CURRENT_USER_CTX_KEY, currentUser)))
		}
	})
}

func GetAuthenticatedUserInfo(r *http.Request) (User, string, float64, error) {
	arTokenEncrypt, err := r.Cookie("dap")
	if err != nil {
		return User{}, "", 0.0, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	}
	aToken, err := decryptToken(arTokenEncrypt.Value)
	if err != nil {
		return User{}, "", 0.0, err
	}
	sessionId, err := verifyToken(aToken, []byte(os.Getenv("ACCESS_REFRESH_TOKEN")))
	if err != nil {
		return User{}, "", 0.0, err
	}

	directoryId, expireTime, err := CacheGetSessionInfo(sessionId)
	if err != nil {
		return User{}, "", 0.0, err
	}

	currentUser, err := CacheGetUserInfo(directoryId)
	if err != nil {
		return User{}, "", 0.0, err
	}
	return currentUser, sessionId, expireTime, nil
}

func verifyToken(tokenToVerify string, tokenKey []byte) (string, error) {
	token, err := jwt.Parse(tokenToVerify, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		}
		return tokenKey, nil
	})
	if err != nil || !token.Valid {
		return "", NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	}
	claims := token.Claims.(jwt.MapClaims)
	if issuer, err := claims.GetIssuer(); err != nil || issuer != os.Getenv("DOMAIN_URL") {
		return "", NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	}
	if directoryId, err := claims.GetSubject(); err != nil {
		return "", NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	} else {
		return directoryId, nil
	}
}

func decryptToken(token string) (string, error) {
	encryptedToken, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Decode token failed | reason: %v", err))
		return "", NewError(http.StatusInternalServerError, DEFAULT_AUTHENTICATION_ERROR)
	}

	cipherBlock, err := aes.NewCipher([]byte(os.Getenv("ENCRYPTION_KEY")))
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Generate cipher block failed | reason: %v", err))
		return "", NewError(http.StatusInternalServerError, DEFAULT_AUTHENTICATION_ERROR)
	}

	if len(encryptedToken) < aes.BlockSize {
		intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, "Encrypted token too short")
		return "", NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	}

	iv := encryptedToken[:aes.BlockSize]
	encryptedToken = encryptedToken[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(cipherBlock, iv)
	stream.XORKeyStream(encryptedToken, encryptedToken)
	return string(encryptedToken), nil
}
