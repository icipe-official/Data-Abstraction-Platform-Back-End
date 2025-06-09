package interfaces

import (
	"context"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

type RouteAbstractionsReviewsCommentsRepository interface {
	RepoAbstractionsReviewsCommentsInsertOne(
		ctx context.Context,
		abstractionsReviews *intdoment.AbstractionsReviewsComments,
		columns []string,
	) (*intdoment.AbstractionsReviewsComments, error)
	RepoAbstractionsFindActiveOneByIDAndAbstractionsDirectoryGroups(
		ctx context.Context,
		id uuid.UUID,
		abstractionsDirectoryGroupsID uuid.UUID,
		columns []string,
	) (*intdoment.Abstractions, error)
	RepoAbstractionsReviewsCommentsSearch(
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
}

type RouteAbstractionsReviewsCommentsApiCoreService interface {
	ServiceAbstractionsReviewsCommentsInsertMany(
		ctx context.Context,
		iamCredential *intdoment.IamCredentials,
		iamAuthorizationRules *intdoment.IamAuthorizationRules,
		authContextDirectoryGroupID uuid.UUID,
		verboseResponse bool,
		data []*intdoment.AbstractionsReviewsComments,
	) (int, *intdoment.MetadataModelVerbRes, error)
	ServiceAbstractionsReviewsCommentsSearch(
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
	ServiceAbstractionsReviewsCommentsGetMetadataModel(ctx context.Context, metadataModelRetrieve MetadataModelRetrieve, targetJoinDepth int) (map[string]any, error)
	ServiceDirectoryGroupsFindOneByIamCredentialID(ctx context.Context, iamCredentialID uuid.UUID) (*intdoment.DirectoryGroups, error)
}
