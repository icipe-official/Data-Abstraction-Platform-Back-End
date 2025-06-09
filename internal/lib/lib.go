package lib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	mathrand "math/rand"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

const (
	URL_SEARCH_PARAM_TARGET_JOIN_DEPTH               string = "target_join_depth"
	URL_SEARCH_PARAM_VERBOSE_RESPONSE                string = "verbose_response"
	URL_SEARCH_PARAM_SKIP_IF_DATA_EXTRACTION         string = "skip_if_data_extraction"
	URL_SEARCH_PARAM_SKIP_IF_FG_DISABLED             string = "skip_if_fg_disabled"
	URL_SEARCH_PARAM_SUB_QUERY                       string = "sub_query"
	URL_SEARCH_PARAM_START_SEARCH_DIRECTORY_GROUP_ID string = "start_search_directory_group_id"
	URL_SEARCH_PARAM_AUTH_CONTEXT_DIRECTORY_GROUP_ID string = "auth_context_directory_group_id"
	URL_SEARCH_PARAM_WHERE_AFTER_JOIN                string = "where_after_join"
	URL_SEARCH_PARAM_DIRECTORY_GROUP_ID              string = "directory_group_id"
)

func UrlSearchParamGetInt(r *http.Request, param string) (int, error) {
	if value, err := strconv.Atoi(r.URL.Query().Get(param)); err == nil {
		return value, nil
	} else {
		return 0, err
	}
}

func UrlSearchParamGetBool(r *http.Request, param string, defaultValue bool) bool {
	switch r.URL.Query().Get(param) {
	case "true":
		return true
	case "false":
		return false
	default:
		return defaultValue
	}
}

func UrlSearchParamGetUuid(r *http.Request, param string) (uuid.UUID, error) {
	if value, err := uuid.FromString(r.URL.Query().Get(param)); err != nil {
		return value, err
	} else {
		return value, nil
	}
}

type SessionData struct {
	OpenidEndpoints struct {
		LoginEndpoint             string `json:"login_endpoint,omitempty"`
		RegistrationEndpoint      string `json:"registration_endpoint,omitempty"`
		AccountManagementEndpoint string `json:"account_management_endpoint,omitempty"`
	} `json:"openid_endpoints,omitempty"`
	IamCredential    *intdoment.IamCredentials `json:"iam_credential,omitempty"`
	DirectoryGroupID *uuid.UUID                `json:"directory_group_id,omitempty"`
}

func DecryptData(encryptionKey string, data string) (string, error) {
	encryptedData, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("decode data string failed, error: %v", err.Error())
	}

	cipherBlock, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return "", fmt.Errorf("generate cipher block failed, error: %v", err)
	}

	if len(encryptedData) < aes.BlockSize {
		return "", errors.New("encryptedData length less than aes.BlockSize")
	}

	iv := encryptedData[:aes.BlockSize]
	encryptedData = encryptedData[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(cipherBlock, iv)
	stream.XORKeyStream(encryptedData, encryptedData)
	return string(encryptedData), nil
}

// data can be converted to a json string and passed as []byte(data)
func EncryptData(encryptionKey string, data []byte) (string, error) {
	cipherBlock, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return "", fmt.Errorf("generate cipher block failed, error: %v", err)
	}

	encryptedData := make([]byte, aes.BlockSize+len(data))
	iv := encryptedData[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("validate iv failed, error: %v", err)
	}

	stream := cipher.NewCFBEncrypter(cipherBlock, iv)
	stream.XORKeyStream(encryptedData[aes.BlockSize:], data)
	return base64.URLEncoding.EncodeToString(encryptedData), nil
}

const (
	ENV_IAM_ENCRYPTION_KEY string = "IAM_ENCRYPTION_KEY"
	ENV_IAM_ENCRYPT_TOKENS string = "IAM_ENCRYPT_TOKENS"

	ENV_WEB_SERVICE_BASE_PATH string = "WEB_SERVICE_BASE_PATH"
	ENV_WEB_SERVICE_BASE_URL  string = "WEB_SERVICE_BASE_URL"
	ENV_WEBSITE_URL           string = "WEBSITE_URL"
)

type CtxKey string

const ERROR_CODE_CTX_KEY = CtxKey("error")

const LOG_ATTR_CTX_KEY CtxKey = "LOG_ATTR"

func AnyToBytes(data any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const HTTP_ERROR_SPLIT string = "->"

func EmailValidationRegex() *regexp.Regexp {
	return regexp.MustCompile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)
}

func GenRandomString(length int, includeSpecialSymbols bool) string {
	stringPool := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	if includeSpecialSymbols {
		stringPool = append(stringPool, []rune("_~-")...)
	}
	randomString := make([]rune, length)
	for index := range randomString {
		randomString[index] = stringPool[mathrand.Intn(len(stringPool))]
	}
	return string(randomString)
}

func NewError(code int, message string) error {
	return fmt.Errorf("%v%v%v", code, HTTP_ERROR_SPLIT, message)
}

func SendJsonResponse(httpStatusCode int, data any, w http.ResponseWriter) {
	if json, err := json.Marshal(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Process JSON response failed"))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatusCode)
		w.Write(json)
	}
}

func SendJsonErrorResponse(errorResponse error, w http.ResponseWriter) {
	httpStatusCode, httpStatusMessage := SplitJsonErrorResponse(errorResponse)
	SendJsonResponse(httpStatusCode, struct {
		Message string `json:"message"`
	}{Message: httpStatusMessage}, w)
}

func SplitJsonErrorResponse(errorResponse error) (int, string) {
	newError := strings.Split(errorResponse.Error(), HTTP_ERROR_SPLIT)
	if httpStatusCode, err := strconv.Atoi(newError[0]); err == nil {
		return httpStatusCode, newError[1]
	}
	return 500, errorResponse.Error()
}

func CheckRequiredEnvVariables(envVars []string) []string {
	envVariablesMissing := make([]string, 0)
	for _, ev := range envVars {
		if os.Getenv(ev) == "" {
			envVariablesMissing = append(envVariablesMissing, ev)
		}
	}

	return envVariablesMissing
}

func WebServiceAppPrefix() string {
	appPrefix := os.Getenv("WEB_SERVICE_APP_PREFIX")
	if len(appPrefix) == 0 {
		return "dap"
	}
	return appPrefix
}

func FunctionNameAndError(function any, err error) error {
	return fmt.Errorf("%v -> %v", FunctionName(function), err)
}

func FunctionName(function any) string {
	return runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
}

// Will return the value as is if json.Marshal returns an error
func JsonStringifyMust(value any) any {
	if jsonData, err := json.Marshal(value); err == nil {
		return string(jsonData)
	}

	return value
}

func TableCollectionUidRegex() *regexp.Regexp {
	return regexp.MustCompile(`^\_([0-9]+)\_([0-9a-zA-Z\_]+)\_([a-zA-Z0-9]+)$`)
}

type TableCollectionUID struct {
	JoinDepth int
	TableName string
	Uid       string
}

func GetTableCollectionUid(tcuid string) *TableCollectionUID {
	tableRegex := TableCollectionUidRegex().FindStringSubmatch(tcuid)

	if len(tableRegex) == 4 {
		if joinDepth, err := strconv.Atoi(tableRegex[1]); err == nil {
			n := new(TableCollectionUID)

			n.JoinDepth = joinDepth
			n.TableName = tableRegex[2]
			n.Uid = tableRegex[3]

			return n
		}
	}

	return nil
}
