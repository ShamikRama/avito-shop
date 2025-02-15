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

type ServiceUserInterface interface {
	SendCoinToUser(ctx context.Context, fromUserID int, toUserName string, amount int) error
	BuyItem(ctx context.Context, userID int, input model.BuyItemRequestDTO) error
	GetUserInfo(ctx context.Context, userID int) (model.InfoResponseDTO, error)
}

func (r *Api) SendCoin(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		r.logger.Error("unidentified user")
		c.JSON(http.StatusUnauthorized, model.ErrorResponseDTO{"пользователь не авторизован"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var input model.SendCoinRequestDTO

	err = c.ShouldBindJSON(&input)
	if err != nil {
		r.logger.Error("error bind json")
		c.JSON(http.StatusBadRequest, model.ErrorResponseDTO{"неверные данные для ввода"})
		return
	}

	err = r.user.SendCoinToUser(ctx, userId, input.ToUser, input.Amount)
	if err != nil {
		switch {
		case errors.Is(err, erorrs.ErrSelfTransfer):
			c.JSON(http.StatusBadRequest, model.ErrorResponseDTO{"нельзя отправить самому себе"})
		case errors.Is(err, erorrs.ErrNotFound):
			c.JSON(http.StatusBadRequest, model.ErrorResponseDTO{"пользователь не найден"})
		case errors.Is(err, erorrs.ErrInsufficientFunds):
			c.JSON(http.StatusBadRequest, model.ErrorResponseDTO{"недостаточно средств"})
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponseDTO{"внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, "деньги успешно отправлены")

}

func (r *Api) BuyItem(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		r.logger.Error("unidentified user")
		c.JSON(http.StatusUnauthorized, model.ErrorResponseDTO{"пользователь не авторизован"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var input model.BuyItemRequestDTO

	err = c.ShouldBindUri(&input)
	if err != nil {
		r.logger.Error("error bind json")
		c.JSON(http.StatusBadRequest, model.ErrorResponseDTO{"неверные данные для ввода"})
		return
	}

	err = r.user.BuyItem(ctx, userID, input)
	if err != nil {
		switch {
		case errors.Is(err, erorrs.ErrNotFound):
			r.logger.Error("user not found", zap.String("item", input.Item))
			c.JSON(http.StatusNotFound, model.ErrorResponseDTO{Error: "товар не найден"})
		case errors.Is(err, erorrs.ErrNotFound):
			c.JSON(http.StatusBadRequest, model.ErrorResponseDTO{"пользователь не найден"})
		case errors.Is(err, erorrs.ErrInsufficientFunds):
			c.JSON(http.StatusBadRequest, model.ErrorResponseDTO{"недостаточно средств"})
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponseDTO{"внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, "товар успешно куплен")

}

func (r *Api) GetUserInfo(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		r.logger.Error("unidentified user")
		c.JSON(http.StatusUnauthorized, model.ErrorResponseDTO{Error: "пользователь не авторизован"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userInfo, err := r.user.GetUserInfo(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, erorrs.ErrNotFound):
			r.logger.Error("user not found", zap.Int("userID", userID))
			c.JSON(http.StatusNotFound, model.ErrorResponseDTO{Error: "пользователь не найден"})
		case errors.Is(err, erorrs.ErrItemNotFound):
			r.logger.Error("no items found for user", zap.Int("userID", userID))
			c.JSON(http.StatusOK, userInfo)
		default:
			r.logger.Error("failed to get user info", zap.Int("userID", userID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponseDTO{Error: "внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, userInfo)
}
