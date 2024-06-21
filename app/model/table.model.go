package model

import (
	"gorm.io/gorm"
)

const TABLE_MODEL = "tables"

type Table struct {
	gorm.Model
	Name        int  `json:"name" db:"name" gorm:"primaryKey"`
	MaxWaitTime *int `json:"maxWaitTime" db:"max_wait_time"`
	MinPlayers  *int `json:"min" db:"min"`
	MaxPlayers  *int `json:"max" db:"max"`
}
