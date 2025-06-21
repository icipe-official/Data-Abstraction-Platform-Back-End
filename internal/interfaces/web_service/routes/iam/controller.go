package iam

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intwebservice "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/web_service"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func ApiCoreRouter(webService *intwebservice.WebService) *chi.Mux {
	router := chi.NewRouter()

	router.Get("/authentication-headers", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

		s := initApiCoreService(ctx, webService)
		if s == nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
			return
		}

		iamAuthenticationHeaders := s.ServiceGetIamAuthenticationHeaders(ctx, webService.IamCookie)

		intlib.SendJsonResponse(http.StatusOK, iamAuthenticationHeaders, w)
		webService.Logger.Log(ctx, slog.LevelInfo, "Get Authentication Headers", ctx.Value(intlib.LOG_ATTR_CTX_KEY))
	})

	router.Get("/openid-endpoints", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

		s := initApiCoreService(ctx, webService)
		if s == nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
			return
		}

		iamOpenidEndpoints := s.ServiceGetIamOpenIDEndpoints(ctx, webService.OpenID)

		intlib.SendJsonResponse(http.StatusOK, iamOpenidEndpoints, w)
		webService.Logger.Log(ctx, slog.LevelInfo, "Get OpenID Endpoints", ctx.Value(intlib.LOG_ATTR_CTX_KEY))
	})

	router.Get("/sign-in", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

		s := initApiCoreService(ctx, webService)
		if s == nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
			return
		}

		openIDToken := new(intdoment.OpenIDToken)
		openIDToken.AccessToken = r.Header.Get(intdoment.OPENID_HEADER_ACCESS_TOKEN)
		if value, err := strconv.Atoi(r.Header.Get(intdoment.OPENID_HEADER_ACCESS_TOKEN_EXPIRES_IN)); err != nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return

		} else {
			openIDToken.ExpiresIn = int64(value)
		}
		openIDToken.RefreshToken = r.Header.Get(intdoment.OPENID_HEADER_REFRESH_TOKEN)
		if value, err := strconv.Atoi(r.Header.Get(intdoment.OPENID_HEADER_REFRESH_TOKEN_EXPIRES_IN)); err != nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			openIDToken.RefreshExpiresIn = int64(value)
		}

		openIDTokenIntrospect, err := s.ServiceOpenIDIntrospectToken(ctx, webService.OpenID, openIDToken)
		if err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		}

		iamCredential, err := s.ServiceGetIamCredentialsByOpenIDSub(ctx, openIDTokenIntrospect)
		if err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		}

		if token, err := intlib.IamPrepOpenIDTokenForClient(webService.Env, openIDToken); err != nil {
			webService.Logger.Log(ctx, slog.LevelError, fmt.Sprintf("Prepare access refresh token for client failed, error: %v", err), ctx.Value(intlib.LOG_ATTR_CTX_KEY))
			if err := s.ServiceOpenIDRevokeToken(ctx, webService.OpenID, openIDToken); err != nil {
				intlib.SendJsonErrorResponse(err, w)
				return
			}
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
		} else {
			intlib.IamSetCookieInResponse(w, webService.IamCookie, token, int(openIDToken.ExpiresIn), int(openIDToken.RefreshExpiresIn))
			intlib.SendJsonResponse(http.StatusOK, iamCredential, w)
			webService.Logger.Log(ctx, slog.LevelInfo+2, fmt.Sprintf("sign in by %v", iamCredential.ID), ctx.Value(intlib.LOG_ATTR_CTX_KEY))
		}
	})

	router.Route("/", func(authedRoutes chi.Router) {
		authedRoutes.Use(intlib.IamAuthenticationMiddleware(webService.Logger, webService.Env, webService.OpenID, webService.IamCookie, webService.PostgresRepository))

		authedRoutes.Get("/refresh-tokens", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

			authedIamCredential, _ := intlib.IamHttpRequestCtxGetAuthedIamCredential(r)

			if newAccessRefreshToken, err := intlib.IamHttpRequestCtxGetNewAccessRefreshToken(r); err != nil {
				intlib.SendJsonErrorResponse(err, w)
			} else {
				intlib.SendJsonResponse(http.StatusOK, newAccessRefreshToken, w)
				webService.Logger.Log(ctx, slog.LevelInfo, fmt.Sprintf("Refresh token by %v", authedIamCredential.ID), ctx.Value(intlib.LOG_ATTR_CTX_KEY))
			}
		})

		authedRoutes.Get("/session", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

			s := initApiCoreService(ctx, webService)
			if s == nil {
				intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
				return
			}

			authedIamCredential, _ := intlib.IamHttpRequestCtxGetAuthedIamCredential(r)

			iamSessionData, err := s.ServiceGetIamSession(ctx, webService.OpenID, authedIamCredential)
			if err != nil {
				intlib.SendJsonErrorResponse(err, w)
				return
			}

			intlib.SendJsonResponse(http.StatusOK, iamSessionData, w)
			logMsg := "get session data"
			if authedIamCredential != nil {
				logMsg += fmt.Sprintf(" by %v", authedIamCredential.ID)
			}
			webService.Logger.Log(ctx, slog.LevelInfo, logMsg, ctx.Value(intlib.LOG_ATTR_CTX_KEY))
		})

		authedRoutes.Get("/sign-out", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

			authedIamCredential, err := intlib.IamHttpRequestCtxGetAuthedIamCredential(r)
			if err != nil {
				intlib.SendJsonErrorResponse(err, w)
				return
			}

			intlib.SendJsonResponse(http.StatusOK, authedIamCredential, w)
			webService.Logger.Log(ctx, slog.LevelInfo+2, fmt.Sprintf("sign out by %v", authedIamCredential.ID), ctx.Value(intlib.LOG_ATTR_CTX_KEY))
		})
	})

	return router
}

func initApiCoreService(ctx context.Context, webService *intwebservice.WebService) intdomint.RouteIamService {
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
