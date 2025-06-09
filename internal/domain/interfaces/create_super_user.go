package interfaces

import (
	"context"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

type CreateSuperUserRepository interface {
	// Will return error if more than one row is returned.
	// Parameters:
	//
	// - columnfields - columns/field data to obtain. Leave empty or nil to get all columns/fields
	RepoIamCredentialsFindOneByID(ctx context.Context, columnField string, value any, columnfields []string) (*intdoment.IamCredentials, error)
	// Parameters:
	//
	// - columnfields - columns/field data to obtain. Leave empty or nil to get all columns/fields
	RepoDirectoryGroupsFindSystemGroup(ctx context.Context, columnfields []string) (*intdoment.DirectoryGroups, error)
	RepoDirectoryGroupsFindSystemGroupRuleAuthorizations(ctx context.Context) ([]intdoment.GroupRuleAuthorizations, error)
	// returns no of successful inserts.
	//
	// will not check if role exists as it only anticipates finding intdoment.GroupRuleAuthorization.ID having been set.
	RepoIamGroupAuthorizationsSystemRolesInsertMany(ctx context.Context, iamCredenial *intdoment.IamCredentials, groupRuleAuthorizations []intdoment.GroupRuleAuthorizations) (int, error)
	RepoDirectoryInsertOneAndUpdateIamCredentials(ctx context.Context, iamCredential *intdoment.IamCredentials) error
}

type CreateSuperUserService interface {
	ServiceGetIamCredentials(ctx context.Context) (*intdoment.IamCredentials, error)
	// return no of system roles assign to iam credential
	ServiceAssignSystemRolesToIamCredential(ctx context.Context, iamCredential *intdoment.IamCredentials) (int, error)
	ServiceCreateDirectoryForIamCredentials(ctx context.Context, iamCredential *intdoment.IamCredentials) error
}
