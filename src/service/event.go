package service

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterEvent(name string, minNumber int) (model.Event, error) {
	event := model.Event{
		Name:      name,
		MinNumber: minNumber,
	}
	if err := event.Create(); err != nil {
		// MySQLのユニーク制約エラー（1062）を型安全に判定
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return event, errors.New("event already exists")
		}
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
