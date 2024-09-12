package service

import (
	"strconv"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
	"github.com/slack-go/slack"
)

func SetUser(slackUser slack.User) error {
	user := model.User{SlackID: slackUser.ID}
	if err := user.Read(); err != nil {
		return err
	}
	if user.ID == 0 {
		staywatchUser, err := GetStayWatchUser(slackUser.Name)
		if err != nil {
			return err
		}
		user = model.User{
			Name:        slackUser.Name,
			SlackID:     slackUser.ID,
			StayWatchID: strconv.FormatInt(staywatchUser.ID, 10),
		}
		if err := user.Create(); err != nil {
			return err
		}
	}

	return nil
}
