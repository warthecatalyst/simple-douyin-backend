package dao

import (
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"gorm.io/gorm"
)

var db *gorm.DB

func DataBaseInitialization() {
	db = initialization.GetDB()
}
