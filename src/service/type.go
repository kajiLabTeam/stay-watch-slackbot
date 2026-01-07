package service

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterType(name string) (model.Type, error) {
	typeObj := model.Type{
		Name: name,
	}
	if err := typeObj.Create(); err != nil {
		// MySQLのユニーク制約エラー（1062）を型安全に判定
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return typeObj, errors.New("type already exists")
		}
		return typeObj, err
	}
	return typeObj, nil
}

func GetTypes() ([]model.Type, error) {
	var t model.Type
	types, err := t.ReadAll()
	if err != nil {
		return types, err
	}
	return types, nil
}
