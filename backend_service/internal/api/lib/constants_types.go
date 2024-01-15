package lib

import (
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const CHUNK_SIZE = 512 * 1024

const (
	GEN_FILE_EXCEL string = "excel"
	GEN_FILE_CSV   string = "csv"
)

const EmailFormat = "Subject: %v\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n<html><body>%v</body></html>"

type JsonMessage struct {
	Message string `json:"message"`
}

type User struct {
	DirectoryID         uuid.UUID `sql:"primary_key"`
	IamEmail            *string
	Name                string
	Contacts            pq.StringArray
	SystemUserCreatedOn time.Time
	ProjectsRoles       []Project
	SessionId           string `json:"-"`
}

type Project struct {
	ProjectID          uuid.UUID `sql:"primary_key"`
	ProjectName        string
	ProjectDescription string
	ProjectCreatedOn   string
	ProjectRoles       []ProjectRoles
}

type ProjectRoles struct {
	ProjectRoleID        string `sql:"primary_key"`
	ProjectRoleCreatedOn time.Time
}

type CtxKey string

const CURRENT_USER_CTX_KEY = CtxKey("current_user")

var INTERNAL_SERVER_ERROR = NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))

const POSTGRES_NOT_FOUND_ERROR = "qrm: no rows in result set"

const DEFAULT_AUTHENTICATION_ERROR = "Could not authenticate user"

const ACCESS_REFRESH_TOKEN_AGE = 60 * 60 * 24

const (
	ROLE_ABSTRACTIONS_ADMIN      string = "abstractions_admin"
	ROLE_CATALOGUE_CREATOR       string = "catalogue_creator"
	ROLE_EDITOR                  string = "editor"
	ROLE_FILE_CREATOR            string = "file_creator"
	ROLE_MODEL_TEMPLATES_CREATOR string = "model_templates_creator"
	ROLE_PROJECT_ADMIN           string = "project_admin"
	ROLE_REVIEWER                string = "reviewer"
	ROLE_EXPLORER                string = "explorer"
)

const (
	STORAGE_AZURE_BLOB_MOUNTED string = "azure_blob_mounted"
	STORAGE_LOCAL              string = "local"
)

var REDIS_SESSION_DB = getRedisSessionDb()

const API_VERSION = "v1"

const COLUMN_UPDATE_SPLIT_LIST = "!!"

var TMP_DIR = os.Getenv("TMP_DIR")
