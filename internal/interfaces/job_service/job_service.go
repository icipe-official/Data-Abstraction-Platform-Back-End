package jobservice

import (
	"net/http"

	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intrepopostgres "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/repository/postgres"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

type JobService struct {
	Logger             intdomint.Logger
	Env                *intlib.EnvVariables
	PostgresRepository *intrepopostgres.PostrgresRepository
	IamCookie          http.Cookie
	FileService        intdomint.FileService
}
