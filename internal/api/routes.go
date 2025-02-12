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
		api.POST("/auth", r.Auth) // сделано

		protected := api.Group("", r.UserIdentity) // сделано
		{
			protected.GET("/info")
			protected.POST("/sendCoin", r.SendCoin) // сделано
			protected.POST("/buy/:item", r.BuyItem) // сделано
		}
	}

	return router
}
