package models

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/Pheethy/psql/helper"
	"github.com/gofrs/uuid"
	"github.com/spf13/cast"
	"golang.org/x/crypto/bcrypt"
)

// class
type User struct {
	TableName struct{}          `json:"-" db:"users" pk:"Id"`
	Id        *uuid.UUID        `json:"id" db:"id" type:"uuid" example:"U00001"`
	Username  string            `json:"username" db:"username" type:"string" example:"john_doe"`
	Password  string            `json:"-" db:"password" type:"string"`
	Email     string            `json:"email" db:"email" type:"string"`
	RoleId    int               `json:"-" db:"role_id" type:"int"`
	Role      string            `json:"role" db:"role" type:"string"`
	CreatedAt *helper.Timestamp `json:"created_at" db:"created_at" type:"timestamp"`
	UpdatedAt *helper.Timestamp `json:"updated_at" db:"updated_at" type:"timestamp"`

	Images []*Image `json:"images" db:"-" fk:"fk_field1:Id, fk_field2:RefId"`
}

func NewUserWithParams(params map[string]interface{}, ptr *User) *User {
	if ptr == nil {
		ptr = new(User)
	}
	for key, val := range params {
		switch key {
		case "id":
			id := uuid.FromStringOrNil(val.(string))
			ptr.Id = &id
		case "username":
			ptr.Username = cast.ToString(val)
		case "password":
			ptr.Password = cast.ToString(val)
		case "email":
			ptr.Email = cast.ToString(val)
		case "created_at":
			if val != nil {
				if reflect.TypeOf(val).Kind() == reflect.String {
					timestamp := helper.NewTimestampFromString(val.(string))
					ptr.CreatedAt = &timestamp
				} else if reflect.TypeOf(val).String() == "time.Time" {
					timestamp := helper.NewTimestampFromTime(val.(time.Time))
					ptr.CreatedAt = &timestamp
				}
			}
		case "updated_at":
			if val != nil {
				if reflect.TypeOf(val).Kind() == reflect.String {
					timestamp := helper.NewTimestampFromString(val.(string))
					ptr.UpdatedAt = &timestamp
				} else if reflect.TypeOf(val).String() == "time.Time" {
					timestamp := helper.NewTimestampFromTime(val.(time.Time))
					ptr.UpdatedAt = &timestamp
				}
			}
		}
	}

	return ptr
}

func (u *User) NewID() {
	id := uuid.Must(uuid.NewV4())
	u.Id = &id
}

func (u *User) SetCreatedAt() {
	ti := helper.NewTimestampFromTime(time.Now())
	u.CreatedAt = &ti
}

func (u *User) SetUpdatedAt() {
	ti := helper.NewTimestampFromTime(time.Now())
	u.UpdatedAt = &ti
}

func (u *User) BcryptHashing() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		return fmt.Errorf("hashing password failed: %s", err.Error())
	}
	u.Password = string(hash)
	return nil
}

func (u *User) ComparePassword(i *User) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(i.Password)); err != nil {
		return false
	}
	return true
}

func (u *User) IsEmail() bool {
	match, err := regexp.MatchString(`^[\w\-.]+@([\w\-]+\.)+[\w\-]{2,4}$`, u.Email)
	if err != nil {
		return false
	}
	return match
}

func (u *User) GetUserClaims() *UserClaims {
	return &UserClaims{
		Id:     u.Id,
		RoleId: int64(u.RoleId),
	}
}
