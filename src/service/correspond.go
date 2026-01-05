package service

import (
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterCorrespond(eventName string, slackUserID string) (model.Correspond, error) {
	event := model.Event{
		Name: eventName,
	}
	if err := event.ReadByName(); err != nil {
		return model.Correspond{}, err
	}

	user := model.User{
		SlackID: slackUserID,
	}
	if err := user.ReadBySlackID(); err != nil {
		return model.Correspond{}, err
	}

	// DBに対応情報を登録
	correspond := model.Correspond{
		UserID:  user.ID,
		EventID: event.ID,
	}
	// DB の UNIQUE 制約により重複登録は自動的にエラーとなる
	if err := correspond.Create(); err != nil {
		return correspond, err
	}

	return correspond, nil
}
