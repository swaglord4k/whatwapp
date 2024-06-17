package model

import "gorm.io/gorm"

const MATCH_MODEL = "matches"

type Match struct {
	gorm.Model
	PlayerId  uint  `json:"playerId" db:"player_id"`
	TableName int   `json:"tableName" db:"table"`
	League    *int  `json:"league" db:"league"`
	Server    *uint `json:"server" db:"server"`
}
