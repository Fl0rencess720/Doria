package jwtc

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

type AuthClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func GenAccessToken(userID int) (string, error) {
	accessSecret := viper.GetString("JWT_ACCESS_SECRET")

	ac := AuthClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        time.Now().String(),
			Issuer:    "Fl0rencess720",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, ac).SignedString([]byte(accessSecret))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func GenRefreshToken() (string, error) {
	refreshSecret := viper.GetString("JWT_REFRESH_SECRET")

	rc := jwt.RegisteredClaims{
		ID:        time.Now().String(),
		Issuer:    "Fl0rencess720",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, rc).SignedString([]byte(refreshSecret))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func GenToken(userID int) (string, string, error) {
	accessToken, err := GenAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := GenRefreshToken()
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func ParseToken(aToken string) (*AuthClaims, bool, error) {
	accessSecret := viper.GetString("JWT_ACCESS_SECRET")

	accessToken, err := jwt.ParseWithClaims(aToken, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(accessSecret), nil
	})
	if err != nil {
		return nil, false, err
	}

	if claims, ok := accessToken.Claims.(*AuthClaims); ok && accessToken.Valid {
		return claims, false, nil
	}

	return nil, true, errors.New("invalid token")
}

func RefreshToken(aToken, rToken string) (string, error) {
	accessSecret := viper.GetString("JWT_ACCESS_SECRET")
	refreshSecret := viper.GetString("JWT_REFRESH_SECRET")
	var claims AuthClaims

	rToken = strings.TrimPrefix(rToken, "Bearer ")

	_, err := jwt.Parse(rToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(refreshSecret), nil
	})
	if err != nil {
		return "", err
	}

	_, err = jwt.ParseWithClaims(aToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(accessSecret), nil
	})
	v, _ := err.(*jwt.ValidationError)
	if v == nil || v.Errors == jwt.ValidationErrorExpired {
		return GenAccessToken(claims.UserID)
	}

	return "", err
}
