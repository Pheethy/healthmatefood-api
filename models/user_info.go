package models

import (
	"reflect"
	"time"

	"github.com/Pheethy/psql/helper"
	"github.com/gofrs/uuid"
	"github.com/spf13/cast"
)

type UserInfo struct {
	TableName    struct{}          `json:"-" db:"user_info" pk:"Id"`
	Id           *uuid.UUID        `json:"id" db:"id" type:"uuid"`
	UserId       *uuid.UUID        `json:"user_id" db:"user_id" type:"uuid" `
	Age          int64             `json:"age" db:"age" type:"int64"`
	Gender       string            `json:"gender" db:"gender" type:"string"`
	Height       float64           `json:"height" db:"height" type:"float64"`
	Weight       float64           `json:"weight" db:"weight" type:"float64"`
	TargetWeight float64           `json:"target_weight" db:"target_weight" type:"float64"`
	ActiveLevel  string            `json:"active_level" db:"active_level" type:"string"`
	CreatedAt    *helper.Timestamp `json:"created_at" db:"created_at" type:"timestamp"`
	UpdatedAt    *helper.Timestamp `json:"updated_at" db:"updated_at" type:"timestamp"`
}

func NewUserInfoWithParams(params map[string]interface{}, ptr *UserInfo) *UserInfo {
	if ptr == nil {
		ptr = new(UserInfo)
	}
	for key, val := range params {
		switch key {
		case "id":
			id := uuid.FromStringOrNil(val.(string))
			ptr.Id = &id
		case "user_id":
			userId := uuid.FromStringOrNil(val.(string))
			ptr.UserId = &userId
		case "age":
			ptr.Age = cast.ToInt64(val)
		case "gender":
			ptr.Gender = cast.ToString(val)
		case "height":
			ptr.Height = cast.ToFloat64(val)
		case "weight":
			ptr.Weight = cast.ToFloat64(val)
		case "target_weight":
			ptr.TargetWeight = cast.ToFloat64(val)
		case "active_level":
			ptr.ActiveLevel = cast.ToString(val)
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

func (u *UserInfo) NewID() {
	id, _ := uuid.NewV4()
	u.Id = &id
}

func (u *UserInfo) SetCreatedAt() {
	time := helper.NewTimestampFromTime(time.Now())
	u.CreatedAt = &time
}

func (u *UserInfo) SetUpdatedAt() {
	time := helper.NewTimestampFromTime(time.Now())
	u.UpdatedAt = &time
}
