package model

import "gorm.io/gorm"

const GAME_MODEL = "games"

type Server struct {
	gorm.Model
}
