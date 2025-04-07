package service

import (
	"errors"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterUser(slackUseID string, userName string) (model.User, error) {
	// userNameをもとに滞在ウォッチからユーザ情報を取得
	users, err := GetStayWatchMember()
	if err != nil {
		return model.User{}, err
	}
	user := model.User{
		Name:    userName,
		SlackID: slackUseID,
		StayWatchID: int64(0),
	}
	if err := user.ReadByName(); err != nil {
		return user, err
	}
	if user.StayWatchID != 0 {
		err := errors.New("user already exists")
		return user, err
	}
	for i, u := range users {
		if u.Name == userName {
			user.StayWatchID = u.ID
			break
		}
		if i == len(users)-1 {
			err := errors.New("user not found")
			return user, err
		}
	}
	// DBにユーザ情報を登録
	if err := user.Create(); err != nil {
		return user, err
	}
	return user, nil
}
