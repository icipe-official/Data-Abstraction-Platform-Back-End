package metadatamodel

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"

	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	inthttp "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func ApiCoreRouter(webService *inthttp.WebService) *chi.Mux {
	router := chi.NewRouter()

	router.Use(intlib.IamAuthenticationMiddleware(webService.Logger, webService.Env, webService.OpenID, webService.IamCookie, webService.PostgresRepository))

	router.Get("/{metadata_model_action_id}/{directory_group_id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

		metadataModelRepoName := chi.URLParam(r, "metadata_model_action_id")
		if len(metadataModelRepoName) == 0 {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		directoryGroupId, err := uuid.FromString(chi.URLParam(r, "directory_group_id"))
		if err != nil || directoryGroupId.IsNil() {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		s := initApiCoreService(ctx, webService)
		if s == nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
			return
		}

		metadataModel, err := s.ServiceMetadataModelGet(ctx, metadataModelRepoName, directoryGroupId)
		if err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		}

		intlib.SendJsonResponse(http.StatusOK, metadataModel, w)
		webService.Logger.Log(ctx, slog.LevelInfo, "default metadata-model retrieved", ctx.Value(intlib.LOG_ATTR_CTX_KEY))
	})

	return router
}

func initApiCoreService(ctx context.Context, webService *inthttp.WebService) intdomint.RouteMetadataModelApiService {
	if value, err := NewService(webService); err != nil {
		errmsg := fmt.Errorf("initialize api core service failed, error: %v", err)
		if value.logger != nil {
			value.logger.Log(ctx, slog.LevelError, errmsg.Error())
		} else {
			log.Println(errmsg)
		}

		return nil
	} else {
		return value
	}
}
