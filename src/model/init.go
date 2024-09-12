package model

import (
	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	db = lib.SqlConnect()
	db.AutoMigrate(&User{}, &Tag{}, &Correspondence{})
}
