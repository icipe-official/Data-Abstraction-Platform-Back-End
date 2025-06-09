package home

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	inthttp "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func ApiCoreRouter(webService *inthttp.WebService) *chi.Mux {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), intlib.LOG_ATTR_CTX_KEY, slog.Attr{Key: intlib.LogSectionAttrKey, Value: slog.StringValue(intlib.LogSectionName(r.URL.Path, webService.Env))})

		s := initApiCoreService(ctx, webService)
		if s == nil {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
			return
		}

		openIDRedirectParams := new(intdoment.OpenIDRedirectParams)
		if param := r.URL.Query().Get(redirect_PARAM_SESSION_STATE); len(param) == 0 {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			openIDRedirectParams.SessionState = param
		}
		if param := r.URL.Query().Get(redirect_PARAM_ISS); len(param) == 0 {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			openIDRedirectParams.Iss = param
		}
		if param := r.URL.Query().Get(redirect_PARAM_CODE); len(param) == 0 {
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)), w)
			return
		} else {
			openIDRedirectParams.Code = param
		}

		openIDToken, err := s.ServiceGetOpenIDToken(ctx, webService.OpenID, openIDRedirectParams)
		if err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		}

		openIDUserInfo, err := s.ServiceGetOpenIDUserInfo(ctx, webService.OpenID, openIDToken)
		if err != nil {
			intlib.SendJsonErrorResponse(err, w)
			return
		}

		iamCredential, err := s.ServiceGetIamCredentialsByOpenIDSub(ctx, openIDUserInfo)
		if err != nil {
			if err := s.ServiceOpenIDRevokeToken(ctx, webService.OpenID, openIDToken); err != nil {
				intlib.SendJsonErrorResponse(err, w)
				return
			}
			intlib.SendJsonErrorResponse(intlib.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)), w)
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
			http.Redirect(w, r, webService.Env.Get(intlib.ENV_WEBSITE_URL), http.StatusSeeOther)
			webService.Logger.Log(ctx, slog.LevelInfo+2, fmt.Sprintf("login by %v", iamCredential.ID), ctx.Value(intlib.LOG_ATTR_CTX_KEY))
		}
	})

	return router
}

func initApiCoreService(ctx context.Context, webService *inthttp.WebService) intdomint.RouteRedirectApiCoreService {
	if value, err := NewService(webService); err != nil {
		errmsg := fmt.Errorf("initialize website service failed, error: %v", err)
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
