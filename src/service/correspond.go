package service

import "github.com/kajiLabTeam/stay-watch-slackbot/model"

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
	if err := correspond.Create(); err != nil {
		return correspond, err
	}

	return correspond, nil
}
