package service

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterTool(name string) (model.Tool, error) {
	tool := model.Tool{
		Name: name,
	}
	if err := tool.Create(); err != nil {
		// MySQLのユニーク制約エラー（1062）を型安全に判定
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return tool, errors.New("tool already exists")
		}
		return tool, err
	}
	return tool, nil
}

func GetTools() ([]model.Tool, error) {
	var t model.Tool
	tools, err := t.ReadAll()
	if err != nil {
		return tools, err
	}
	return tools, nil
}
