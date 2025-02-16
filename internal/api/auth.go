package api

import (
	"avito-shop/internal/erorrs"
	"avito-shop/internal/model"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceAuthInterface
type ServiceAuthInterface interface {
	Authorization(ctx context.Context, dto model.AuthRequestDTO) (string, error)
	GenerateJwtToken(userId int) (string, error)
}

func (r *Api) Auth(c *gin.Context) {
	var input model.AuthRequestDTO

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := c.ShouldBindJSON(&input)
	if err != nil {
		r.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.ErrorResponseDTO{Error: "неправильное данные для ввода"})
		return
	}

	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponseDTO{Error: "поля имя и пароль обязательны к заполнению"})
		return
	}

	token, err := r.auth.Authorization(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, erorrs.ErrNotFound):
			c.JSON(http.StatusUnauthorized, model.ErrorResponseDTO{Error: "неправильны данные для входа"})
		case errors.Is(err, erorrs.ErrUserExist):
			c.JSON(http.StatusBadRequest, model.ErrorResponseDTO{Error: "пользователь уже существует"})
		default:
			r.logger.Error("Authentication error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponseDTO{Error: "внутренняя ошибка сервера"})
		}
		return
	}

	c.Header("Authorization", "Bearer "+token)

	c.JSON(http.StatusOK, token)
}
