package service

import (
	"strconv"
	"time"

	"github.com/slack-go/slack"
)

func GetUsers() ([]*slack.OptionBlockObject, error) {
	users, err := GetStayWatchMember()
	if err != nil {
		return nil, err
	}
	var obo []*slack.OptionBlockObject
	for _, user := range users {
		obo = append(obo, &slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: user.Name}, Value: strconv.FormatInt(user.ID, 10)})
	}
	return obo, nil
}

func GetProbability(userID int) (Probability, string, error) {
	users, err := GetStayWatchMember()
	if err != nil {
		return Probability{}, "", err
	}

	var probability Probability
	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	timeStr := now.Format("15:04")
	w := now.Weekday()
	// 月曜を0とする変換
	weekday := (int(w) + 6) % 7

	url := buildSingleUserURL(staywatch.Probability+"/visit", userID, weekday, timeStr)

	var r StayWatchResponse
	if err := stayWatchClient.Get(url, &r); err != nil {
		return probability, "", err
	}

	probability.UserID = userID
	for _, user := range users {
		if user.ID == int64(userID) {
			probability.UserName = user.Name
			break
		}
	}
	probability.Probability = r.Result[0].Probability
	return probability, timeStr, nil
}
