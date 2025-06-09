package interfaces

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type RouteMetadataModelRepository interface {
	RepoMetadataModelFindOneByDirectoryGroupID(
		ctx context.Context,
		metadataModelRepositoryName string,
		metadataMetadataModelIDFieldColumn string,
		metadataDirectoryGroupIDFieldColumn string,
		directoryGroupID uuid.UUID,
	) (map[string]any, error)
}

type RouteMetadataModelApiService interface {
	ServiceMetadataModelGet(ctx context.Context, metadataModelRepositoryName string, directoryGroupID uuid.UUID) (map[string]any, error)
}
