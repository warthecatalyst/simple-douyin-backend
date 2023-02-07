package dao

import (
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"gorm.io/gorm"
	"sync"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
)

func dataBaseInitialization() {
	dbOnce.Do(func() {
		db = initialization.GetDB()
	})
}
