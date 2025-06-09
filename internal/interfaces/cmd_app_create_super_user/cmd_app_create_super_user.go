package cmdappcreatesuperuser

import (
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
)

type CmdCreateSuperUserService struct {
	repo   intdomint.CreateSuperUserRepository
	logger intdomint.Logger
}

func NewCmdCreateSuperUserService(repo intdomint.CreateSuperUserRepository, logger intdomint.Logger) *CmdCreateSuperUserService {
	return &CmdCreateSuperUserService{
		repo:   repo,
		logger: logger,
	}
}
