package repository

import (
	"context"
	"errors"
	"fmt"
	"healthmatefood-api/config"
	"healthmatefood-api/constants"
	"healthmatefood-api/models"
	"healthmatefood-api/service/auth"
	"math"
	"time"

	"github.com/Pheethy/psql/orm"
	"github.com/Pheethy/sqlx"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

type authRepository struct {
	cfg    config.IJwtConfig
	psqlDB *sqlx.DB
}

func NewAuthRepository(cfg config.IJwtConfig, psqlDB *sqlx.DB) auth.IAuthRepository {
	return &authRepository{
		psqlDB: psqlDB,
		cfg:    cfg,
	}
}

func (a *authRepository) NewAccessToken(payload *models.UserClaims) string {
	mapClaims := &models.MapClaims{
		Payload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "healthmatefood-api",
			Subject:   "access-token",
			Audience:  []string{"customer", "admin"},
			ExpiresAt: jwtTimeDuration(a.cfg.AccessExpiresAt()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return a.SignToken(mapClaims)

}

func (a *authRepository) NewRefreshToken(payload *models.UserClaims) string {
	mapClaims := &models.MapClaims{
		Payload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "healthmatefood-api",
			Subject:   "refresh-token",
			Audience:  []string{"customer", "admin"},
			ExpiresAt: jwtTimeDuration(a.cfg.RefreshExpiresAt()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return a.SignToken(mapClaims)
}

func (a *authRepository) SignToken(mapClaims *models.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	tokenStr, _ := token.SignedString([]byte(a.cfg.SecretKey()))
	return tokenStr
}

func (a *authRepository) ParseToken(tokenStr string) (*models.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &models.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.cfg.SecretKey(), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, errors.New("token is malformed")
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token is expired")
		} else {
			return nil, fmt.Errorf("parse token error:%s", err.Error())
		}
	}

	/* check payload maching mapClaims */
	if claims, ok := token.Claims.(*models.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("invalid token")
	}
}

func (a *authRepository) NewAccessTokenWithExpiresAt(payload *models.UserClaims, exp int) string {
	mapClaims := &models.MapClaims{
		Payload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "odor-api-auth",
			Subject:   "refresh-token",
			Audience:  []string{"customer", "admin"},
			ExpiresAt: jwtTimeRepeatAdapter(exp),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return a.SignToken(mapClaims)
}

func (a *authRepository) GetExpiresAt(mapClaims *models.MapClaims) int {
	return int(mapClaims.ExpiresAt.Unix())
}

func jwtTimeDuration(second int) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(time.Duration(int64(second) * int64(math.Pow10(9)))))
}

func jwtTimeRepeatAdapter(second int) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Unix(int64(second), 0))
}

func (m *authRepository) FindAccessToken(ctx context.Context, userId *uuid.UUID, accessToken string) bool {
	var ok bool
	sql := `
		SELECT
			(CASE WHEN count(*) = 1 THEN TRUE ELSE FALSE END)
		FROM
			oauth
		WHERE
			oauth.user_id = $1::uuid
		AND
			oauth.access_token = $2::text
	`
	stmt, err := m.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return false
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &ok, userId, accessToken); err != nil {
		return false
	}

	return ok
}

func (m *authRepository) FetchRoles(ctx context.Context) ([]*models.Roles, error) {
	sql := fmt.Sprintf(`
		SELECT
			%s
		FROM
			roles
		ORDER BY
      roles.id DESC;
	`,
		orm.GetSelector(models.Roles{}),
	)

	stmt, err := m.psqlDB.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx)
	if err != nil {
		return nil, err
	}

	roles, err := m.ormRoles(ctx, rows)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (m *authRepository) ormRoles(ctx context.Context, rows *sqlx.Rows) ([]*models.Roles, error) {
	mapper, err := orm.OrmContext(ctx, new(models.Roles), rows, orm.NewMapperOption())
	if err != nil {
		return nil, err
	}
	roles := mapper.GetData().([]*models.Roles)
	if len(roles) == 0 {
		return nil, errors.New(constants.ERROR_ROLES_NOT_FOUND)
	}

	return roles, nil
}
