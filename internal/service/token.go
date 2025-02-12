package service

import (
	"avito-shop/internal/erorrs"
	"fmt"
	"github.com/golang-jwt/jwt"
	"time"
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId int
}

const (
	signingKey = "qrkjk#4#%35FSFJlja#4353KSFjH"
	tokenTTL   = 7 * time.Hour
)

func generateJwtToken(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		}, userId,
	})

	return token.SignedString([]byte(signingKey))
}

func ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken,
		&tokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, erorrs.ErrSigningMethod
			}

			return []byte(signingKey), nil
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
