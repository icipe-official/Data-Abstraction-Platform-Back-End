package lib

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/httplog/v2"
)

const (
	LOG_ACTION_CREATE string = "create"
	LOG_ACTION_UPDATE string = "update"
	LOG_ACTION_DELETE string = "delete"
)

func LogSectionName(path string, env *EnvVariables) string {
	return strings.Replace(path, env.env[ENV_WEB_SERVICE_BASE_PATH], "/", 1)
}

func LogAction(action string, repoName string) string {
	return "action " + action + " " + repoName + " executed"
}

const LogSectionAttrKey string = "section"

func LogGetServiceName(serviceSubName string) string {
	lsn := os.Getenv("LOG_SERVICE_BASE_NAME")
	if lsn == "" {
		return fmt.Sprintf("%v-%v", "Data-Abstraction-Platform", serviceSubName)
	} else {
		return fmt.Sprintf("%v-%v", lsn, serviceSubName)
	}
}

func LogGetOptionBool(envName string) bool {
	envValue := os.Getenv(envName)
	if envValue == "true" {
		return true
	} else {
		return false
	}
}

func LogGetLevel() int {
	if lv, err := strconv.Atoi(os.Getenv("LOG_LEVEL")); err != nil {
		return 0
	} else {
		return lv
	}
}

func LogNewHttpLogger() *httplog.Logger {
	return httplog.NewLogger(LogGetServiceName("web-service"), httplog.Options{
		JSON:             LogGetOptionBool("LOG_USE_JSON"),
		LogLevel:         slog.Level(LogGetLevel()),
		Concise:          LogGetOptionBool("LOG_COINCISE"),
		RequestHeaders:   LogGetOptionBool("LOG_REQUEST_HEADERS"),
		MessageFieldName: "message",
		TimeFieldFormat:  time.RFC3339,
		Tags: map[string]string{
			"version": os.Getenv("LOG_APP_VERSION"),
		},
	})
}
