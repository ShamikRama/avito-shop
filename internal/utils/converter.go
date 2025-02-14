package utils

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/model"
)

func AuthRequestToUser(dto model.AuthRequestDTO) domain.User {
	return domain.User{
		Username:     dto.Username,
		PasswordHash: dto.Password,
	}
}
