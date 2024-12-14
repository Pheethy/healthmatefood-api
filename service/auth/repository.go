package auth

import (
	"context"
	"healthmatefood-api/models"

	"github.com/gofrs/uuid"
)

type IAuthRepository interface {
	FindAccessToken(ctx context.Context, userId *uuid.UUID, accessToken string) bool
	FetchRoles(ctx context.Context) ([]*models.Roles, error)
	NewAccessToken(payload *models.UserClaims) string
	NewRefreshToken(payload *models.UserClaims) string
	NewAccessTokenWithExpiresAt(payload *models.UserClaims, exp int) string
	SignToken(mapClaims *models.MapClaims) string
	ParseToken(tokenStr string) (*models.MapClaims, error)
}
