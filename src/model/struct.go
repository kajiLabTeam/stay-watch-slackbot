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

// Status represents the status of an activity (start, end, pose)
type Status struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null"` // start, end, pose
	Logs []Log  `gorm:"foreignKey:StatusID"`
}

// Types represents the type of an event
type Types struct {
	gorm.Model
	Name   string  `gorm:"uniqueIndex;not null"`
	Events []Event `gorm:"foreignKey:TypesID"`
}

// Tool represents tools used in events
type Tool struct {
	gorm.Model
	Name   string  `gorm:"uniqueIndex;not null"`
	Events []Event `gorm:"foreignKey:ToolID"`
}

// Event represents an activity event (formerly Tag)
type Event struct {
	gorm.Model
	Name            string          `gorm:"not null"` // スケジュール、人生ゲーム、入退室、勉強会、ミーティング、作業中
	TypesID         *uint           // FK to Types
	Types           *Types          `gorm:"foreignKey:TypesID"`
	ToolID          *uint           // FK to Tool
	Tool            *Tool           `gorm:"foreignKey:ToolID"`
	Correspondences []Correspondence `gorm:"foreignKey:EventID"`
}

// Correspondence represents the relationship between Event and Tag
type Correspondence struct {
	gorm.Model
	EventID uint
	Event   Event `gorm:"foreignKey:EventID"`
	TagID   uint
	Tag     Tag   `gorm:"foreignKey:TagID"`
	Logs    []Log `gorm:"foreignKey:CorrespondenceID"`
}

// Log represents activity logs
type Log struct {
	gorm.Model
	Action            string
	CorrespondenceID  uint
	Correspondence    Correspondence `gorm:"foreignKey:CorrespondenceID"`
	StatusID          uint
	Status            Status `gorm:"foreignKey:StatusID"`
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
	db.AutoMigrate(&User{}, &Tag{}, &Correspond{}, &Status{}, &Types{}, &Tool{}, &Event{}, &Correspondence{}, &Log{})
}
