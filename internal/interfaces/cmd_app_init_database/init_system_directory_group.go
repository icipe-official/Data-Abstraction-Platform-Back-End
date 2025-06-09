package initdatabase

import (
	"context"
	"log"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func (n *CmdInitDatabaseService) ServiceInitSystemDirectoryGroup(ctx context.Context) error {
	systemGroup, err := n.repo.RepoDirectoryGroupsFindSystemGroup(ctx, []string{intdoment.DirectoryGroupsRepository().ID})
	if err != nil {
		return intlib.FunctionNameAndError(n.ServiceInitSystemDirectoryGroup, err)
	}
	if systemGroup != nil {
		return nil
	}

	systemGroup, err = n.repo.RepoDirectoryGroupsCreateSystemGroup(ctx, []string{intdoment.DirectoryGroupsRepository().ID})
	if err != nil {
		return intlib.FunctionNameAndError(n.ServiceInitSystemDirectoryGroup, err)
	}
	log.Printf("System Group with id %s created", systemGroup.ID[0])

	return nil
}
