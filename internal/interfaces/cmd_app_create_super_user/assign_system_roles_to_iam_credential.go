package cmdappcreatesuperuser

import (
	"context"
	"errors"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *CmdCreateSuperUserService) ServiceAssignSystemRolesToIamCredential(ctx context.Context, iamCredential *intdoment.IamCredentials) (int, error) {
	systemGroupRuleAuthorizations, err := n.repo.RepoDirectoryGroupsFindSystemGroupRuleAuthorizations(ctx)
	if err != nil {
		return 0, intlib.FunctionNameAndError(n.ServiceAssignSystemRolesToIamCredential, err)
	}

	if len(systemGroupRuleAuthorizations) < 1 {
		return 0, intlib.FunctionNameAndError(n.ServiceAssignSystemRolesToIamCredential, errors.New("systemGroupRuleAuthorizations is empty"))
	}

	successfulInserts, err := n.repo.RepoIamGroupAuthorizationsSystemRolesInsertMany(ctx, iamCredential, systemGroupRuleAuthorizations)
	if err != nil {
		return 0, intlib.FunctionNameAndError(n.ServiceAssignSystemRolesToIamCredential, err)
	}

	return successfulInserts, nil
}
