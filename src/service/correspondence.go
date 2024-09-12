package service

import "github.com/kajiLabTeam/stay-watch-slackbot/model"

func SetCorrespondences(slackUserID string, Tags []string) error {
	user := model.User{SlackID: slackUserID}
	if err := user.Read(); err != nil {
		return err
	}
	for _, tag := range Tags {
		tagID := StringToUint(tag)
		err := model.Correspondence{TagID: tagID, UserID: user.ID}.Create()
		if err != nil {
			return err
		}
	}
	return nil
}
