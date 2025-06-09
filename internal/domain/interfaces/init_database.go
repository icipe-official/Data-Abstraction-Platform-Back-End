package interfaces

import (
	"context"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

type InitDatabaseRepository interface {
	// return no of data upserted successfully.
	RepoGroupAuthorizationRulesUpsertMany(ctx context.Context, data []intdoment.GroupAuthorizationRules) (int, error)
	// Parameters:
	//
	// - columnfields - columns/field data to obtain. Leave empty or nil to get all columns/fields
	RepoDirectoryGroupsFindSystemGroup(ctx context.Context, columnfields []string) (*intdoment.DirectoryGroups, error)
	// Parameters:
	//
	// - columnfields - columns/field data to return after insert. Leave empty or nil to return all columns/fields
	RepoDirectoryGroupsCreateSystemGroup(ctx context.Context, columnfields []string) (*intdoment.DirectoryGroups, error)
}

type InitDatabaseService interface {
	// return no of data upserted successfully.
	ServiceGroupAuthorizationRulesCreate(ctx context.Context) (int, error)
	ServiceInitSystemDirectoryGroup(ctx context.Context) error
}
