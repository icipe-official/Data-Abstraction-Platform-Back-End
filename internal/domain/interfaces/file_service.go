package interfaces

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

type FormFileUpload interface {
	FormFile(key string) (multipart.File, *multipart.FileHeader, error)
	FormValue(key string) string
}

type FileService interface {
	Create(ctx context.Context, storageFile *intdoment.StorageFiles, file io.Reader) error
	Delete(ctx context.Context, storageFile *intdoment.StorageFiles) error
	Download(ctx context.Context, storageFile *intdoment.StorageFiles, w http.ResponseWriter, r *http.Request) error
}
