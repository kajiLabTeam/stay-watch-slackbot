package service

import (
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterEvent(name string, minNumber int) (model.Event, error) {
	event := model.Event{
		Name:      name,
		MinNumber: minNumber,
	}
	if err := event.Create(); err != nil {
		return event, err
	}
	return event, nil
}

func GetEvents() ([]model.Event, error) {
	var e model.Event
	events, err := e.ReadAll()
	if err != nil {
		return events, err
	}
	return events, nil
}
