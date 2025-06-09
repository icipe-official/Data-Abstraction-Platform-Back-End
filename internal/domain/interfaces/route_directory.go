package interfaces

import (
	"context"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

type RouteDirectoryRepository interface {
	RepoDirectoryFindOneForDeletionByID(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		datum *intdoment.Directory,
		columns []string,
	) (*intdoment.Directory, *intdoment.IamAuthorizationRule, error)
	RepoDirectoryGroupsFindSystemGroup(ctx context.Context, columns []string) (*intdoment.DirectoryGroups, error)
	RepoDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID, columns []string) (*intdoment.DirectoryGroups, error)
	RepoDirectorySearch(
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
	RepoIamGroupAuthorizationsGetAuthorized(
		ctx context.Context,
		iamAuthInfo *intdoment.IamCredentials,
		authContextDirectoryGroupID uuid.UUID,
		groupAuthorizationRules []*intdoment.IamGroupAuthorizationRule,
		currentIamAuthorizationRules *intdoment.IamAuthorizationRules,
	) ([]*intdoment.IamAuthorizationRule, error)
	RepoDirectoryDeleteOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.Directory) error
	RepoDirectoryUpdateOne(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		fieldAnyMetadataModelGet FieldAnyMetadataModel,
		authContextDirectoryGroupID uuid.UUID,
		datum *intdoment.Directory,
	) error
	RepoDirectoryInsertOne(
		ctx context.Context,
		datum *intdoment.Directory,
		authContextDirectoryGroupID uuid.UUID,
		iamAuthorizationRule *intdoment.IamAuthorizationRule,
		fieldAnyMetadataModelGet FieldAnyMetadataModel,
		columns []string,
	) (*intdoment.Directory, error)
	RepoMetadataModelFindOneByDirectoryGroupID(
		ctx context.Context,
		metadataModelRepositoryName string,
		metadataMetadataModelIDFieldColumn string,
		metadataDirectoryGroupIDFieldColumn string,
		directoryGroupID uuid.UUID,
	) (map[string]any, error)
	RepoDirectoryGroupsSubGroupsFindOneBySubGroupID(ctx context.Context, parentGroupID uuid.UUID, subGroupID uuid.UUID) (*intdoment.DirectoryGroupsSubGroups, error)
}

type RouteDirectoryApiCoreService interface {
	ServiceDirectoryDeleteMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		verboseResponse bool,
		data []*intdoment.Directory,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceDirectoryUpdateMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		fieldAnyMetadataModelGet FieldAnyMetadataModel,
		verboseResponse bool,
		data []*intdoment.Directory,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceDirectoryInsertMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		fieldAnyMetadataModelGet FieldAnyMetadataModel,
		verboseResponse bool,
		data []*intdoment.Directory,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceDirectorySearch(
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
	ServiceDirectoryGetMetadataModel(ctx context.Context, metadataModelRetrieve MetadataModelRetrieve, targetJoinDepth int) (map[string]any, error)
	ServiceDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID) (*intdoment.DirectoryGroups, error)
}
