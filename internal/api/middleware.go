package api

import (
	"avito-shop/internal/model"
	"avito-shop/internal/service"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authHeader = "Authorization"
	userCtx    = "user_id"
)

// здесь еще добавить проверку ттл токена
func (r *Api) UserIdentity(c *gin.Context) {
	header := c.GetHeader(authHeader)
	if header == "" {
		r.logger.Info("Empty header")
		c.JSON(http.StatusUnauthorized, model.ErrorResponseDTO{"вы не авторизованы"})
		c.Abort()
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		r.logger.Info("Wrong header")
		c.JSON(http.StatusInternalServerError, model.ErrorResponseDTO{"вы не авторизованы"})
		c.Abort()
		return
	}

	token := headerParts[1]
	if token == "" {
		r.logger.Info("Empty token")
		c.JSON(http.StatusInternalServerError, model.ErrorResponseDTO{"вы не авторизованы"})
		c.Abort()
		return
	}

	userId, err := service.ParseToken(token)
	if err != nil {
		r.logger.Error("Failed to parse the token")
		c.JSON(http.StatusInternalServerError, model.ErrorResponseDTO{"вы не авторизованы"})
		c.Abort()
		return
	}

	c.Set(userCtx, userId)
	c.Next()
}

func getUserId(c *gin.Context) (int, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		return 0, errors.New("user id not found")
	}

	idInt, ok := id.(int)
	if !ok {
		return 0, errors.New("user id not int")
	}

	return idInt, nil

}
