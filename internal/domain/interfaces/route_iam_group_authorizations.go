package interfaces

import (
	"context"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

type RouteIamGroupAuthorizationsRepository interface {
	RepoIamGroupAuthorizationsSearch(
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
	RepoDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID, columns []string) (*intdoment.DirectoryGroups, error)
	RepoDirectoryGroupsFindSystemGroup(ctx context.Context, columns []string) (*intdoment.DirectoryGroups, error)
	RepoDirectoryGroupsSubGroupsFindOneBySubGroupID(ctx context.Context, parentGroupID uuid.UUID, subGroupID uuid.UUID) (*intdoment.DirectoryGroupsSubGroups, error)
	RepoIamGroupAuthorizationsFindOneInactiveRule(ctx context.Context, iamGroupAuthorizationID uuid.UUID, columns []string) (*intdoment.IamGroupAuthorizations, error)
	RepoIamGroupAuthorizationsFindOneActiveRule(ctx context.Context, iamCredentialID uuid.UUID, groupRuleAuthorizationID uuid.UUID, columns []string) (*intdoment.IamGroupAuthorizations, error)
	RepoIamGroupAuthorizationsDeleteOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.IamGroupAuthorizations) error
	RepoIamGroupAuthorizationsInsertOne(ctx context.Context, iamAuthRule *intdoment.IamAuthorizationRule, datum *intdoment.IamGroupAuthorizations, columns []string) (*intdoment.IamGroupAuthorizations, error)
	RepoGroupRuleAuthorizationsFindActiveOneByID(ctx context.Context, groupRuleAuthorizationID uuid.UUID, columns []string) (*intdoment.GroupRuleAuthorizations, error)
	RepoGroupRuleAuthorizationsFindOneByIamGroupAuthorizationID(ctx context.Context, iamGroupAuthorizationID uuid.UUID, columns []string) (*intdoment.GroupRuleAuthorizations, error)
}

type RouteIamGroupAuthorizationsApiCoreService interface {
	ServiceIamGroupAuthorizationsDeleteMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		verboseResponse bool,
		data []*intdoment.IamGroupAuthorizations,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceIamGroupAuthorizationsInsertMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		verboseResponse bool,
		data []*intdoment.IamGroupAuthorizations,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceIamGroupAuthorizationsSearch(
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
	ServiceIamGroupAuthorizationsGetMetadataModel(ctx context.Context, metadataModelRetrieve MetadataModelRetrieve, targetJoinDepth int) (map[string]any, error)
	ServiceDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID) (*intdoment.DirectoryGroups, error)
}
