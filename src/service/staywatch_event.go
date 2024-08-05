package service

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/conf"
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

var url string

func init() {
	s := conf.GetStayWatchConfig()
	url = s.GetString("stay-watch.url")
}

func GetUsers() (obo []*slack.OptionBlockObject, err error) {
	getUsersURL := url + "/users/2"
	req, _ := http.NewRequest("GET", getUsersURL, nil)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return
	}
	body, _ := io.ReadAll(resp.Body)
	var users []Users
	if err = json.Unmarshal(body, &users); err != nil {
		return
	}
	for _, user := range users {
		obo = append(obo, &slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: user.Name}, Value: strconv.FormatInt(user.ID, 5)})
	}
	return
}

func GetProbability() (probability Probability, time_str string, err error) {
	time := time.Now()
	time_str = time.Format("15:04:05")
	date := time.Format("2006-01-02")
	getProbabilityURL := url + "/probability/reporting/before?user_id=1&date=" + date + "&time=" + time_str
	req, _ := http.NewRequest("GET", getProbabilityURL, nil)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return
	}
	body, _ := io.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &probability); err != nil {
		return
	}
	return
}
