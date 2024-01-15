package iam

import (
	"data_administration_platform/internal/api/lib"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func Router() *chi.Mux {
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Use(lib.AuthenticationMiddleware)

		r.Get("/session", func(w http.ResponseWriter, r *http.Request) {
			lib.SendJsonResponse(lib.CtxGetCurrentUser(r), w)
		})

		r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
			currentUser := lib.CtxGetCurrentUser(r)

			if err := lib.CacheDeleteSessionInfo(currentUser.SessionId); err != nil {
				intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Session info %v for %v not deleted from cache | reason: %v", currentUser.SessionId, currentUser.DirectoryID.String(), err))
			}

			http.SetCookie(w, lib.GetCookie("dap", "", 0))
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("Logout %v", currentUser.DirectoryID))
			lib.SendJsonResponse(lib.JsonMessage{Message: "logout successful"}, w)
		})
	})

	router.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		var Login iam

		if err := json.NewDecoder(r.Body).Decode(&Login.Login); err != nil {
			lib.SendErrorResponse(lib.NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)), w)
			return
		}

		currentUser, err := Login.getUserByEmailandPassword()
		if err != nil {
			lib.SendErrorResponse(err, w)
			return
		}

		sessionId := uuid.New().String()
		if err = lib.CacheSetSessionInfo(sessionId, currentUser.DirectoryID.String()); err != nil {
			lib.SendErrorResponse(err, w)
			return
		}

		if err = lib.CacheSetUserInfo(*currentUser); err != nil {
			lib.SendErrorResponse(err, w)
			return
		}

		if arTokenEncrypt, err := lib.GenerateAccessRefreshToken(sessionId); err != nil {
			lib.SendErrorResponse(err, w)
		} else {
			http.SetCookie(w, lib.GetCookie("dap", arTokenEncrypt, lib.ACCESS_REFRESH_TOKEN_AGE))
			lib.SendJsonResponse(currentUser, w)
			intpkglib.Log(intpkglib.LOG_INFO, currentSection, fmt.Sprintf("Login %v", currentUser.DirectoryID))
		}
	})

	return router
}
