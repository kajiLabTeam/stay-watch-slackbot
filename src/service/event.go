package service

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterEvent(name string, minNumber int, typeID uint, toolIDs []uint) (model.Event, error) {
	event := model.Event{
		Name:      name,
		MinNumber: minNumber,
		TypeID:    typeID,
	}

	// Toolsを取得してeventに関連付ける
	if len(toolIDs) > 0 {
		var tools []model.Tool
		for _, id := range toolIDs {
			tool := model.Tool{}
			tool.ID = id
			if err := tool.ReadByID(); err != nil {
				return event, err
			}
			tools = append(tools, tool)
		}
		event.Tools = tools
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
