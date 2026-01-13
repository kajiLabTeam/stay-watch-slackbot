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

// BatchRegisterTools は複数のToolを一括登録する
func BatchRegisterTools(names []string) ([]model.Tool, map[string]string, error) {
	var tools []model.Tool
	errors := make(map[string]string)

	for _, name := range names {
		tool, err := RegisterTool(name)
		if err != nil {
			errors[name] = err.Error()
			continue
		}
		tools = append(tools, tool)
	}

	return tools, errors, nil
}
