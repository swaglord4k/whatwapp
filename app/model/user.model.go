package model

import "gorm.io/gorm"

const USER_MODEL = "users"

type User struct {
	gorm.Model
	Email    *string
	Password *string
	Role     Role
}
