package model

import "gorm.io/gorm"

const PLAYER_MODEL = "players"

type Player struct {
	gorm.Model
	Username string `json:"username" db:"username"`
	League   int    `json:"league" db:"league"`
}
