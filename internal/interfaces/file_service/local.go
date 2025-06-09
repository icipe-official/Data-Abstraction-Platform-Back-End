package filemanagement

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

const (
	CHUNK_SIZE = 512 * 1024
)

func (n *Local) Download(ctx context.Context, storageFile *intdoment.StorageFiles, w http.ResponseWriter, _ *http.Request) error {
	if len(storageFile.ID) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.ID is empty"))
	}

	if len(storageFile.StorageFileMimeType[0]) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.StorageFileMimeType is empty"))
	}

	var file *os.File

	filePath := n.folderpath + "/" + storageFile.ID[0].String()
	if value, err := os.Open(filePath); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("open file failed, err: %v", err), "storageFile", intlib.JsonStringifyMust(storageFile))
		return intlib.FunctionNameAndError(n.Download, fmt.Errorf("open file failed, err: %v", err))
	} else {
		file = value
	}

	if file == nil {
		return intlib.FunctionNameAndError(n.Download, errors.New("file not found"))
	}

	fileBuffer := make([]byte, CHUNK_SIZE)
	w.Header().Set("Content-Type", storageFile.StorageFileMimeType[0])
	w.Header().Set("Cache-Control", "private, max-age=0")
	w.WriteHeader(http.StatusOK)

	n.logger.Log(ctx, slog.LevelDebug, "sending file...", "storageFile", intlib.JsonStringifyMust(storageFile), "file", intlib.JsonStringifyMust(file), "file", intlib.JsonStringifyMust(file))
	for i := 0; ; i++ {
		if noOfBytes, err := file.Read(fileBuffer); err != nil {
			if err != io.EOF {
				n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("read file in chunks failed, error: %v", err), "storageFile", intlib.JsonStringifyMust(storageFile), "file", intlib.JsonStringifyMust(file))
			}
			break
		} else {
			n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("%v: reading %v bytes from file", i+1, noOfBytes), "storageFile", intlib.JsonStringifyMust(storageFile), "file", intlib.JsonStringifyMust(file))
			w.Write(fileBuffer[:noOfBytes])
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			} else {
				n.logger.Log(ctx, slog.LevelError, "read file in chunks failed, error: could not create flusher", "storageFile", intlib.JsonStringifyMust(storageFile), "file", intlib.JsonStringifyMust(file))
			}
		}
	}
	n.logger.Log(ctx, slog.LevelDebug, "...sending file complete", "storageFile", intlib.JsonStringifyMust(storageFile), "file", intlib.JsonStringifyMust(file))

	return nil
}

func (n *Local) Create(ctx context.Context, storageFile *intdoment.StorageFiles, file io.Reader) error {
	if len(storageFile.ID) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.ID is empty"))
	}

	if _, err := os.Stat(n.folderpath); os.IsNotExist(err) {
		if err := os.Mkdir(n.folderpath, os.ModeDir|0644); err != nil {
			n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("create directory group folder failed, err: %v", err), "storageFile", intlib.JsonStringifyMust(storageFile))
			return intlib.FunctionNameAndError(n.Create, fmt.Errorf("create directory group folder failed, err: %v", err))
		}
	}

	out, err := os.Create(n.folderpath + "/" + storageFile.ID[0].String())
	if err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("create file failed, err: %v", err), "storageFile", intlib.JsonStringifyMust(storageFile))
		return intlib.FunctionNameAndError(n.Create, fmt.Errorf("create file failed, err: %v", err))
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("write to file failed, err: %v", err), "storageFile", intlib.JsonStringifyMust(storageFile))
		return intlib.FunctionNameAndError(n.Create, fmt.Errorf("write to file failed, err: %v", err))
	}

	return nil
}

func (n *Local) Delete(ctx context.Context, storageFile *intdoment.StorageFiles) error {
	if len(storageFile.ID) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.ID is empty"))
	}
	if len(storageFile.DirectoryGroupsID) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.DirectoryGroupsID is empty"))
	}

	if len(storageFile.StorageFileMimeType[0]) < 1 {
		return intlib.FunctionNameAndError(n.Create, errors.New("storageFile.StorageFileMimeType is empty"))
	}

	if err := os.Remove(n.folderpath + "/" + storageFile.ID[0].String()); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("delete file failed, err: %v", err), "storageFile", intlib.JsonStringifyMust(storageFile))
		return intlib.FunctionNameAndError(n.Download, fmt.Errorf("open file failed, err: %v", err))
	}

	return nil
}

const (
	ENV_LOCAL_FOLDER_PATH string = "STORAGE_LOCAL_FOLDER_PATH"
)

func NewLocalFileService(logger intdomint.Logger) (*Local, error) {
	n := new(Local)
	n.logger = logger

	if value := os.Getenv(ENV_LOCAL_FOLDER_PATH); len(value) > 0 {
		n.folderpath = value
	} else {
		return nil, errors.New("folder path not set")
	}

	return n, nil
}

type Local struct {
	logger     intdomint.Logger
	folderpath string
}
