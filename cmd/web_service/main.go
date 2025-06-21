package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	intfs "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/file_service"
	intopenidkeycloak "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/openid/keycloak"
	intrepopostgres "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/repository/postgres"
	intwebservice "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/web_service"
	inthttprouters "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/web_service/routers"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func main() {
	webService := new(intwebservice.WebService)
	webService.Logger = httplog.NewLogger(intlib.LogGetServiceName("web-service"), httplog.Options{
		JSON:             intlib.LogGetOptionBool("LOG_USE_JSON"),
		LogLevel:         slog.Level(intlib.LogGetLevel()),
		Concise:          intlib.LogGetOptionBool("LOG_COINCISE"),
		RequestHeaders:   intlib.LogGetOptionBool("LOG_REQUEST_HEADERS"),
		MessageFieldName: "message",
		TimeFieldFormat:  time.RFC3339,
		Tags: map[string]string{
			"version": os.Getenv("LOG_APP_VERSION"),
		},
	})

	if value, err := intlib.NewEnvMap(); err != nil {
		webService.Logger.Log(context.TODO(), slog.LevelError, fmt.Sprintf("Setup env variables failed, error: %v", err))
		os.Exit(1)
	} else {
		webService.Env = value
	}
	webService.IamCookie = intlib.IamInitCookie(webService.Env)

	webService.Logger.Log(context.TODO(), slog.Level(2), "Setting up open id configuration...")
	if value, err := intopenidkeycloak.NewKeycloakOpenID(webService.Logger, webService.Env.Get(intlib.ENV_WEB_SERVICE_BASE_URL), webService.Env.Get(intlib.ENV_WEB_SERVICE_BASE_PATH)); err != nil {
		webService.Logger.Log(context.TODO(), slog.LevelError, fmt.Sprintf("Initialize Keycloak OpenID failed, error: %v", err))
		os.Exit(1)
	} else {
		webService.OpenID = value
	}

	func() {
		logAttribute := slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue("startup")}
		ctx := context.TODO()
		if value, err := intrepopostgres.NewPostgresRepository(ctx, webService.Logger); err != nil {
			webService.Logger.Log(ctx, slog.LevelError, fmt.Sprintf("Setup postgres connection pool failed, error: %v", err), logAttribute)
			os.Exit(1)
		} else {
			webService.PostgresRepository = value
			if err := webService.PostgresRepository.Ping(ctx); err != nil {
				webService.Logger.Log(context.TODO(), slog.LevelWarn, fmt.Sprintf("Ping postgres database failed, error: %v", err), logAttribute)
			} else {
				webService.Logger.Log(context.TODO(), slog.LevelInfo, "Ping postgres database successful", logAttribute)
			}
		}
	}()

	func() {
		logAttribute := slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue("startup")}
		ctx := context.TODO()

		webService.Logger.Log(context.TODO(), slog.Level(2), "Setting up file service...")
		if value, err := intfs.NewS3FileService(webService.Logger); err == nil {
			webService.Logger.Log(ctx, slog.LevelInfo, "S3 setup complete", logAttribute)
			webService.FileService = value
		} else {
			webService.Logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("Setup s3 file service failed, error: %v", err), logAttribute)
		}

		if webService.FileService == nil {
			if value, err := intfs.NewLocalFileService(webService.Logger); err == nil {
				webService.Logger.Log(ctx, slog.LevelInfo, "Local folder setup complete", logAttribute)
				webService.FileService = value
			} else {
				webService.Logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("Setup local file service failed, error: %v", err), logAttribute)
			}
		}

		if webService.FileService == nil {
			webService.Logger.Log(ctx, slog.LevelError, fmt.Sprintln("No file service setup"), logAttribute)
			os.Exit(1)
		}
	}()

	router := chi.NewRouter()
	if httpwebServiceLogger, ok := webService.Logger.(*httplog.Logger); ok {
		router.Use(httplog.RequestLogger(httpwebServiceLogger))
	}
	router.Use(middleware.Heartbeat(webService.Env.Get(intlib.ENV_WEB_SERVICE_BASE_PATH) + "ping"))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(os.Getenv("WEB_SERVICE_CORS_URLS"), " "),
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT"},
		AllowedHeaders:   []string{"Accept", "Content-Type", intlib.IamCookieGetAccessTokenName(webService.IamCookie.Name), intlib.IamCookieGetRefreshTokenName(webService.IamCookie.Name)},
		AllowCredentials: true,
	}))

	inthttprouters.InitApiCoreRouter(router, webService)

	// Start http server
	webService.Logger.Log(context.TODO(), slog.Level(2), fmt.Sprintf("Server will be listening on port: %v at base path '%v'", os.Getenv("WEB_SERVICE_PORT"), webService.Env.Get(intlib.ENV_WEB_SERVICE_BASE_PATH)), slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue("startup")})
	if err := http.ListenAndServe(":"+os.Getenv("WEB_SERVICE_PORT"), router); err != nil {
		webService.Logger.Log(context.TODO(), slog.LevelError, fmt.Sprintf("Could not start server , error: %v", err), slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue("startup")})
		os.Exit(1)
	}
}
