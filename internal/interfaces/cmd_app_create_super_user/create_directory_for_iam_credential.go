package cmdappcreatesuperuser

import (
	"context"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

func (n *CmdCreateSuperUserService) ServiceCreateDirectoryForIamCredentials(ctx context.Context, iamCredential *intdoment.IamCredentials) error {
	return n.repo.RepoDirectoryInsertOneAndUpdateIamCredentials(ctx, iamCredential)
}
