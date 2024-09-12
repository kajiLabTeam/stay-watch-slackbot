package service

import (
	"strconv"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
	"github.com/slack-go/slack"
)

func GetAllTagsObject() (optionObject []*slack.OptionBlockObject, err error) {
	tags, err := model.ReadAllTags()
	if err != nil {
		return nil, err
	}
	for _, tag := range tags {
		optionObject = append(optionObject, &slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: tag.TagName}, Value: strconv.FormatInt(int64(tag.ID), 5)})
	}
	return
}
