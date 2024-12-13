package handler

import (
	"errors"
	"fmt"
	"healthmatefood-api/config"
	"healthmatefood-api/constants"
	"healthmatefood-api/models"
	"healthmatefood-api/service/auth"
	"math"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type authHandler struct {
	cfg       config.IJwtConfig
	mapClaims *MapClaims
}

type MapClaims struct {
	Payload *models.UserClaims
	jwt.RegisteredClaims
}

func NewAuthHandler(tokenType constants.TokenType, cfg config.IJwtConfig, payload *models.UserClaims) (auth.IAuthHandler, error) {
	switch tokenType {
	case constants.TokenTypeAccess:
		return newAccessToken(cfg, payload), nil
	case constants.TokenTypeRefresh:
		return newRefreshToken(cfg, payload), nil
	default:
		return nil, errors.New("invalid token type")
	}
}

func newAccessToken(cfg config.IJwtConfig, payload *models.UserClaims) auth.IAuthHandler {
	return &authHandler{
		cfg: cfg,
		mapClaims: &MapClaims{
			Payload: payload,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "odor-api-auth",
				Subject:   "access-token",
				Audience:  []string{"customer", "admin"},
				ExpiresAt: jwtTimeDuration(cfg.AccessExpiresAt()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		},
	}
}

func newRefreshToken(cfg config.IJwtConfig, payload *models.UserClaims) auth.IAuthHandler {
	return &authHandler{
		cfg: cfg,
		mapClaims: &MapClaims{
			Payload: payload,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "odor-api-auth",
				Subject:   "refresh-token",
				Audience:  []string{"customer", "admin"},
				ExpiresAt: jwtTimeDuration(cfg.RefreshExpiresAt()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		},
	}
}

func (a *authHandler) SignToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, a.mapClaims)
	tokenStr, _ := token.SignedString([]byte(a.cfg.SecretKey()))
	return tokenStr
}

func ParseToken(cfg config.IJwtConfig, tokenStr string) (auth.IAuthHandler, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return cfg.SecretKey(), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, errors.New("token is malformed")
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token is expired")
		} else {
			return nil, errors.New(fmt.Sprintf("parse token error:%s", err.Error()))
		}
	}

	/* check payload maching mapClaims */
	if claims, ok := token.Claims.(*MapClaims); ok && token.Valid {
		return &authHandler{
			cfg:       cfg,
			mapClaims: claims,
		}, nil
	} else {
		return nil, errors.New("invalid token")
	}
}

func RepeatToken(cfg config.IJwtConfig, payload *models.UserClaims, exp int) string {
	token := &authHandler{
		cfg: cfg,
		mapClaims: &MapClaims{
			Payload: payload,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "odor-api-auth",
				Subject:   "refresh-token",
				Audience:  []string{"customer", "admin"},
				ExpiresAt: jwtTimeRepeatAdapter(exp),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		},
	}
	return token.SignToken()
}

func (a *authHandler) GetExpiresAt() int {
	return int(a.mapClaims.ExpiresAt.Unix())
}

func jwtTimeDuration(second int) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(time.Duration(int64(second) * int64(math.Pow10(9)))))
}

func jwtTimeRepeatAdapter(second int) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Unix(int64(second), 0))
}
