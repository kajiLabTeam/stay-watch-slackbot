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

// Status represents the status of an activity (start, end, pose)
type Status struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);uniqueIndex;not null"` // start, end, pose
	Logs []Log  `gorm:"foreignKey:StatusID"`
}

// Types represents the type of an event
type Type struct {
	gorm.Model
	Name   string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Events []Event `gorm:"foreignKey:TypeID"`
}

// Tool represents tools used in events
type Tool struct {
	gorm.Model
	Name   string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Events []Event `gorm:"many2many:event_tools;"`
}

// Event represents an activity event
type Event struct {
	gorm.Model
	Name        string `gorm:"not null"` // スケジュール、人生ゲーム、入退室、勉強会、ミーティング、作業中
	MinNumber   int    `gorm:"default:2"` // 最低必要人数
	TypeID      uint
	Type        Type         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Tools       []Tool       `gorm:"many2many:event_tools;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Corresponds []Correspond `gorm:"foreignKey:EventID"`
}

// Correspond represents the relationship between Event and User
type Correspond struct {
	gorm.Model
	EventID uint
	Event   Event `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID  uint
	User    User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Log represents activity logs
type Log struct {
	gorm.Model
	EventID  uint
	Event    Event `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	StatusID uint
	Status   Status `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type UserDetail struct {
	User             User
	VisitProbability float64
	VisitTime        string
	DepartureTime    string
}

var db *gorm.DB

func init() {
	db = lib.SqlConnect()
	db.AutoMigrate(&User{}, &Status{}, &Type{}, &Tool{}, &Event{}, &Correspond{}, &Log{})
}
