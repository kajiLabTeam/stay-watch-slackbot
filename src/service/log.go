package service

import (
	"fmt"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

// LogEntryInput はログ登録のための入力データを表す
type LogEntryInput struct {
	EventID                uint
	StatusID               uint
	EventTime              string  // RFC3339形式 JST (例: "2006-01-02T15:04:05+09:00")
	ParticipateStayWatchIDs []int64 // 参加メンバの stay_watch_id（空可）
	RoomStayWatchIDs        []int64 // 在室メンバの stay_watch_id（空可）
}

// resolveUserIDs は stay_watch_id のスライスから対応する内部 user_id のスライスを取得する
// 解決できない stay_watch_id があればエラーを返す
func resolveUserIDs(stayWatchIDs []int64) ([]uint, error) {
	if len(stayWatchIDs) == 0 {
		return nil, nil
	}
	u := model.User{}
	users, err := u.ReadByStayWatchIDs(stayWatchIDs)
	if err != nil {
		return nil, err
	}
	if len(users) != len(stayWatchIDs) {
		found := make(map[int64]bool, len(users))
		for _, user := range users {
			found[user.StayWatchID] = true
		}
		var missing []int64
		for _, id := range stayWatchIDs {
			if !found[id] {
				missing = append(missing, id)
			}
		}
		return nil, fmt.Errorf("unknown stay_watch_id(s): %v", missing)
	}
	userIDs := make([]uint, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}
	return userIDs, nil
}

// RegisterLog は単一のログを登録する（JST検証付き）
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

	// 時刻をパース（JSTのみ許可）
	eventTimeJST, err := lib.ParseJST(input.EventTime)
	if err != nil {
		return model.Log{}, fmt.Errorf("invalid event_time: %v", err)
	}

	// stay_watch_id を内部 user_id に解決
	roomUserIDs, err := resolveUserIDs(input.RoomStayWatchIDs)
	if err != nil {
		return model.Log{}, fmt.Errorf("room_users: %v", err)
	}
	participateUserIDs, err := resolveUserIDs(input.ParticipateStayWatchIDs)
	if err != nil {
		return model.Log{}, fmt.Errorf("participate_users: %v", err)
	}

	// ログを作成（中間テーブル含めトランザクション）
	log := model.Log{
		EventID:   input.EventID,
		StatusID:  input.StatusID,
		EventTime: eventTimeJST,
	}

	if err := log.CreateWithUsers(roomUserIDs, participateUserIDs); err != nil {
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
