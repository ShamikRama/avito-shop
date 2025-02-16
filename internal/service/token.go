package service

import (
	"avito-shop/internal/erorrs"
	"fmt"
	"github.com/golang-jwt/jwt"
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId int
}

func ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken,
		&tokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, erorrs.ErrSigningMethod
			}

			return []byte("qrkjk#4#%35FSFJlja#4353KSFjH"), nil
		})
	if err != nil {
		return 0, fmt.Errorf("error parse token %s", err)
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, fmt.Errorf("invalid type of token claims %s", err)
	}

	return claims.UserId, nil
}
