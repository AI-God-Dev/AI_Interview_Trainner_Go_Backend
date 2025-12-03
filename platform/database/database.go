package database

import (
	"gorm.io/gorm"
)

var (
	DBConn *gorm.DB // TODO: refactor to DI later
)
