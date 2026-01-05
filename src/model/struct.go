// Package model provides database models and data access methods.
package model

import (
	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"gorm.io/gorm"
)

// User はシステムのユーザーを表す
type User struct {
	gorm.Model
	Name        string
	SlackID     string
	StayWatchID int64
	Corresponds []Correspond `gorm:"foreignKey:UserID"`
}

// Status は活動のステータス（start, end, pose）を表す
type Status struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);uniqueIndex;not null"` // start, end, pose
	Logs []Log  `gorm:"foreignKey:StatusID"`
}

// Type はイベントのタイプを表す
type Type struct {
	gorm.Model
	Name   string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Events []Event `gorm:"foreignKey:TypeID"`
}

// Tool はイベントで使用されるツールを表す
type Tool struct {
	gorm.Model
	Name   string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Events []Event `gorm:"many2many:event_tools;"`
}

// Event は活動イベントを表す
type Event struct {
	gorm.Model
	Name        string `gorm:"type:varchar(255);uniqueIndex;not null"` // スケジュール、人生ゲーム、入退室、勉強会、ミーティング、作業中
	MinNumber   int    `gorm:"default:2"` // 最低必要人数
	TypeID      uint
	Type        Type         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Tools       []Tool       `gorm:"many2many:event_tools;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Corresponds []Correspond `gorm:"foreignKey:EventID"`
}

// Correspond はEventとUserの関係を表す
type Correspond struct {
	gorm.Model
	EventID uint
	Event   Event `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID  uint
	User    User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Log は活動ログを表す
type Log struct {
	gorm.Model
	EventID  uint
	Event    Event `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	StatusID uint
	Status   Status `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// UserDetail は来訪予測を含む詳細なユーザー情報を表す
type UserDetail struct {
	User             User
	VisitProbability float64
	VisitTime        string
	DepartureTime    string
}

var db *gorm.DB

func init() {
	db = lib.SQLConnect()
	db.AutoMigrate(&User{}, &Status{}, &Type{}, &Tool{}, &Event{}, &Correspond{}, &Log{})
}
