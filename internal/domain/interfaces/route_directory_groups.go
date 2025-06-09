package interfaces

import (
	"context"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

type RouteDirectoryGroupsRepository interface {
	RepoDirectoryGroupsFindSystemGroup(ctx context.Context, columns []string) (*intdoment.DirectoryGroups, error)
	RepoDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID, columns []string) (*intdoment.DirectoryGroups, error)
	RepoDirectoryGroupsSearch(
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
	RepoDirectoryGroupsDeleteOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.DirectoryGroups) error
	RepoDirectoryGroupsUpdateOne(ctx context.Context, datum *intdoment.DirectoryGroups, fieldAnyMetadataModelGet FieldAnyMetadataModel) error
	RepoDirectoryGroupsInsertOne(
		ctx context.Context,
		datum *intdoment.DirectoryGroups,
		authContextDirectoryGroupID uuid.UUID,
		iamAuthorizationRule *intdoment.IamAuthorizationRule,
		fieldAnyMetadataModelGet FieldAnyMetadataModel,
		columns []string,
	) (*intdoment.DirectoryGroups, error)
	RepoMetadataModelFindOneByDirectoryGroupID(
		ctx context.Context,
		metadataModelRepositoryName string,
		metadataMetadataModelIDFieldColumn string,
		metadataDirectoryGroupIDFieldColumn string,
		directoryGroupID uuid.UUID,
	) (map[string]any, error)
	RepoDirectoryGroupsSubGroupsFindOneBySubGroupID(ctx context.Context, parentGroupID uuid.UUID, subGroupID uuid.UUID) (*intdoment.DirectoryGroupsSubGroups, error)
	RepoDirectoryGroupsCheckIfSystemGroup(ctx context.Context, directoryGroupID uuid.UUID) (bool, error)
}

type RouteDirectoryGroupsApiCoreService interface {
	ServiceDirectoryGroupsDeleteMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		verboseResponse bool,
		data []*intdoment.DirectoryGroups,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceDirectoryGroupsUpdateMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		fieldAnyMetadataModelGet FieldAnyMetadataModel,
		verboseResponse bool,
		data []*intdoment.DirectoryGroups,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceDirectoryGroupsInsertMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		fieldAnyMetadataModelGet FieldAnyMetadataModel,
		verboseResponse bool,
		data []*intdoment.DirectoryGroups,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceDirectoryGroupsSearch(
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
	ServiceDirectoryGroupsGetMetadataModel(ctx context.Context, metadataModelRetrieve MetadataModelRetrieve, targetJoinDepth int) (map[string]any, error)
	ServiceDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID) (*intdoment.DirectoryGroups, error)
}
