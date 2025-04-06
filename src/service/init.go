package service

import (
	"github.com/kajiLabTeam/stay-watch-slackbot/conf"
)

type StayWatch struct {
	BaseURL     string
	Users       string
	Probability string
	Time        string
}

type Users struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Probability struct {
	UserId      int     `json:"userId"`
	UserName    string  `json:"userName"`
	Probability float64 `json:"probability"`
}

type Result struct {
	UserID         int64   `json:"userId"`
	Probability    float64 `json:"probability"`
	PredictionTime string  `json:"predictionTime"`
}

type StayWatchResponse struct {
	Weekday   int      `json:"weekday"`
	Time      string   `json:"time"`
	IsForward bool     `json:"isForward"`
	Result    []Result `json:"result"`
}

var users []Users

var staywatch StayWatch

func init() {
	s := conf.GetStayWatchConfig()
	staywatch.BaseURL = s.GetString("staywatch.url")
	staywatch.Users = s.GetString("staywatch.users")
	staywatch.Probability = s.GetString("staywatch.probability")
	staywatch.Time = s.GetString("staywatch.time")
}
