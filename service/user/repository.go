package user

import (
	"context"
	"healthmatefood-api/models"
	"sync"

	"github.com/gofrs/uuid"
)

type IUserRepository interface {
	FetchAllUsers(ctx context.Context, args *sync.Map) ([]*models.User, error)
	FetchOneUserById(ctx context.Context, id *uuid.UUID) (*models.User, error)
	FetchOneUserByEmail(ctx context.Context, email string) (*models.User, error)
	FetchOneOAuthByRefreshToken(ctx context.Context, refreshToken string) (*models.OAuth, error)
	UpsertUser(ctx context.Context, user *models.User) error
	UpsertImages(ctx context.Context, user *models.User) error
	UpsertOAuth(ctx context.Context, oauth *models.OAuth) error
	UpsertUserInfo(ctx context.Context, userInfo *models.UserInfo) error
}
