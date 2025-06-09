package http

import (
	"net/http"

	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intrepopostgres "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/repository/postgres"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

type WebService struct {
	Logger             intdomint.Logger
	OpenID             intdomint.OpenID
	Env                *intlib.EnvVariables
	PostgresRepository *intrepopostgres.PostrgresRepository
	IamCookie          http.Cookie
	FileService        intdomint.FileService
}
