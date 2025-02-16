package api

import (
	"avito-shop/internal/logger"
	"github.com/gin-gonic/gin"
)

type Api struct {
	logger logger.Logger
	auth   ServiceAuthInterface
	user   ServiceUserInterface
}

func NewApi(logger logger.Logger, auth ServiceAuthInterface, user ServiceUserInterface) *Api {
	return &Api{
		logger: logger,
		auth:   auth,
		user:   user,
	}
}

func (r *Api) InitRoutes() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	{
		api.POST("/auth", r.Auth)

		protected := api.Group("", r.UserIdentity)
		{
			protected.GET("/info", r.GetUserInfo)
			protected.POST("/sendCoin", r.SendCoin)
			protected.POST("/buy/:item", r.BuyItem)
		}
	}

	return router
}
