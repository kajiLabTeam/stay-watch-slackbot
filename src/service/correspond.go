package service

import (
	"errors"

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
	corresponds, err := correspond.ReadByUserID()
	if err != nil {
		return correspond, err
	}
	for _, c := range corresponds {
		if c.EventID == event.ID {
			err := errors.New("correspond already exists")
			return correspond, err
		}
	}
	if err := correspond.Create(); err != nil {
		return correspond, err
	}

	return correspond, nil
}
