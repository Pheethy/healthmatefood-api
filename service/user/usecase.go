package user

import (
	"context"
	"healthmatefood-api/models"
	"mime/multipart"
	"sync"

	"github.com/gofrs/uuid"
)

type IUserUsecase interface {
	FetchUserPassport(ctx context.Context, req *models.User) (*models.UserPassport, error)
	FetchAllUsers(ctx context.Context, args *sync.Map) ([]*models.User, error)
	FetchOneUserById(ctx context.Context, id *uuid.UUID) (*models.User, error)
	UpsertUser(ctx context.Context, user *models.User, isAdmin bool, files []*multipart.FileHeader) error
	UpsertUserInfo(ctx context.Context, userInfo *models.UserInfo) error
	RefreshUserPassport(ctx context.Context, refreshToken string) (*models.UserPassport, error)
}
