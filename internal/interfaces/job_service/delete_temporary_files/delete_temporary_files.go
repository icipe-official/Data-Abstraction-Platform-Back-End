package deletetemporaryfiles

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intjobservice "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/job_service"
)

func DeleteTemporaryFiles(jobService *intjobservice.JobService, stop chan bool) {
	ctx := context.Background()

	jobService.Logger.Log(ctx, slog.LevelInfo, "Initializing job service to delete temporary files")

	s := initJobService(ctx, jobService)

	ticker := time.NewTicker(time.Hour)

	go func() {
		for {
			select {
			case <-stop:
				ticker.Stop()
				jobService.Logger.Log(ctx, slog.LevelInfo, "Terminating job service to delete temporary files succeeded")
				return
			case t := <-ticker.C:
				jobService.Logger.Log(ctx, slog.LevelInfo, "Running job to delete temporary files", "time", t.Format(time.RFC3339Nano))

				s.ServiceStorageFilesTemporaryDelete(ctx, jobService.FileService)
			}
		}
	}()

	jobService.Logger.Log(ctx, slog.LevelInfo, "initialization of job service to delete temporary files succeeded")
}

func (n *service) ServiceStorageFilesTemporaryDelete(ctx context.Context, fileService intdomint.FileService) {
	if result, err := n.repo.RepoStorageFilesDeleteTemporaryFiles(ctx, fileService); err != nil {
		n.logger.Log(ctx, slog.LevelError, err.Error())
	} else {
		n.logger.Log(ctx, slog.LevelInfo+1, fmt.Sprintf("In delete files: %v successful & %v failed", result.Success, len(result.Failed)))
		if len(result.Failed) > 0 {
			for _, f := range result.Failed {
				n.logger.Log(ctx, slog.LevelError, "Delete file failed", "error", f)
			}
		}
	}
}

func initJobService(ctx context.Context, jobService *intjobservice.JobService) intdomint.StorageFilesTemporaryService {
	if value, err := NewService(jobService); err != nil {
		errmsg := fmt.Errorf("initialize job service failed, error: %v", err)
		if value.logger != nil {
			value.logger.Log(ctx, slog.LevelError, errmsg.Error())
		} else {
			log.Println(errmsg)
		}

		return nil
	} else {
		return value
	}
}

type service struct {
	repo   intdomint.StorageFilesTemporaryRepository
	logger intdomint.Logger
}

func NewService(webService *intjobservice.JobService) (*service, error) {
	n := new(service)

	n.repo = webService.PostgresRepository
	n.logger = webService.Logger

	if n.logger == nil {
		return n, errors.New("webService.Logger is empty")
	}

	if n.repo == nil {
		return n, errors.New("webService.PostgresRepository is empty")
	}

	return n, nil
}
