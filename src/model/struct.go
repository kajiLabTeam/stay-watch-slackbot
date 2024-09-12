package model

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	TagName       string            `json:"tagName"`
	Correspondences []Correspondence `gorm:"foreignkey:TagID" json:"correspondences"`
}

type User struct {
	gorm.Model
	Name            string            `json:"name"`
	SlackID         string            `json:"slackId"`
	StayWatchID     string            `json:"stayWatchId"`
	Correspondences []Correspondence `gorm:"foreignkey:UserID" json:"correspondences"`
}

type Correspondence struct {
	gorm.Model
	TagID uint `json:"tagId"`
	UserID  uint `json:"userId"`
}
