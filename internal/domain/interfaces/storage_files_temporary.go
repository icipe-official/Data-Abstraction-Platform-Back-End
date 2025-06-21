package interfaces

import (
	"context"
	"io"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

type StorageFilesTemporaryRepository interface {
	RepoStorageFilesDeleteTemporaryFiles(ctx context.Context, fileService FileService) (*intdoment.StorageFilesTemporaryDelete, error)

	RepoStorageFilesTemporaryDeleteOne(
		ctx context.Context,
		iamAuthRule *intdoment.IamAuthorizationRule,
		fileService FileService,
		datum *intdoment.StorageFilesTemporary,
	) error

	RepoStorageFilesTemporaryInsertOne(
		ctx context.Context,
		fileService FileService,
		datum *intdoment.StorageFilesTemporary,
		file io.Reader,
		columns []string,
	) (*intdoment.StorageFilesTemporary, error)

	RepoStorageFilesTemporarySearch(
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

type StorageFilesTemporaryService interface {
	ServiceStorageFilesTemporaryDelete(ctx context.Context, fileService FileService)
}
