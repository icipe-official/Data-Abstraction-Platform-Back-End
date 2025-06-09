package metadatamodel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	inthttp "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http"

	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

type service struct {
	repo   intdomint.RouteMetadataModelRepository
	logger intdomint.Logger
}

func NewService(webService *inthttp.WebService) (*service, error) {
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

func (n *service) ServiceMetadataModelGet(ctx context.Context, metadataModelRepositoryName string, directoryGroupID uuid.UUID) (map[string]any, error) {
	metadataMetadataModelIDColumnName := ""
	metadataDirectoryGroupIDColumnName := ""
	switch metadataModelRepositoryName {
	case intdoment.MetadataModelsDirectoryRepository().RepositoryName:
		metadataMetadataModelIDColumnName = intdoment.MetadataModelsDirectoryRepository().MetadataModelsID
		metadataDirectoryGroupIDColumnName = intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID
	case intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName:
		metadataMetadataModelIDColumnName = intdoment.MetadataModelsDirectoryGroupsRepository().MetadataModelsID
		metadataDirectoryGroupIDColumnName = intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID
	default:
		return nil, intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	if value, err := n.repo.RepoMetadataModelFindOneByDirectoryGroupID(ctx, metadataModelRepositoryName, metadataMetadataModelIDColumnName, metadataDirectoryGroupIDColumnName, directoryGroupID); err != nil {
		n.logger.Log(ctx, slog.LevelError, fmt.Sprintf("Get metadata-model failed, error: %v", err), intlib.FunctionName(n.ServiceMetadataModelGet))
		return nil, intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	} else {
		if value == nil {
			return nil, intlib.NewError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}
		return value, nil
	}
}
