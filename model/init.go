package model

import (
	"gorm.io/gorm"
)

var (
	Db            *gorm.DB
	SessionLength int
)
