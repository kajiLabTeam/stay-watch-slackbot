package service

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterStatus(name string) (model.Status, error) {
	status := model.Status{
		Name: name,
	}
	if err := status.Create(); err != nil {
		// MySQLのユニーク制約エラー（1062）を型安全に判定
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return status, errors.New("status already exists")
		}
		return status, err
	}
	return status, nil
}

func GetStatuses() ([]model.Status, error) {
	var s model.Status
	statuses, err := s.ReadAll()
	if err != nil {
		return statuses, err
	}
	return statuses, nil
}
