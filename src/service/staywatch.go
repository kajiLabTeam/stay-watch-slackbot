// Package service provides business logic and external API integration for the Slack bot.
package service

import (
	"net/url"
	"strconv"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

// buildStayWatchURL はStayWatch APIのURLを構築する共通ヘルパー関数（複数ユーザー用）
func buildStayWatchURL(endpoint string, userIDs []int64, weekday int, timeStr string) string {
	u, _ := url.Parse(staywatch.BaseURL + endpoint)
	q := u.Query()
	for _, userID := range userIDs {
		q.Add("user-id", strconv.Itoa(int(userID)))
	}
	if weekday >= 0 {
		q.Add("weekday", strconv.Itoa(weekday))
	}
	if timeStr != "" {
		q.Add("time", timeStr)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// buildSingleUserURL はStayWatch APIのURLを構築する共通ヘルパー関数（単一ユーザー用）
func buildSingleUserURL(endpoint string, userID int, weekday int, timeStr string) string {
	u, _ := url.Parse(staywatch.BaseURL + endpoint)
	q := u.Query()
	q.Add("user-id", strconv.Itoa(userID))
	if weekday >= 0 {
		q.Add("weekday", strconv.Itoa(weekday))
	}
	if timeStr != "" {
		q.Add("time", timeStr)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// GetStayWatchMember StayWatchからメンバー一覧を取得する
func GetStayWatchMember() ([]StaywatchUsers, error) {
	var users []StaywatchUsers
	if err := stayWatchClient.Get(staywatch.BaseURL+staywatch.Users, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// GetStayWatchProbability 指定されたユーザーの来訪確率を取得する
func GetStayWatchProbability(users []model.User, weekday time.Weekday) []Probability {
	var userIDs []int64
	for _, user := range users {
		userIDs = append(userIDs, user.StayWatchID)
	}

	url := buildStayWatchURL(staywatch.Probability+"/visit", userIDs, int(weekday), "23:59")

	var r StayWatchResponse
	if err := stayWatchClient.Get(url, &r); err != nil {
		return []Probability{}
	}

	var probability []Probability
	for _, result := range r.Result {
		probability = append(probability, Probability{
			UserId:      int(result.UserID),
			Probability: result.Probability,
		})
	}
	return probability
}

// filterByThreshold 確率が閾値を超えるユーザーをフィルタリングする
func filterByThreshold(pro []Probability, threshold float64) []model.User {
	// StayWatchID を収集
	var stayWatchIDs []int64
	for _, u := range pro {
		if u.Probability >= threshold {
			stayWatchIDs = append(stayWatchIDs, int64(u.UserId))
		}
	}

	if len(stayWatchIDs) == 0 {
		return []model.User{}
	}

	// バッチで取得（N+1 クエリ問題を回避）
	var user model.User
	users, err := user.ReadByStayWatchIDs(stayWatchIDs)
	if err != nil {
		return []model.User{}
	}
	return users
}

// fetchPredictionTime 予測時刻を取得する（visit または departure）
func fetchPredictionTime(users []model.User, weekday time.Weekday, action string) []Result {
	var userIDs []int64
	for _, user := range users {
		userIDs = append(userIDs, user.StayWatchID)
	}

	url := buildStayWatchURL(staywatch.Time+"/"+action, userIDs, int(weekday), "")

	var r StayWatchResponse
	if err := stayWatchClient.Get(url, &r); err != nil {
		return []Result{}
	}
	return r.Result
}

// mergePredictions visit と departure の結果をマージする
func mergePredictions(vr []Result, dr []Result) []Prediction {
	var predictions []Prediction
	for _, v := range vr {
		for _, d := range dr {
			if v.UserID == d.UserID {
				predictions = append(predictions, Prediction{
					UserID:    v.UserID,
					Visit:     v.PredictionTime,
					Departure: d.PredictionTime,
				})
				break
			}
		}
	}
	return predictions
}
