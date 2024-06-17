package model

import "gorm.io/gorm"

const SERVER_MODEL = "servers"

type Server struct {
	gorm.Model
}
