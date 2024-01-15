package lib

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const currentSection = "Identity and Access Management"

const sessionCacheTtl = 24 * time.Hour

func CacheDeleteSessionInfo(sessionId string) error {
	err := RedisClient(REDIS_SESSION_DB).Del(context.Background(), fmt.Sprintf(`ssn_%v`, sessionId)).Err()
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Cache delete session %v failed | reason: %v", sessionId, err))
		return NewError(http.StatusInternalServerError, "Could not authenticate user")
	}
	return nil
}

func CacheGetSessionInfo(sessionId string) (string, float64, error) {
	if directoryId, err := RedisClient(REDIS_SESSION_DB).Get(context.Background(), fmt.Sprintf(`ssn_%v`, sessionId)).Result(); err != nil {
		intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get directory id from cache for session %v failed | reason: %v", sessionId, err))
		return "", 0.0, NewError(http.StatusInternalServerError, "Could not authenticate user")
	} else {
		return directoryId, RedisClient(REDIS_SESSION_DB).TTL(context.Background(), fmt.Sprintf(`ssn_%v`, sessionId)).Val().Minutes(), nil
	}
}

func CacheSetSessionInfo(sessionId string, directoryId string) error {
	if err := RedisClient(REDIS_SESSION_DB).Set(context.Background(), fmt.Sprintf(`ssn_%v`, sessionId), directoryId, sessionCacheTtl).Err(); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Caching session id failed for %v | reason: %v", directoryId, err))
		return NewError(http.StatusInternalServerError, "Could not authenticate user")
	}
	return nil
}

func CacheDeleteUserInfo(directoryId string) error {
	err := RedisClient(REDIS_SESSION_DB).Del(context.Background(), fmt.Sprintf(`dir_%v`, directoryId)).Err()
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Cache delete user info failed for %v | reason: %v", directoryId, err))
		return NewError(http.StatusInternalServerError, "Could not authenticate user")
	}
	return nil
}

func CacheGetUserInfo(directoryId string) (User, error) {
	value, err := RedisClient(REDIS_SESSION_DB).Get(context.Background(), fmt.Sprintf(`dir_%v`, directoryId)).Result()
	if err != nil {
		intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Get user info from cache for %v failed | reason: %v", directoryId, err))
		return User{}, NewError(http.StatusInternalServerError, "Could not authenticate user")
	}
	var currentUser User
	err = json.Unmarshal([]byte(value), &currentUser)
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Converting user info to json, for caching, for %v | reason: %v", directoryId, err))
		return User{}, NewError(http.StatusInternalServerError, "Could not authenticate user")
	}
	return currentUser, nil
}

func CacheSetUserInfo(currentUser User) error {
	value, err := json.Marshal(currentUser)
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Converting user info to json, for caching, for %v | reason: %v", currentUser.DirectoryID.String(), err))
		return NewError(http.StatusInternalServerError, "Could not authenticate user")
	}
	err = RedisClient(REDIS_SESSION_DB).Set(context.Background(), fmt.Sprintf(`dir_%v`, currentUser.DirectoryID.String()), string(value), sessionCacheTtl).Err()
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Caching user info failed for %v | reason: %v", currentUser.DirectoryID.String(), err))
		return NewError(http.StatusInternalServerError, "Could not authenticate user")
	}
	return nil
}

func GenerateAccessRefreshToken(directoryId string) (string, error) {
	arToken, err := generateToken(directoryId, sessionCacheTtl).SignedString([]byte(os.Getenv("ACCESS_REFRESH_TOKEN")))
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Generate access-refresh token for %v failed | reason: %v", directoryId, err))
		return "", NewError(http.StatusInternalServerError, DEFAULT_AUTHENTICATION_ERROR)
	}
	arTokenEncrypt, err := encryptToken(directoryId, arToken)
	if err != nil {
		return "", err
	}
	return arTokenEncrypt, nil
}

func GetCookie(name string, value string, maxAge int) *http.Cookie {
	var cookie = http.Cookie{
		Name:   name,
		Value:  value,
		MaxAge: maxAge,
	}

	if os.Getenv("GO_DEV") == "true" {
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteStrictMode
		cookie.Secure = false
	} else {
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteStrictMode
		cookie.Secure = true
		cookie.Domain = os.Getenv("DOMAIN")
	}

	if os.Getenv("BASE_PATH") == "" {
		cookie.Path = "/"
	} else {
		cookie.Path = os.Getenv("BASE_PATH")
	}

	return &cookie
}

func encryptToken(directoryId, token string) (string, error) {
	tokenToEncrypt := []byte(token)
	cipherBlock, err := aes.NewCipher([]byte(os.Getenv("ENCRYPTION_KEY")))
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("IAM | generate cipher block failed for %v | reason: %v", directoryId, err))
		return "", NewError(http.StatusInternalServerError, DEFAULT_AUTHENTICATION_ERROR)
	}

	encryptedToken := make([]byte, aes.BlockSize+len(tokenToEncrypt))
	iv := encryptedToken[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("IAM | validate iv failed for %v | reason: %v", directoryId, err))
		return "", NewError(http.StatusInternalServerError, DEFAULT_AUTHENTICATION_ERROR)
	}

	stream := cipher.NewCFBEncrypter(cipherBlock, iv)
	stream.XORKeyStream(encryptedToken[aes.BlockSize:], tokenToEncrypt)
	return base64.URLEncoding.EncodeToString(encryptedToken), nil
}

func generateToken(directoryId string, ttl time.Duration) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    os.Getenv("DOMAIN_URL"),
			Subject:   directoryId,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	)
}
