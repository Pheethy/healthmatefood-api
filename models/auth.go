package models

import (
	"time"

	"github.com/Pheethy/psql/helper"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	OAuthId      *uuid.UUID `json:"oauth_id" db:"oauth_id" type:"uuid"`
	AccessToken  string     `json:"access_token" db:"access_token" type:"string"`
	RefreshToken string     `json:"refresh_token" db:"refresh_token" type:"string"`
}

type UserClaims struct {
	Id     *uuid.UUID `json:"id" db:"id" type:"uuid"`
	RoleId int64      `json:"role_id" db:"role_id" type:"int"`
}

type UserPassport struct {
	User  *User  `json:"user"`
	Token *Token `json:"token"`
}

type MapClaims struct {
	Payload *UserClaims
	jwt.RegisteredClaims
}

func (a *MapClaims) GetExpiresAt() int {
	return int(a.ExpiresAt.Unix())
}

type OAuth struct {
	TableName    struct{}          `json:"-" db:"oauth" pk:"Id"`
	Id           *uuid.UUID        `json:"id" db:"id" type:"uuid"`
	UserId       *uuid.UUID        `json:"user_id" db:"user_id" type:"uuid"`
	AccessToken  string            `json:"access_token" db:"access_token" type:"string"`
	RefreshToken string            `json:"refresh_token" db:"refresh_token" type:"string"`
	CreatedAt    *helper.Timestamp `json:"created_at" db:"created_at" type:"timestamp"`
	UpdatedAt    *helper.Timestamp `json:"updated_at" db:"updated_at" type:"timestamp"`
}

func (o *OAuth) NewId() {
	id := uuid.Must(uuid.NewV4())
	o.Id = &id
}

func (o *OAuth) SetData(userId *uuid.UUID, accessToken string, refreshToken string) {
	o.NewId()
	o.UserId = userId
	o.AccessToken = accessToken
	o.RefreshToken = refreshToken
}

func (o *OAuth) SetCreatedAt() {
	ti := helper.NewTimestampFromTime(time.Now())
	o.CreatedAt = &ti
}

func (o *OAuth) SetUpdatedAt() {
	ti := helper.NewTimestampFromTime(time.Now())
	o.UpdatedAt = &ti
}
