package models

import "go-cygnus/utils/db"

type BaseModel struct {
	ID        uint         `json:"id" gorm:"primary_key"`
	CreatedAt db.JSONTime  `json:"created_at"`
	UpdatedAt db.JSONTime  `json:"updated_at"`
	DeletedAt *db.JSONTime `sql:"index" json:"deleted_at"`
}
