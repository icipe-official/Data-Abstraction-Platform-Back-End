package interfaces

import (
	"context"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

type RouteMetadataModelsRepository interface {
	RepoMetadataModelsDeleteOne(
		ctx context.Context,
		iamAuthRule *intdoment.IamAuthorizationRule,
		datum *intdoment.MetadataModels,
	) error
	RepoMetadataModelsFindOneForDeletionByID(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		datum *intdoment.MetadataModels,
		columns []string,
	) (*intdoment.MetadataModels, *intdoment.IamAuthorizationRule, error)
	RepoMetadataModelsUpdateOne(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		datum *intdoment.MetadataModels,
	) error
	RepoMetadataModelsInsertOne(
		ctx context.Context,
		iamAuthRule *intdoment.IamAuthorizationRule,
		directoryID uuid.UUID,
		directoryGroupID uuid.UUID,
		datum *intdoment.MetadataModels,
		columns []string,
	) (*intdoment.MetadataModels, error)
	RepoIamGroupAuthorizationsGetAuthorized(
		ctx context.Context,
		iamAuthInfo *intdoment.IamCredentials,
		authContextDirectoryGroupID uuid.UUID,
		groupAuthorizationRules []*intdoment.IamGroupAuthorizationRule,
		currentIamAuthorizationRules *intdoment.IamAuthorizationRules,
	) ([]*intdoment.IamAuthorizationRule, error)
	RepoDirectoryGroupsFindSystemGroup(ctx context.Context, columns []string) (*intdoment.DirectoryGroups, error)
	RepoDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID, columns []string) (*intdoment.DirectoryGroups, error)
	RepoMetadataModelsSearch(
		ctx context.Context,
		mmsearch *intdoment.MetadataModelSearch,
		repo IamRepository,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		startSearchDirectoryGroupID uuid.UUID,
		authContextDirectoryGroupID uuid.UUID,
		skipIfFGDisabled bool,
		skipIfDataExtraction bool,
		whereAfterJoin bool,
	) (*intdoment.MetadataModelSearchResults, error)
}

type RouteMetadataModelsApiService interface {
	ServiceMetadataModelsDeleteMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		verboseResponse bool,
		data []*intdoment.MetadataModels,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceMetadataModelsUpdateMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		verboseResponse bool,
		data []*intdoment.MetadataModels,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceMetadataModelsInsertMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		verboseResponse bool,
		data []*intdoment.MetadataModels,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceMetadataModelsSearch(
		ctx context.Context,
		mmsearch *intdoment.MetadataModelSearch,
		repo IamRepository,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		startSearchDirectoryGroupID uuid.UUID,
		authContextDirectoryGroupID uuid.UUID,
		skipIfFGDisabled bool,
		skipIfDataExtraction bool,
		whereAfterJoin bool,
	) (*intdoment.MetadataModelSearchResults, error)
	ServiceMetadataModelsGetMetadataModel(ctx context.Context, metadataModelRetrieve MetadataModelRetrieve, targetJoinDepth int) (map[string]any, error)
	ServiceDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID) (*intdoment.DirectoryGroups, error)
}
