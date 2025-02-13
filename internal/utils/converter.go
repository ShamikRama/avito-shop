package utils

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/model"
)

// AuthRequestDTO -> domain.User (для регистрации)
func AuthRequestToUser(dto model.AuthRequestDTO) domain.User {
	return domain.User{
		Username:     dto.Username,
		PasswordHash: dto.Password,
	}
}

// domain.User -> AuthResponseDTO
func UserToAuthResponse(user domain.User, token string) model.AuthResponseDTO {
	return model.AuthResponseDTO{
		Token: token,
	}
}
