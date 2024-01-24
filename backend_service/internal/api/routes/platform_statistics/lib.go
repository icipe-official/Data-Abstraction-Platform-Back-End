package platformstatistics

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"

	"github.com/google/uuid"
)

const currentSection = "Stats"

type platformstats struct {
	CurrentUser   lib.User
	ProjectID     uuid.UUID
	ProjectStats  model.ProjectStatistics
	ProjectsStats []model.ProjectStatistics
	PlatformStats model.PlatformStatistics
}
