package routers

import (
	"github.com/go-chi/chi/v5"

	inthttp "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http"
	"github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/abstractions"
	abstractionsdirectorygroups "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/abstractions/directory-groups"
	abstractionsreviews "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/abstractions/reviews"
	abstractionsreviewscomments "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/abstractions/reviews/comments"
	"github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/directory"
	directorygroups "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/directory/groups"
	authorizationrules "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/group/authorization-rules"
	ruleauthorizations "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/group/rule-authorizations"
	"github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/iam"
	"github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/iam/credentials"
	groupauthorizations "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/iam/group-authorizations"
	metadatamodel "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/metadata-model"
	metadatamodels "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/metadata-models"
	metadatamodelsdirectory "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/metadata-models/directory"
	metadatamodelsdirectorygroups "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/metadata-models/directory/groups"
	redirect "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/redirect"
	storagefiles "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/http/routes/storage/files"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func InitApiCoreRouter(router *chi.Mux, webService *inthttp.WebService) {
	router.Route(webService.Env.Get(intlib.ENV_WEB_SERVICE_BASE_PATH), func(baseRouter chi.Router) {
		baseRouter.Mount("/redirect", redirect.ApiCoreRouter(webService))

		baseRouter.Route("/iam", func(iamRouter chi.Router) {
			iamRouter.Mount("/credentials", credentials.ApiCoreRouter(webService))
			iamRouter.Mount("/group-authorizations", groupauthorizations.ApiCoreRouter(webService))
			iamRouter.Mount("/", iam.ApiCoreRouter(webService))
		})

		baseRouter.Route("/", func(authedRouter chi.Router) {
			authedRouter.Use(intlib.IamAuthenticationMiddleware(webService.Logger, webService.Env, webService.OpenID, webService.IamCookie, webService.PostgresRepository))

			authedRouter.Route("/directory", func(directoryRouter chi.Router) {
				directoryRouter.Mount("/groups", directorygroups.ApiCoreRouter(webService))
				directoryRouter.Mount("/", directory.ApiCoreRouter(webService))
			})
			authedRouter.Mount("/metadata-model", metadatamodel.ApiCoreRouter(webService))
			authedRouter.Route("/metadata-models", func(metadataModelsRouter chi.Router) {
				metadataModelsRouter.Route("/directory", func(directoryRouter chi.Router) {
					directoryRouter.Mount("/groups", metadatamodelsdirectorygroups.ApiCoreRouter(webService))
					directoryRouter.Mount("/", metadatamodelsdirectory.ApiCoreRouter(webService))
				})
				metadataModelsRouter.Mount("/", metadatamodels.ApiCoreRouter(webService))
			})
			authedRouter.Route("/group", func(groupRouter chi.Router) {
				groupRouter.Mount("/rule-authorizations", ruleauthorizations.ApiCoreRouter(webService))
				groupRouter.Mount("/authorization-rules", authorizationrules.ApiCoreRouter(webService))
			})
			authedRouter.Route("/storage", func(storageRouter chi.Router) {
				storageRouter.Mount("/files", storagefiles.ApiCoreRouter(webService))
			})
			authedRouter.Route("/abstractions", func(abstractionsRouter chi.Router) {
				abstractionsRouter.Route("/reviews", func(reviewsRouter chi.Router) {
					reviewsRouter.Mount("/comments", abstractionsreviewscomments.ApiCoreRouter(webService))
					reviewsRouter.Mount("/", abstractionsreviews.ApiCoreRouter(webService))
				})
				abstractionsRouter.Mount("/directory-groups", abstractionsdirectorygroups.ApiCoreRouter(webService))
				abstractionsRouter.Mount("/", abstractions.ApiCoreRouter(webService))
			})
		})
	})
}
