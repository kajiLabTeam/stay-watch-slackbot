package service

import (
	"os"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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

var (
	staywatch       StayWatch
	stayWatchClient *lib.StayWatchClient
)

func init() {
	staywatch.BaseURL = getEnv("STAYWATCH_URL", "")
	staywatch.Users = getEnv("STAYWATCH_USERS_PATH", "")
	staywatch.Probability = getEnv("STAYWATCH_PROBABILITY_PATH", "")
	staywatch.Time = getEnv("STAYWATCH_TIME_PATH", "")
	staywatch.APIKey = getEnv("STAYWATCH_API_KEY", "")

	// StayWatch API クライアントを初期化
	stayWatchClient = lib.NewStayWatchClient(staywatch.APIKey)
}
