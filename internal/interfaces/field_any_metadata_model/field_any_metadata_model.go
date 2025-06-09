package fieldanymetadatamodel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
)

type FieldAnyMetadataModelGet struct {
	logger     intdomint.Logger
	repo       intdomint.FieldAnyMetadataModelRepository
	mmretrieve intdomint.MetadataModelRetrieve
}

func NewFieldAnyMetadataModelGet(logger intdomint.Logger, repo intdomint.FieldAnyMetadataModelRepository, metadataModelRetrieve intdomint.MetadataModelRetrieve) *FieldAnyMetadataModelGet {
	n := new(FieldAnyMetadataModelGet)
	n.logger = logger
	n.repo = repo
	n.mmretrieve = metadataModelRetrieve

	return n
}

func (n *FieldAnyMetadataModelGet) GetMetadataModel(ctx context.Context, actionID string, currentFgKey string, tableCollectionUid string, argument any) (any, error) {
	n.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf("actionID: %s; currentFgKey: %s; tableCollectionUid: %s; argument:%+v", actionID, currentFgKey, tableCollectionUid, argument))

	switch actionID {
	case intdoment.MetadataModelsDirectoryRepository().RepositoryName, intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName:
		if argArray, ok := argument.([]any); ok && len(argArray) > 0 {
			directoryGroupID := uuid.Nil
			if value, ok := argArray[0].(uuid.UUID); ok {
				directoryGroupID = value
			} else {
				if v, ok := argArray[0].(string); ok {
					if vUUID, err := uuid.FromString(v); err == nil {
						directoryGroupID = vUUID
					}
				}
			}
			if directoryGroupID.IsNil() {
				return nil, fmt.Errorf("in actionID %s, directoryGroupID is nil", actionID)
			}

			metadataMetadataModelIDColumnName := ""
			metadataDirectoryGroupIDColumnName := ""
			switch actionID {
			case intdoment.MetadataModelsDirectoryRepository().RepositoryName:
				metadataMetadataModelIDColumnName = intdoment.MetadataModelsDirectoryRepository().MetadataModelsID
				metadataDirectoryGroupIDColumnName = intdoment.MetadataModelsDirectoryRepository().DirectoryGroupsID
			case intdoment.MetadataModelsDirectoryGroupsRepository().RepositoryName:
				metadataMetadataModelIDColumnName = intdoment.MetadataModelsDirectoryGroupsRepository().MetadataModelsID
				metadataDirectoryGroupIDColumnName = intdoment.MetadataModelsDirectoryGroupsRepository().DirectoryGroupsID
			default:
				return nil, fmt.Errorf("actionID %s not recognized", actionID)
			}

			return n.repo.RepoMetadataModelFindOneByDirectoryGroupID(ctx, actionID, metadataMetadataModelIDColumnName, metadataDirectoryGroupIDColumnName, directoryGroupID)
		}

		return nil, fmt.Errorf("actionID %s arguments not valid", actionID)
	case intdoment.AbstractionsDirectoryGroupsRepository().RepositoryName:
		if argArray, ok := argument.([]any); ok && len(argArray) > 0 {
			directoryGroupID := uuid.Nil
			if value, ok := argArray[0].(uuid.UUID); ok {
				directoryGroupID = value
			} else {
				if v, ok := argArray[0].(string); ok {
					if vUUID, err := uuid.FromString(v); err == nil {
						directoryGroupID = vUUID
					}
				}
			}
			if directoryGroupID.IsNil() {
				return nil, fmt.Errorf("in actionID %s, directoryGroupID is nil", actionID)
			}

			return n.repo.RepoMetadataModelFindOneByAbstractionsDirectoryGroupsID(ctx, directoryGroupID)
		}

		return nil, fmt.Errorf("actionID %s arguments not valid", actionID)
	}

	return nil, fmt.Errorf("actionID %s not recognized", actionID)
}
