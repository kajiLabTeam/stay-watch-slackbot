package service

import (
	"github.com/kajiLabTeam/stay-watch-slackbot/conf"
	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
)

// StayWatch はStayWatch APIの設定を保持する
type StayWatch struct {
	BaseURL     string
	Users       string
	Probability string
	Time        string
	APIKey      string
}

// StaywatchUsers はStayWatchシステムのユーザーを表す
type StaywatchUsers struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Probability はユーザーの来訪確率を表す
type Probability struct {
	UserID      int     `json:"userId"`
	UserName    string  `json:"userName"`
	Probability float64 `json:"probability"`
}

// Result はStayWatch APIからの予測結果を表す
type Result struct {
	UserID         int64   `json:"userId"`
	Probability    float64 `json:"probability"`
	PredictionTime string  `json:"predictionTime"`
}

// StayWatchResponse はStayWatch APIからのレスポンスを表す
type StayWatchResponse struct {
	Weekday   int      `json:"weekday"`
	Time      string   `json:"time"`
	IsForward bool     `json:"isForward"`
	Result    []Result `json:"result"`
}

// Prediction はユーザーの予測来訪・退室時刻を表す
type Prediction struct {
	UserID    int64
	Visit     string
	Departure string
}

var staywatch StayWatch
var stayWatchClient *lib.StayWatchClient

func init() {
	s := conf.GetStayWatchConfig()
	staywatch.BaseURL = s.GetString("staywatch.url")
	staywatch.Users = s.GetString("staywatch.users")
	staywatch.Probability = s.GetString("staywatch.probability")
	staywatch.Time = s.GetString("staywatch.time")
	staywatch.APIKey = s.GetString("staywatch.api_key")

	// StayWatch API クライアントを初期化
	stayWatchClient = lib.NewStayWatchClient(staywatch.APIKey)
}
