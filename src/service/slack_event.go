package service

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/slack-go/slack"
)

type Users struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Probability struct {
	UserId      int     `json:"userId"`
	UserName    string  `json:"userName"`
	Probability float64 `json:"probability"`
}

func SlackCallbackEvent() {
}

func SlackAppMentionEvent() {
}

func GetUsers() ([]*slack.OptionBlockObject, error) {
	url := ""
	req, _ := http.NewRequest("GET", url, nil)
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
	var users []Users
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, err
	}
	var obo []*slack.OptionBlockObject
	for _, user := range users {
		obo = append(obo, &slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: user.Name}, Value: strconv.FormatInt(user.ID, 5)})
	}
	return obo, nil
}

func GetProbability() (Probability, string, error) {
	var probability Probability
	time := time.Now()
	time_str := time.Format("15:04:05")
	date := time.Format("2006-01-02")
	url := "https://staywatch-backend.kajilab.net/api/v1/probability/reporting/before?user_id=1&date=" + date + "&time=" + time_str
	req, _ := http.NewRequest("GET", url, nil)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return probability,"", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return probability, "", err
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &probability); err != nil {
		return probability, "", err
	}
	return probability, time_str, nil
}
