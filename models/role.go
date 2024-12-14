package models

type Roles struct {
	TableName string `db:"roles" json:"-" pk:"Id"`
	Id        int64  `db:"id" json:"id" type:"int64"`
	Name      string `db:"name" json:"name" type:"string"`
}
