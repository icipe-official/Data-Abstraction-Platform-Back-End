package groups

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intfieldanymm "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/field_any_metadata_model"
	intmmretrieve "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/metadata_model_retrieve"
	intwebservice "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/web_service"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func ApiCoreRouter(webService *intwebservice.WebService) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/delete", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

		authedIamCredential, err := intlib.IamHttpRequestCtxGetAuthedIamCredential(r)
		if err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		}

		data := make([]*intdoment.DirectoryGroups, 0)
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		s := initApiCoreService(ctx, webService)
		if s == nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
			return
		}

		verboseResponse := intlib.UrlSearchParamGetBool(r, intlib.URL_SEARCH_PARAM_VERBOSE_RESPONSE, false)

		var authContextDirectoryGroupID uuid.UUID
		if value, err := intlib.UrlSearchParamGetUuid(r, intlib.URL_SEARCH_PARAM_AUTH_CONTEXT_DIRECTORY_GROUP_ID); err != nil {
			if directoryGroup, err := s.ServiceDirectoryGroupsFindOneByIamCredentialID(ctx, authedIamCredential.ID[0]); err != nil {
				intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
				return
			} else {
				if directoryGroup == nil {
					intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
					return
				}
				authContextDirectoryGroupID = directoryGroup.ID[0]
			}
		} else {
			authContextDirectoryGroupID = value
		}

		if code, verbres, err := s.ServiceDirectoryGroupsDeleteMany(
			ctx,
			authedIamCredential,
			nil,
			authContextDirectoryGroupID,
			verboseResponse,
			data,
		); err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		} else {
			intlib.SendJsonResponse(code, verbres, w)
			webService.Logger.Log(
				ctx,
				slog.LevelInfo+1,
				intlib.LogAction(intlib.LOG_ACTION_DELETE, intdoment.DirectoryGroupsRepository().RepositoryName),
				ctx.Value(intlib.LOG_ATTR_CTX_KEY),
				"authenicated iam credential",
				intlib.JsonStringifyMust(authedIamCredential),
				"verbose response data",
				intlib.JsonStringifyMust(verbres.MetadataModelVerboseResponse.Data),
			)
		}
	})

	router.Post("/update", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

		authedIamCredential, err := intlib.IamHttpRequestCtxGetAuthedIamCredential(r)
		if err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		}

		data := make([]*intdoment.DirectoryGroups, 0)
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		s := initApiCoreService(ctx, webService)
		if s == nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
			return
		}

		verboseResponse := intlib.UrlSearchParamGetBool(r, intlib.URL_SEARCH_PARAM_VERBOSE_RESPONSE, false)

		var authContextDirectoryGroupID uuid.UUID
		if value, err := intlib.UrlSearchParamGetUuid(r, intlib.URL_SEARCH_PARAM_AUTH_CONTEXT_DIRECTORY_GROUP_ID); err != nil {
			if directoryGroup, err := s.ServiceDirectoryGroupsFindOneByIamCredentialID(ctx, authedIamCredential.ID[0]); err != nil {
				intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
				return
			} else {
				if directoryGroup == nil {
					intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
					return
				}
				authContextDirectoryGroupID = directoryGroup.ID[0]
			}
		} else {
			authContextDirectoryGroupID = value
		}

		if code, verbres, err := s.ServiceDirectoryGroupsUpdateMany(
			ctx,
			authedIamCredential,
			nil,
			authContextDirectoryGroupID,
			intfieldanymm.NewFieldAnyMetadataModelGet(webService.Logger, webService.PostgresRepository, nil),
			verboseResponse,
			data,
		); err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		} else {
			intlib.SendJsonResponse(code, verbres, w)
			webService.Logger.Log(
				ctx,
				slog.LevelInfo+1,
				intlib.LogAction(intlib.LOG_ACTION_UPDATE, intdoment.DirectoryGroupsRepository().RepositoryName),
				ctx.Value(intlib.LOG_ATTR_CTX_KEY),
				"authenicated iam credential",
				intlib.JsonStringifyMust(authedIamCredential),
				"verbose response data",
				intlib.JsonStringifyMust(verbres.MetadataModelVerboseResponse.Data),
			)
		}
	})

	router.Post("/create", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

		authedIamCredential, err := intlib.IamHttpRequestCtxGetAuthedIamCredential(r)
		if err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		}

		data := make([]*intdoment.DirectoryGroups, 0)
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		}

		s := initApiCoreService(ctx, webService)
		if s == nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
			return
		}

		verboseResponse := intlib.UrlSearchParamGetBool(r, intlib.URL_SEARCH_PARAM_VERBOSE_RESPONSE, false)

		var authContextDirectoryGroupID uuid.UUID
		if value, err := intlib.UrlSearchParamGetUuid(r, intlib.URL_SEARCH_PARAM_AUTH_CONTEXT_DIRECTORY_GROUP_ID); err != nil {
			if directoryGroup, err := s.ServiceDirectoryGroupsFindOneByIamCredentialID(ctx, authedIamCredential.ID[0]); err != nil {
				intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
				return
			} else {
				if directoryGroup == nil {
					intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
					return
				}
				authContextDirectoryGroupID = directoryGroup.ID[0]
			}
		} else {
			authContextDirectoryGroupID = value
		}

		if code, verbres, err := s.ServiceDirectoryGroupsInsertMany(
			ctx,
			authedIamCredential,
			nil,
			authContextDirectoryGroupID,
			intfieldanymm.NewFieldAnyMetadataModelGet(webService.Logger, webService.PostgresRepository, nil),
			verboseResponse,
			data,
		); err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		} else {
			intlib.SendJsonResponse(code, verbres, w)
			webService.Logger.Log(
				ctx,
				slog.LevelInfo+1,
				intlib.LogAction(intlib.LOG_ACTION_CREATE, intdoment.DirectoryGroupsRepository().RepositoryName),
				ctx.Value(intlib.LOG_ATTR_CTX_KEY),
				"authenicated iam credential",
				intlib.JsonStringifyMust(authedIamCredential),
				"verbose response data",
				intlib.JsonStringifyMust(verbres.MetadataModelVerboseResponse.Data),
			)
		}
	})

	router.Route("/search", func(searchRouter chi.Router) {
		searchRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

			authedIamCredential, err := intlib.IamHttpRequestCtxGetAuthedIamCredential(r)
			if err != nil {
				intlib.SendJsonErrorResponse(err, w)
				return
			}

			mmSearch := new(intdoment.MetadataModelSearch)
			json.NewDecoder(r.Body).Decode(mmSearch)

			s := initApiCoreService(ctx, webService)
			if s == nil {
				intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
				return
			}

			var authContextDirectoryGroupID uuid.UUID
			if value, err := intlib.UrlSearchParamGetUuid(r, intlib.URL_SEARCH_PARAM_AUTH_CONTEXT_DIRECTORY_GROUP_ID); err != nil {
				if directoryGroup, err := s.ServiceDirectoryGroupsFindOneByIamCredentialID(ctx, authedIamCredential.ID[0]); err != nil {
					intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
					return
				} else {
					if directoryGroup == nil {
						intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
						return
					}
					authContextDirectoryGroupID = directoryGroup.ID[0]
				}
			} else {
				authContextDirectoryGroupID = value
			}

			mmSearchModelWasEmpty := false
			if mmSearch.MetadataModel == nil {
				if value, err := s.ServiceDirectoryGroupsGetMetadataModel(
					ctx,
					intmmretrieve.NewMetadataModelRetrieve(webService.Logger, webService.PostgresRepository, authContextDirectoryGroupID, authedIamCredential, nil),
					1,
				); err != nil {
					intlib.SendJsonErrorResponse(err, w)
					return
				} else {
					mmSearch.MetadataModel = value
					mmSearchModelWasEmpty = true
				}
			}

			var startSearchDirectoryGroupID uuid.UUID
			if value, err := intlib.UrlSearchParamGetUuid(r, intlib.URL_SEARCH_PARAM_START_SEARCH_DIRECTORY_GROUP_ID); err == nil {
				startSearchDirectoryGroupID = value
			} else {
				startSearchDirectoryGroupID = authContextDirectoryGroupID
			}

			skipIfDataExtraction := intlib.UrlSearchParamGetBool(r, intlib.URL_SEARCH_PARAM_SKIP_IF_DATA_EXTRACTION, true)
			skipIfFGDisabled := intlib.UrlSearchParamGetBool(r, intlib.URL_SEARCH_PARAM_SKIP_IF_FG_DISABLED, true)
			whereAfterJoin := intlib.UrlSearchParamGetBool(r, intlib.URL_SEARCH_PARAM_WHERE_AFTER_JOIN, false)

			if searchResults, err := s.ServiceDirectoryGroupsSearch(
				ctx,
				mmSearch,
				webService.PostgresRepository,
				authedIamCredential,
				nil,
				startSearchDirectoryGroupID,
				authContextDirectoryGroupID,
				skipIfFGDisabled,
				skipIfDataExtraction,
				whereAfterJoin,
			); err != nil {
				intlib.SendJsonErrorResponse(err, w)
				return
			} else {
				if !mmSearchModelWasEmpty {
					searchResults.MetadataModel = nil
				}
				intlib.SendJsonResponse(http.StatusOK, searchResults, w)
				webService.Logger.Log(
					ctx,
					slog.LevelInfo+1,
					fmt.Sprintln(
						intdoment.DirectoryGroupsRepository().RepositoryName,
						"searched successfully",
					),
					ctx.Value(intlib.LOG_ATTR_CTX_KEY),
					"authenicated iam credential",
					intlib.JsonStringifyMust(authedIamCredential),
				)
			}
		})

		searchRouter.Get("/metadata-model", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

			authedIamCredential, err := intlib.IamHttpRequestCtxGetAuthedIamCredential(r)
			if err != nil {
				intlib.SendJsonErrorResponse(err, w)
				return
			}

			s := initApiCoreService(ctx, webService)
			if s == nil {
				intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
				return
			}

			var authContextDirectoryGroupID uuid.UUID
			if value, err := intlib.UrlSearchParamGetUuid(r, intlib.URL_SEARCH_PARAM_AUTH_CONTEXT_DIRECTORY_GROUP_ID); err != nil {
				if directoryGroup, err := s.ServiceDirectoryGroupsFindOneByIamCredentialID(ctx, authedIamCredential.ID[0]); err != nil {
					intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
					return
				} else {
					if directoryGroup == nil {
						intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
						return
					}
					authContextDirectoryGroupID = directoryGroup.ID[0]
				}
			} else {
				authContextDirectoryGroupID = value
			}

			targetJoinDepth := 1
			if value, err := intlib.UrlSearchParamGetInt(r, intlib.URL_SEARCH_PARAM_TARGET_JOIN_DEPTH); err == nil {
				targetJoinDepth = value
			} else {
				targetJoinDepth = 1
			}

			if value, err := s.ServiceDirectoryGroupsGetMetadataModel(
				ctx,
				intmmretrieve.NewMetadataModelRetrieve(webService.Logger, webService.PostgresRepository, authContextDirectoryGroupID, authedIamCredential, nil),
				targetJoinDepth,
			); err != nil {
				intlib.SendJsonErrorResponse(err, w)
				return
			} else {
				intlib.SendJsonResponse(http.StatusOK, value, w)
				webService.Logger.Log(
					ctx,
					slog.LevelInfo+1,
					fmt.Sprintln(
						intdoment.DirectoryGroupsRepository().RepositoryName,
						"metadata-model successfully retrieved",
					),
					ctx.Value(intlib.LOG_ATTR_CTX_KEY),
					"authenicated iam credential",
					intlib.JsonStringifyMust(authedIamCredential),
				)
			}
		})
	})

	return router
}

func initApiCoreService(ctx context.Context, webService *intwebservice.WebService) intdomint.RouteDirectoryGroupsApiCoreService {
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
