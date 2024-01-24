package platformstatistics

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"data_administration_platform/internal/pkg/data_administration_platform/public/view"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"fmt"
	"net/http"

	jet "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

func (n *platformstats) getProjectStats() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	if n.ProjectID != uuid.Nil {
		n.ProjectStats = model.ProjectStatistics{}
		if err := view.ProjectStatistics.SELECT(view.ProjectStatistics.AllColumns).
			WHERE(view.ProjectStatistics.ProjectID.EQ(jet.UUID(n.ProjectID))).
			Query(db, &n.ProjectStats); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get stats for project %v failed | reason: %v", n.ProjectID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get project stats")
		}
	} else {
		n.ProjectsStats = []model.ProjectStatistics{}
		if err := view.ProjectStatistics.SELECT(view.ProjectStatistics.AllColumns).Query(db, &n.ProjectsStats); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get projects stats failed | reason: %v", err))
			return lib.NewError(http.StatusInternalServerError, "Could not get projects stats")
		}
	}

	return nil
}

func (n *platformstats) getPlatfromStats() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return lib.INTERNAL_SERVER_ERROR
	}
	defer db.Close()

	n.PlatformStats = model.PlatformStatistics{}
	if err := view.PlatformStatistics.SELECT(view.PlatformStatistics.AllColumns).Query(db, &n.PlatformStats); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get platform stats failed | reason: %v", err))
		return lib.NewError(http.StatusInternalServerError, "Could not get platform stats")
	}

	return nil
}
