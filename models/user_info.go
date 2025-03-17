package models

import (
	"reflect"
	"time"

	"github.com/Pheethy/psql/helper"
	"github.com/gofrs/uuid"
	"github.com/spf13/cast"
)

type ActiveLevel string

const (
	sedentary  ActiveLevel = "SEDENTARY"
	light      ActiveLevel = "LIGHT"
	moderate   ActiveLevel = "MODERATE"
	active     ActiveLevel = "ACTIVE"
	veryActive ActiveLevel = "VERY_ACTIVE"
)

type UserInfo struct {
	TableName         struct{}          `json:"-" db:"user_info" pk:"Id"`
	Id                *uuid.UUID        `json:"id" db:"id" type:"uuid"`
	UserId            *uuid.UUID        `json:"user_id" db:"user_id" type:"uuid" `
	Firstname         string            `json:"firstname" db:"firstname" type:"string"`
	Lastname          string            `json:"lastname" db:"lastname" type:"string"`
	Gender            string            `json:"gender" db:"gender" type:"string"`
	Height            float64           `json:"height" db:"height" type:"float64"`
	Weight            float64           `json:"weight" db:"weight" type:"float64"`
	Target            string            `json:"target" db:"target" type:"string"`
	TargetWeight      float64           `json:"target_weight" db:"target_weight" type:"float64"`
	ActiveLevel       ActiveLevel       `json:"active_level" db:"active_level" type:"string"`
	Age               float64           `json:"age" db:"age" type:"float64"`
	BMR               float64           `json:"bmr" db:"bmr" type:"float64"`
	CaloriesLimit     float64           `json:"calories_limit" db:"calories_limit" type:"float64"`
	MedicalCondition  string            `json:"medical_condition" db:"medical_condition" type:"string"`
	FoodOrIngredients []string          `json:"food_or_ingredients" db:"food_or_ingredients" type:"string"`
	DOB               *helper.Timestamp `json:"dob" db:"dob" type:"timestamp"`
	CreatedAt         *helper.Timestamp `json:"created_at" db:"created_at" type:"timestamp"`
	UpdatedAt         *helper.Timestamp `json:"updated_at" db:"updated_at" type:"timestamp"`
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
		case "firstname":
			ptr.Firstname = cast.ToString(val)
		case "lastname":
			ptr.Lastname = cast.ToString(val)
		case "gender":
			ptr.Gender = cast.ToString(val)
		case "height":
			ptr.Height = cast.ToFloat64(val)
		case "weight":
			ptr.Weight = cast.ToFloat64(val)
		case "target":
			ptr.Target = cast.ToString(val)
		case "target_weight":
			ptr.TargetWeight = cast.ToFloat64(val)
		case "active_level":
			ptr.ActiveLevel = ActiveLevel(cast.ToString(val))
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
		case "dob":
			if val != nil {
				if reflect.TypeOf(val).Kind() == reflect.String {
					timestamp := helper.NewTimestampFromString(val.(string))
					ptr.DOB = &timestamp
				} else if reflect.TypeOf(val).String() == "time.Time" {
					timestamp := helper.NewTimestampFromTime(val.(time.Time))
					ptr.DOB = &timestamp
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

func (u *UserInfo) GetAge() {
	currentTime := time.Now()
	// ใช้ Year() แทน YearDay() เพื่อดึงปี
	age := currentTime.Year() - u.DOB.ToTime().Year()

	// ตรวจสอบว่าถึงวันเกิดในปีนี้หรือยัง
	if currentTime.Month() < u.DOB.ToTime().Month() ||
		(currentTime.Month() == u.DOB.ToTime().Month() && currentTime.Day() < u.DOB.ToTime().Day()) {
		age--
	}

	u.Age = float64(age)
}

func (u *UserInfo) GetCaloriesLimit() {
	switch u.ActiveLevel {
	case "SEDENTARY":
		u.CaloriesLimit = u.BMR * 1.2
	case "LIGHT":
		u.CaloriesLimit = u.BMR * 1.375
	case "MODERATE":
		u.CaloriesLimit = u.BMR * 1.55
	case "ACTIVE":
		u.CaloriesLimit = u.BMR * 1.725
	case "VERY_ACTIVE":
		u.CaloriesLimit = u.BMR * 1.9
	default:
		u.CaloriesLimit = u.BMR * 1.2
	}
}

func (u *UserInfo) GetBMR() {
	u.GetAge()
	if u.Gender == "MALE" {
		switch {
		case u.Age < 3:
			u.BMR = 59.512*u.Weight - 30.4
		case u.Age < 10:
			u.BMR = 22.706*u.Weight + 504.3
		case u.Age < 18:
			u.BMR = 17.686*u.Weight + 658.2
		case u.Age < 30:
			u.BMR = 15.057*u.Weight + 692.2
		case u.Age < 60:
			u.BMR = 11.472*u.Weight + 873.1
		default:
			u.BMR = 11.711*u.Weight + 587.7
		}
		return
	}

	switch {
	case u.Age < 3:
		u.BMR = 58.317*u.Weight - 31.1
	case u.Age < 10:
		u.BMR = 20.315*u.Weight + 485.9
	case u.Age < 18:
		u.BMR = 13.384*u.Weight + 692.6
	case u.Age < 30:
		u.BMR = 14.818*u.Weight + 486.6
	case u.Age < 60:
		u.BMR = 8.126*u.Weight + 845.6
	default:
		u.BMR = 9.082*u.Weight + 658.5
	}
}
