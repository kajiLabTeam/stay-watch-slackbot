package service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/slack-go/slack"
)

func SlackCallbackEvent() {
}

func SlackAppMentionEvent() {
}

func GetUsers() ([]*slack.OptionBlockObject, error) {
	req, err := http.NewRequest("GET", staywatch.BaseURL+staywatch.Users, nil)
	if err != nil {
		return nil, err
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, err
	}
	var obo []*slack.OptionBlockObject
	for _, user := range users {
		obo = append(obo, &slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: user.Name}, Value: strconv.FormatInt(user.ID, 10)})
	}
	return obo, nil
}

func GetProbability(userID int) (Probability, string, error) {
	var probability Probability
	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	time_str := now.Format("15:04")
	w := now.Weekday()
	// 月曜を0とする変換
	weekday := (int(w) + 6) % 7
	u, _ := url.Parse(staywatch.BaseURL + staywatch.Probability + "/visit")
	q := u.Query()
	q.Add("user-id", strconv.Itoa(userID))
	q.Add("weekday", strconv.Itoa(weekday))
	q.Add("time", time_str)
	u.RawQuery = q.Encode()
	res, err := http.Get(u.String())
	if err != nil {
		return probability, "", err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	var r StayWatchResponse
	if err := json.Unmarshal(b, &r); err != nil {
		return probability, "", err
	}
	probability.UserId = userID
	for _, user := range users {
		if user.ID == int64(userID) {
			probability.UserName = user.Name
			break
		}
	}
	probability.Probability = r.Result[0].Probability
	return probability, time_str, nil
}
