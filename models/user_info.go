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
	Id           *uuid.UUID        `json:"id" db:"id" type:"int" example:"1"`
	UserId       *uuid.UUID        `json:"user_id" db:"user_id" type:"int" example:"1"`
	Age          int64             `json:"age" db:"age" type:"int" example:"25"`
	Gender       string            `json:"gender" db:"gender" type:"string" example:"male"`
	Height       float64           `json:"height" db:"height" type:"float" example:"1.8"`
	Weight       float64           `json:"weight" db:"weight" type:"float" example:"80"`
	TargetWeight float64           `json:"target_weight" db:"target_weight" type:"float" example:"80"`
	ActiveLevel  string            `json:"active_level" db:"active_level" type:"string" example:"active"`
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
