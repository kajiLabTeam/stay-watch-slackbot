package service

import (
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
	"github.com/kajiLabTeam/stay-watch-slackbot/prediction"
)

// ActivityTimeRange 活動の予測時間帯
type ActivityTimeRange struct {
	Start string // "HH:MM"
	End   string // "HH:MM"
}

// getActivityProbability イベントごとの活動確率を取得する
func getActivityProbability(eventID uint, dayOfWeek time.Weekday, targetTime string) (float64, error) {
	// 1. ログを取得
	logs, weeks, err := model.ReadLogsByEventIDAndDayOfWeek(eventID, dayOfWeek)
	if err != nil || len(logs) == 0 {
		return 0.0, nil // データ不足時は 0.0 を返す
	}

	// 2. Status が "start" のログをフィルタリング
	var timeStrings []string
	for _, log := range logs {
		if log.Status.Name == "start" {
			timeStr := log.CreatedAt.Format("15:04")
			timeStrings = append(timeStrings, timeStr)
		}
	}

	if len(timeStrings) == 0 {
		return 0.0, nil
	}

	// 3. prediction パッケージで確率計算
	probability, err := prediction.GetProbability(timeStrings, targetTime, weeks)
	if err != nil {
		return 0.0, err
	}

	return probability, nil
}

// getActivityTimeRange イベントの活動予測時刻範囲を取得する
func getActivityTimeRange(eventID uint, dayOfWeek time.Weekday) (ActivityTimeRange, error) {
	logs, weeks, err := model.ReadLogsByEventIDAndDayOfWeek(eventID, dayOfWeek)
	if err != nil || len(logs) == 0 {
		return ActivityTimeRange{Start: "00:00", End: "23:59"}, nil
	}

	// start と end のログを分離
	var startTimes []string
	var endTimes []string
	for _, log := range logs {
		timeStr := log.CreatedAt.Format("15:04")
		switch log.Status.Name {
		case "start":
			startTimes = append(startTimes, timeStr)
		case "end":
			endTimes = append(endTimes, timeStr)
		}
	}

	// 開始時刻の予測
	var startTime string
	if len(startTimes) > 0 {
		startMinutes, err := prediction.GetMostLikelyTime(startTimes, weeks)
		if err != nil {
			startTime = "00:00"
		} else {
			startTime = lib.MinutesToTime(startMinutes)
		}
	} else {
		startTime = "00:00"
	}

	// 終了時刻の予測
	var endTime string
	if len(endTimes) > 0 {
		endMinutes, err := prediction.GetMostLikelyTime(endTimes, weeks)
		if err != nil {
			endTime = "23:59"
		} else {
			endTime = lib.MinutesToTime(endMinutes)
		}
	} else {
		endTime = "23:59"
	}

	return ActivityTimeRange{Start: startTime, End: endTime}, nil
}

// getUserActivityEventIDs はユーザーが登録している活動のイベントIDセットを取得する
func getUserActivityEventIDs(userID uint) map[uint]bool {
	correspond := model.Correspond{UserID: userID}
	corresponds, err := correspond.ReadByUserID()
	if err != nil {
		return make(map[uint]bool)
	}

	eventIDs := make(map[uint]bool)
	for _, c := range corresponds {
		eventIDs[c.EventID] = true
	}

	return eventIDs
}

// filterByCommonActivities は受信者と共通の活動を持つユーザーをフィルタリングする（OR条件）
func filterByCommonActivities(candidates []model.User, receiverActivityEventIDs map[uint]bool, receiverUserID uint) []model.User {
	var filtered []model.User

	for _, candidate := range candidates {
		// 受信者自身は除外
		if candidate.ID == receiverUserID {
			continue
		}

		// 候補者の活動イベントIDを取得
		candidateEventIDs := getUserActivityEventIDs(candidate.ID)

		// 共通の活動があるかチェック（OR条件）
		hasCommonActivity := false
		for eventID := range candidateEventIDs {
			if receiverActivityEventIDs[eventID] {
				hasCommonActivity = true
				break
			}
		}

		if hasCommonActivity {
			filtered = append(filtered, candidate)
		}
	}

	return filtered
}

// getUserActivityTags はユーザーが登録している活動名のリストを取得する
func getUserActivityTags(userID uint) ([]string, error) {
	correspond := model.Correspond{UserID: userID}
	corresponds, err := correspond.ReadByUserID()
	if err != nil {
		// エラー時は空配列を返して処理を継続
		return []string{}, nil
	}

	var tags []string
	for _, c := range corresponds {
		tags = append(tags, c.Event.Name)
	}

	return tags, nil
}
