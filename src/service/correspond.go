package service

import (
	"errors"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterCorrespond(tagName string, slackUserID string) (model.Correspond, error) {
	tag := model.Tag{
		Name: tagName,
	}
	if err := tag.ReadByName(); err != nil {
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
		UserID: user.ID,
		TagID:  tag.ID,
	}
	corresponds, err := correspond.ReadByUserID()
	if err != nil {
		return correspond, err
	}
	for _, c := range corresponds {
		if c.TagID == tag.ID {
			err := errors.New("correspond already exists")
			return correspond, err
		}
	}
	if err := correspond.Create(); err != nil {
		return correspond, err
	}

	return correspond, nil
}
