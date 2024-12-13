package models

import (
	"time"

	"github.com/Pheethy/psql/helper"
	"github.com/gofrs/uuid"
)

type Image struct {
	TableName struct{}          `json:"-" db:"images" pk:"Id"`
	Id        *uuid.UUID        `json:"id" db:"id" type:"uuid"`
	FileName  string            `json:"filename" db:"filename" type:"string"`
	URL       string            `json:"url" db:"url" type:"string"`
	RefId     *uuid.UUID        `json:"ref_id" db:"ref_id" type:"uuid"`
	RefType   string            `json:"ref_type" db:"ref_type" type:"string"`
	CreatedAt *helper.Timestamp `json:"created_at" db:"created_at" type:"timestamp"`
	UpdatedAt *helper.Timestamp `json:"updated_at" db:"updated_at" type:"timestamp"`
}

func (p *Image) NewUUID() {
	id, _ := uuid.NewV4()
	p.Id = &id
}

func (p *Image) SetCreatedAt() {
	time := helper.NewTimestampFromTime(time.Now())
	p.CreatedAt = &time
}

func (p *Image) SetUpdatedAt() {
	time := helper.NewTimestampFromTime(time.Now())
	p.UpdatedAt = &time
}
