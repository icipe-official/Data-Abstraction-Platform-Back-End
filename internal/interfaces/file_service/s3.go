package filemanagement

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	ENV_S3_ENDPOINT   string = "STORAGE_S3_ENDPOINT"
	ENV_S3_ACCESS_KEY string = "STORAGE_S3_ACCESS_KEY"
	ENV_S3_SECRET_KEY string = "STORAGE_S3_SECRET_KEY"
	ENV_S3_USE_SSL    string = "STORAGE_S3_USE_SSL"
	ENV_S3_BUCKET     string = "STORAGE_S3_BUCKET"
)

func (n *S3) Download(ctx context.Context, storageFile *intdoment.StorageFiles, w http.ResponseWriter, r *http.Request) error {
	if len(storageFile.ID) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.ID is empty"))
	}

	reqParams := make(url.Values)
	// Set request parameters for content-disposition.
	// reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", storageFile.ID[0].String()))

	// Generates a presigned url which expires in a day.
	presignedURL, err := n.client.PresignedGetObject(context.Background(), n.bucket, storageFile.ID[0].String(), time.Second*60*60, reqParams)
	if err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("get presigned url for file failed, err: %v", err), "storageFile", intlib.JsonStringifyMust(storageFile))
		return intlib.FunctionNameAndError(n.Create, fmt.Errorf("get presigned url for file failed, err: %v", err))
	}

	w.Header().Set("Content-Type", storageFile.StorageFileMimeType[0])
	http.Redirect(w, r, presignedURL.String(), http.StatusFound)
	return nil
}

func (n *S3) Create(ctx context.Context, storageFile *intdoment.StorageFiles, file io.Reader) error {
	if len(storageFile.ID) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.ID is empty"))
	}

	if len(storageFile.SizeInBytes) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.SizeInBytes is empty"))
	}

	if len(storageFile.StorageFileMimeType) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.StorageFileMimeType is empty"))
	}

	minioPutOptions := minio.PutObjectOptions{
		ContentType: storageFile.StorageFileMimeType[0],
	}

	minioPutOptions.UserMetadata = map[string]string{}
	if len(storageFile.OriginalName) > 0 {
		minioPutOptions.UserMetadata[intdoment.StorageFilesRepository().OriginalName] = storageFile.OriginalName[0]
	}
	if len(storageFile.Tags) > 0 {
		minioPutOptions.UserMetadata[intdoment.StorageFilesRepository().Tags] = strings.Join(storageFile.Tags, " , ")
	}

	if uploadInfo, err := n.client.PutObject(ctx, n.bucket, storageFile.ID[0].String(), file, storageFile.SizeInBytes[0], minioPutOptions); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("upload file failed, err: %v", err), "storageFile", intlib.JsonStringifyMust(storageFile))
		return intlib.FunctionNameAndError(n.Create, fmt.Errorf("upload file failed, err: %v", err))
	} else {
		n.logger.Log(ctx, slog.LevelInfo+2, "file uploaded", "uploadInfo", intlib.JsonStringifyMust(uploadInfo))
		return nil
	}
}

func (n *S3) Delete(ctx context.Context, storageFile *intdoment.StorageFiles) error {
	if len(storageFile.ID) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.ID is empty"))
	}

	if err := n.client.RemoveObject(ctx, n.bucket, storageFile.ID[0].String(), minio.RemoveObjectOptions{}); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("delete file failed, err: %v", err), "storageFile", intlib.JsonStringifyMust(storageFile))
		return intlib.FunctionNameAndError(n.Create, fmt.Errorf("delete file failed, err: %v", err))
	}

	return nil
}

func NewS3FileService(logger intdomint.Logger) (*S3, error) {
	n := new(S3)
	n.logger = logger

	s3Endpoint := os.Getenv(ENV_S3_ENDPOINT)
	if len(s3Endpoint) == 0 {
		return nil, fmt.Errorf("%s not set", ENV_S3_ENDPOINT)
	}

	s3AccessKey := os.Getenv(ENV_S3_ACCESS_KEY)
	if len(s3AccessKey) == 0 {
		return nil, fmt.Errorf("%s not set", ENV_S3_ACCESS_KEY)
	}

	s3SecretKey := os.Getenv(ENV_S3_SECRET_KEY)
	if len(s3SecretKey) == 0 {
		return nil, fmt.Errorf("%s not set", ENV_S3_SECRET_KEY)
	}

	s3UseSSL := true
	if os.Getenv(ENV_S3_USE_SSL) == "false" {
		s3UseSSL = false
	}

	n.bucket = "data-abstraction-platform"
	if value := os.Getenv(ENV_S3_BUCKET); len(value) > 0 {
		n.bucket = value
	}

	if value, err := minio.New(s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3AccessKey, s3SecretKey, ""),
		Secure: s3UseSSL,
	}); err == nil {
		n.client = value
	} else {
		return nil, err
	}

	return n, nil
}

type S3 struct {
	client *minio.Client
	bucket string
	logger intdomint.Logger
}
