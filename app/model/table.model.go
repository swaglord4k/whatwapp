package model

import "gorm.io/gorm"

const TABLE_MODEL = "tables"

type Table struct {
	gorm.Model
	Name int  `json:"name" db:"name" gorm:"primaryKey"`
	Min  *int `json:"min" db:"min"`
	Max  *int `json:"max" db:"max"`
}
