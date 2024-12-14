package usecase

import (
	"context"
	"errors"
	"fmt"
	"healthmatefood-api/config"
	"healthmatefood-api/constants"
	"healthmatefood-api/models"
	auth_handler "healthmatefood-api/service/auth/handler"
	"healthmatefood-api/service/file"
	"healthmatefood-api/service/user"
	"healthmatefood-api/utils"
	"math"
	"mime/multipart"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	cfg      config.Iconfig
	userRepo user.IUserRepository
	fileUs   file.IFileUsecase
}

func NewUserUsecase(cfg config.Iconfig, userRepo user.IUserRepository, fileUs file.IFileUsecase) user.IUserUsecase {
	return &userUsecase{
		cfg:      cfg,
		userRepo: userRepo,
		fileUs:   fileUs,
	}
}

func (u *userUsecase) FetchUserPassport(ctx context.Context, req *models.User) (*models.UserPassport, error) {
	passport := new(models.UserPassport)

	/* Find User By Email */
	user, err := u.userRepo.FetchOneUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	/* Compare password */
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New(constants.ERROR_PASSWORD_IS_INVALID)
	}

	/* New Auth With Access Token */
	authAccess, err := auth_handler.NewAuthHandler(constants.TokenTypeAccess, u.cfg.Jwt(), user.GetUserClaims())
	if err != nil {
		return nil, err
	}
	/* New Auth With Refresh Token */
	authRefresh, err := auth_handler.NewAuthHandler(constants.TokenTypeRefresh, u.cfg.Jwt(), user.GetUserClaims())
	if err != nil {
		return nil, err
	}

	/* Insert OAuth */
	oauth := new(models.OAuth)
	oauth.SetData(user.Id, authAccess.SignToken(), authRefresh.SignToken())
	oauth.SetCreatedAt()
	oauth.SetUpdatedAt()
	if err := u.userRepo.UpsertOAuth(ctx, oauth); err != nil {
		return nil, err
	}
	/* Set Passport */
	passport.User = user
	passport.Token = &models.Token{
		OAuthId:      oauth.Id,
		AccessToken:  oauth.AccessToken,
		RefreshToken: oauth.RefreshToken,
	}
	return passport, nil
}

func (u *userUsecase) FetchAllUsers(ctx context.Context, args *sync.Map) ([]*models.User, error) {
	return u.userRepo.FetchAllUsers(ctx, args)
}

func (u *userUsecase) FetchOneUserById(ctx context.Context, id *uuid.UUID) (*models.User, error) {
	return u.userRepo.FetchOneUserById(ctx, id)
}

func (u *userUsecase) FetchOneUserInfoByUserId(ctx context.Context, userId *uuid.UUID) (*models.UserInfo, error) {
	return u.userRepo.FetchOneUserInfoByUserId(ctx, userId)
}

func (u *userUsecase) UpsertUser(ctx context.Context, user *models.User, isAdmin bool, files []*multipart.FileHeader) error {
	if len(files) > 0 {
		if err := u.prepareImage(ctx, user, files); err != nil {
			return err
		}
	}
	switch isAdmin {
	case true:
		user.RoleId = constants.USER_ROLE_ADMIN
	case false:
		user.RoleId = constants.USER_ROLE_CUSTOMER
	}
	if err := u.userRepo.UpsertUser(ctx, user); err != nil {
		return err
	}
	if err := u.userRepo.UpsertImages(ctx, user); err != nil {
		return err
	}
	return nil
}

func (u *userUsecase) UpsertUserInfo(ctx context.Context, userInfo *models.UserInfo) error {
	return u.userRepo.UpsertUserInfo(ctx, userInfo)
}

func (u *userUsecase) RefreshUserPassport(ctx context.Context, refreshToken string) (*models.UserPassport, error) {
	passport := new(models.UserPassport)
	token, err := auth_handler.ParseToken(u.cfg.Jwt(), refreshToken)
	if err != nil {
		return nil, err
	}
	oauth, err := u.userRepo.FetchOneOAuthByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	user, err := u.userRepo.FetchOneUserById(ctx, oauth.UserId)
	if err != nil {
		return nil, err
	}
	authAccess, err := auth_handler.NewAuthHandler(constants.TokenTypeAccess, u.cfg.Jwt(), user.GetUserClaims())
	if err != nil {
		return nil, err
	}
	newRefreshToken := auth_handler.RepeatToken(u.cfg.Jwt(), user.GetUserClaims(), token.GetExpiresAt())
	/* Update OAuth */
	oauth.AccessToken = authAccess.SignToken()
	oauth.RefreshToken = newRefreshToken
	oauth.SetUpdatedAt()
	if err := u.userRepo.UpsertOAuth(ctx, oauth); err != nil {
		return nil, err
	}
	/* Set Passport */
	passport.User = user
	passport.Token = &models.Token{
		OAuthId:      oauth.Id,
		AccessToken:  oauth.AccessToken,
		RefreshToken: oauth.RefreshToken,
	}
	return passport, nil
}

func (u *userUsecase) prepareImage(ctx context.Context, user *models.User, files []*multipart.FileHeader) error {
	if len(files) > 0 {
		reqFile := make([]*models.FileReq, 0)
		for _, file := range files {
			ext := strings.TrimPrefix(filepath.Ext(file.Filename), ".")
			if ok := u.validateFileType(ext); !ok {
				return errors.New("file type is invalid")
			}

			if file.Size > int64(u.cfg.App().FileLimit()) {
				return fmt.Errorf("file size must less than %d MiB", int(math.Ceil(float64(u.cfg.App().FileLimit())/math.Pow(1024, 2))))
			}

			filename := utils.RandFileName(ext)
			reqFile = append(reqFile, &models.FileReq{
				File:        file,
				Destination: constants.USER_IMAGE_DESTINETION + "/" + filename,
				Extension:   ext,
				FileName:    file.Filename,
			})
		}

		/* upload images to google cloud platfrom */
		filesResp, err := u.fileUs.UploadToGCP(ctx, reqFile)
		if err != nil {
			return fmt.Errorf("upload product image failed: %v", err.Error())
		}
		user.Images = models.FilesResp(filesResp).GetImagesFromFilesResp(user)
	}

	return nil
}

func (p *userUsecase) validateFileType(ext string) bool {
	if ext == "" {
		return false
	}

	expMap := []string{"png", "jpg", "jpeg"}
	for index := range expMap {
		if expMap[index] == ext {
			return true
		}
	}
	return false
}
