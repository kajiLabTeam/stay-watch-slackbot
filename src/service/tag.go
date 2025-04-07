package service

import (
	"errors"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterTag(tagName string, minNum int) (model.Tag, error) {
	// tagNameをもとに滞在ウォッチからタグ情報を取得
	tag := model.Tag{
		Name:      tagName,
		MinNumber: 0,
	}
	if err := tag.ReadByName(); err != nil {
		return tag, err
	}
	if tag.MinNumber != 0 {
		err := errors.New("tag already exists")
		return tag, err
	}
	tag.MinNumber = minNum
	// DBにタグ情報を登録
	if err := tag.Create(); err != nil {
		return tag, err
	}
	return tag, nil
}

func GetTags() ([]model.Tag, error) {
	var t *model.Tag
	tags, err := t.ReadAll()
	if err != nil {
		return nil, err
	}
	return tags, nil
}
