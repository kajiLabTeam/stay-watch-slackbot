// Package model provides database models and data access methods.
package model

import (
	"log"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"gorm.io/gorm"
)

// User はシステムのユーザーを表す
type User struct {
	gorm.Model
	Name        string
	SlackID     string
	StayWatchID int64
	IconURL     string
	EventUsers  []EventUser `gorm:"foreignKey:UserID"`
}

// Status は活動のステータス（start, end, pose）を表す
type Status struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);uniqueIndex;not null"` // start, end, pose
	Logs []Log  `gorm:"foreignKey:StatusID"`
}

// Event は活動イベントを表す
type Event struct {
	gorm.Model
	Name       string `gorm:"type:varchar(255);uniqueIndex;not null"` // スマブラ、人生ゲーム など
	Code       string `gorm:"type:varchar(255);uniqueIndex;not null"` // イベントを一意に定める識別子（例: 1, 2, 0437ac48be2a81）
	MinNumber  int    `gorm:"default:2"`                              // 最低必要人数
	EventUsers []EventUser `gorm:"foreignKey:EventID"`
}

// EventUser は Event と User の関係を表す中間テーブル
type EventUser struct {
	gorm.Model
	EventID uint  `gorm:"uniqueIndex:idx_event_users_event_user"`
	Event   Event `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID  uint  `gorm:"uniqueIndex:idx_event_users_event_user"`
	User    User  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Log は活動ログを表す
type Log struct {
	ID               uint `gorm:"primarykey"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
	EventTime        time.Time
	EventID          uint
	Event            Event `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	StatusID         uint
	Status           Status `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RoomUsers        []User `gorm:"many2many:logs_user_rooms;"`
	ParticipateUsers []User `gorm:"many2many:logs_user_participates;"`
}

// LogsUserRoom は Log と User の中間テーブル（在室ユーザー）
type LogsUserRoom struct {
	gorm.Model
	LogID  uint
	Log    Log `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID uint
	User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// LogsUserParticipate は Log と User の中間テーブル（参加ユーザー）
type LogsUserParticipate struct {
	gorm.Model
	LogID  uint
	Log    Log `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID uint
	User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
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
	if err := db.AutoMigrate(&User{}, &Status{}, &Event{}, &EventUser{}, &Log{}, &LogsUserRoom{}, &LogsUserParticipate{}); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
}
