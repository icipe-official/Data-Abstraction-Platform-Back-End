package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/httplog/v2"
	intfs "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/file_service"
	intjobservice "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/job_service"
	intjobdeltmpfiles "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/job_service/delete_temporary_files"
	intrepopostgres "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/repository/postgres"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func main() {
	jobService := new(intjobservice.JobService)
	jobService.Logger = httplog.NewLogger(intlib.LogGetServiceName("job-service"), httplog.Options{
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

	func() {
		logAttribute := slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue("startup")}
		ctx := context.TODO()
		if value, err := intrepopostgres.NewPostgresRepository(ctx, jobService.Logger); err != nil {
			jobService.Logger.Log(ctx, slog.LevelError, fmt.Sprintf("Setup postgres connection pool failed, error: %v", err), logAttribute)
			os.Exit(1)
		} else {
			jobService.PostgresRepository = value
			if err := jobService.PostgresRepository.Ping(ctx); err != nil {
				jobService.Logger.Log(context.TODO(), slog.LevelWarn, fmt.Sprintf("Ping postgres database failed, error: %v", err), logAttribute)
			} else {
				jobService.Logger.Log(context.TODO(), slog.LevelInfo, "Ping postgres database successful", logAttribute)
			}
		}
	}()

	func() {
		logAttribute := slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue("startup")}
		ctx := context.TODO()

		jobService.Logger.Log(context.TODO(), slog.Level(2), "Setting up file service...")
		if value, err := intfs.NewS3FileService(jobService.Logger); err == nil {
			jobService.Logger.Log(ctx, slog.LevelInfo, "S3 setup complete", logAttribute)
			jobService.FileService = value
		} else {
			jobService.Logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("Setup s3 file service failed, error: %v", err), logAttribute)
		}

		if jobService.FileService == nil {
			if value, err := intfs.NewLocalFileService(jobService.Logger); err == nil {
				jobService.Logger.Log(ctx, slog.LevelInfo, "Local folder setup complete", logAttribute)
				jobService.FileService = value
			} else {
				jobService.Logger.Log(ctx, slog.LevelWarn, fmt.Sprintf("Setup local file service failed, error: %v", err), logAttribute)
			}
		}

		if jobService.FileService == nil {
			jobService.Logger.Log(ctx, slog.LevelError, fmt.Sprintln("No file service setup"), logAttribute)
			os.Exit(1)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	stop := make(chan bool, 1)

	go intjobdeltmpfiles.DeleteTemporaryFiles(jobService, stop)

	jobService.Logger.Log(context.Background(), slog.LevelInfo, "Job service running.")

	<-sigs
	stop <- true

	jobService.Logger.Log(context.Background(), slog.LevelInfo, "Job service stopping...")
	time.Sleep(500 * time.Millisecond)
}
