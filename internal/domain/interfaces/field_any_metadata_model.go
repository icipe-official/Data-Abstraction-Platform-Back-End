package interfaces

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type FieldAnyMetadataModel interface {
	GetMetadataModel(ctx context.Context, actionID string, currentFgKey string, tableCollectionUid string, argument any) (any, error)
}

type FieldAnyMetadataModelRepository interface {
	RepoMetadataModelFindOneByDirectoryGroupID(
		ctx context.Context,
		metadataModelRepositoryName string,
		metadataMetadataModelIDFieldColumn string,
		metadataDirectoryGroupIDFieldColumn string,
		directoryGroupID uuid.UUID,
	) (map[string]any, error)
	RepoMetadataModelFindOneByAbstractionsDirectoryGroupsID(ctx context.Context, directoryGroupID uuid.UUID) (map[string]any, error)
}
