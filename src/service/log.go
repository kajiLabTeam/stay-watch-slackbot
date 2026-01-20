package service

import (
	"fmt"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

// LogEntryInput はログ登録のための入力データを表す
type LogEntryInput struct {
	EventID   uint
	StatusID  uint
	CreatedAt string // JST形式 "2006-01-02 15:04:05"
}

// RegisterLog は単一のログを登録する（JST→UTC変換、外部キー検証）
func RegisterLog(input LogEntryInput) (model.Log, error) {
	// Event存在確認
	event := model.Event{}
	event.ID = input.EventID
	if err := event.ReadByID(); err != nil {
		return model.Log{}, fmt.Errorf("event_id %d not found", input.EventID)
	}

	// Status存在確認
	status := model.Status{}
	status.ID = input.StatusID
	if err := status.ReadByID(); err != nil {
		return model.Log{}, fmt.Errorf("status_id %d not found", input.StatusID)
	}

	// JST時刻をパースしてUTCに変換
	createdAtJST, err := time.ParseInLocation("2006-01-02 15:04:05", input.CreatedAt, lib.JST)
	if err != nil {
		return model.Log{}, fmt.Errorf("invalid created_at format: %s", input.CreatedAt)
	}
	createdAtUTC := createdAtJST.UTC()

	// ログを作成
	log := model.Log{
		EventID:  input.EventID,
		StatusID: input.StatusID,
	}
	// CreatedAtを手動で設定するためにgorm.Modelを直接操作
	log.CreatedAt = createdAtUTC

	if err := log.Create(); err != nil {
		return model.Log{}, err
	}

	return log, nil
}

// BatchRegisterLogs は複数のログを一括登録する
func BatchRegisterLogs(inputs []LogEntryInput) ([]model.Log, map[string]string, error) {
	var logs []model.Log
	errors := make(map[string]string)

	for i, input := range inputs {
		log, err := RegisterLog(input)
		if err != nil {
			errors[fmt.Sprintf("%d", i)] = err.Error()
			continue
		}
		logs = append(logs, log)
	}

	return logs, errors, nil
}
