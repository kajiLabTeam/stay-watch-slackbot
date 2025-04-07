package model

import (
	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name        string
	SlackID     string
	StayWatchID int64
	Corresponds []Correspond `gorm:"foreignKey:UserID"`
}

type Tag struct {
	gorm.Model
	Name        string
	MinNumber   int
	Corresponds []Correspond `gorm:"foreignKey:TagID"`
}

type Correspond struct {
	gorm.Model
	TagID  uint
	UserID uint
}

type UserDetail struct {
	User             User
	VisitProbability float64
	VisitTime        string
	DeartureTime     string
}

var db *gorm.DB

func init() {
	db = lib.SqlConnect()
	db.AutoMigrate(&User{}, &Tag{}, &Correspond{})
}
