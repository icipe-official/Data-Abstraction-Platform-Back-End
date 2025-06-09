package lib

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type EnvVariables struct {
	env map[string]string
}

func NewEnvMap() (*EnvVariables, error) {
	n := new(EnvVariables)

	n.env = make(map[string]string)

	n.env[ENV_WEB_SERVICE_BASE_PATH] = func() string {
		basePath := os.Getenv(os.Getenv(ENV_WEB_SERVICE_BASE_PATH))

		if !strings.HasPrefix(basePath, "/") {
			basePath = "/" + basePath
		}

		if !strings.HasSuffix(basePath, "/") {
			basePath += "/"
		}

		return basePath
	}()

	if value := os.Getenv(ENV_WEB_SERVICE_BASE_URL); len(value) > 0 {
		if uri, err := url.Parse(value); err == nil {
			n.env[ENV_WEB_SERVICE_BASE_URL] = uri.String()
		} else {
			return nil, fmt.Errorf("parse %s as url failed, error: %v", ENV_WEB_SERVICE_BASE_URL, err)
		}
	} else {
		return nil, fmt.Errorf("env variable %s not set", ENV_WEB_SERVICE_BASE_URL)
	}

	if value := os.Getenv(ENV_WEBSITE_URL); len(value) > 0 {
		if uri, err := url.Parse(value); err == nil {
			n.env[ENV_WEBSITE_URL] = uri.String()
		} else {
			return nil, fmt.Errorf("parse %s as url failed, error: %v", ENV_WEBSITE_URL, err)
		}
	} else {
		return nil, fmt.Errorf("env variable %s not set", ENV_WEBSITE_URL)
	}

	if key, err := func() (string, error) {
		value := os.Getenv(ENV_IAM_ENCRYPTION_KEY)
		iekLength := len(value)
		if iekLength != 16 && iekLength != 24 && iekLength != 32 {
			return "string", fmt.Errorf("env variable %s can only be 16, 24, or 32 characters in length ONLY", ENV_IAM_ENCRYPTION_KEY)
		}

		return value, nil
	}(); err != nil {
		return nil, err
	} else {
		n.env[ENV_IAM_ENCRYPTION_KEY] = key
	}

	n.env[ENV_IAM_ENCRYPT_TOKENS] = os.Getenv(ENV_IAM_ENCRYPT_TOKENS)

	return n, nil
}

func (n *EnvVariables) Set(key string, value string) {
	n.env[key] = value
}

func (n *EnvVariables) Get(key string) string {
	if envString, ok := n.env[key]; ok {
		return envString
	}

	return ""
}
