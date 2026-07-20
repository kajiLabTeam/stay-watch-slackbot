package service

import (
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterEventUser(eventName string, slackUserID string) (model.EventUser, error) {
	event := model.Event{
		Name: eventName,
	}
	if err := event.ReadByName(); err != nil {
		return model.EventUser{}, err
	}

	user := model.User{
		SlackID: slackUserID,
	}
	if err := user.ReadBySlackID(); err != nil {
		return model.EventUser{}, err
	}

	// DBに対応情報を登録
	eventUser := model.EventUser{
		UserID:  user.ID,
		EventID: event.ID,
	}
	// DB の UNIQUE 制約により重複登録は自動的にエラーとなる
	if err := eventUser.Create(); err != nil {
		return eventUser, err
	}

	return eventUser, nil
}
