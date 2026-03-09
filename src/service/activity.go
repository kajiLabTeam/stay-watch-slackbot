package service

import (
	"fmt"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
	"github.com/kajiLabTeam/stay-watch-slackbot/prediction"
)

// ActivityProbability は活動名と1時間ごとの発生確率を表す
type ActivityProbability struct {
	ActivityName  string    `json:"activity_name"`
	Probabilities []float64 `json:"probabilities"` // length 24, index = hour (0-23 JST), value = 0.0〜1.0
}

// ActivityTimeRange は活動の予測時間帯を表す
type ActivityTimeRange struct {
	Start string // "HH:MM"
	End   string // "HH:MM"
}

// calculateWeeks 最初のログから今日までの週数を計算する
func calculateWeeks(logs []model.Log) int {
	if len(logs) == 0 {
		return 0
	}
	oldestLog := logs[0]
	for _, log := range logs {
		if log.CreatedAt.Before(oldestLog.CreatedAt) {
			oldestLog = log
		}
	}
	now := time.Now().UTC()
	days := int(now.Sub(oldestLog.CreatedAt).Hours() / 24)
	return (days / 7) + 1
}

// GetActivityProbability イベントごとの活動確率を取得する
func GetActivityProbability(eventID uint, dayOfWeek time.Weekday, targetTime string) (float64, error) {
	// 1. ログを取得
	logs, err := model.ReadLogsByEventIDAndDayOfWeek(eventID, dayOfWeek)
	if err != nil || len(logs) == 0 {
		return 0.0, nil // データ不足時は 0.0 を返す
	}

	// 2. 週数を計算
	weeks := calculateWeeks(logs)

	// 3. Status が "start" のログをフィルタリング（日付付き）
	// DBはUTCなのでそのまま使用（targetTimeもUTC）
	var datetimeStrings []string
	for _, log := range logs {
		if log.Status.Name == "start" {
			datetimeStr := log.CreatedAt.Format("2006-01-02 15:04")
			datetimeStrings = append(datetimeStrings, datetimeStr)
		}
	}

	if len(datetimeStrings) == 0 {
		return 0.0, nil
	}

	// 3. prediction パッケージで確率計算（日付重複を排除）
	probability, err := prediction.GetProbabilityByUniqueDate(datetimeStrings, targetTime, weeks)
	if err != nil {
		return 0.0, err
	}

	return probability, nil
}

// getActivityTimeRange イベントの活動予測時刻範囲を取得する
func getActivityTimeRange(eventID uint, dayOfWeek time.Weekday) (ActivityTimeRange, error) {
	logs, err := model.ReadLogsByEventIDAndDayOfWeek(eventID, dayOfWeek)
	if err != nil || len(logs) == 0 {
		return ActivityTimeRange{Start: "00:00", End: "23:59"}, nil
	}

	weeks := calculateWeeks(logs)

	// start と end のログを分離
	// Slack表示用にJSTに変換
	var startTimes []string
	var endTimes []string
	for _, log := range logs {
		timeStr := lib.FormatTimeJST(log.CreatedAt)
		switch log.Status.Name {
		case "start":
			startTimes = append(startTimes, timeStr)
		case "end":
			endTimes = append(endTimes, timeStr)
		}
	}

	startTime := predictTime(startTimes, weeks, "00:00")
	endTime := predictTime(endTimes, weeks, "23:59")

	return ActivityTimeRange{Start: startTime, End: endTime}, nil
}

// predictTime は時刻リストから最尤時刻を予測する。データ不足やエラー時はデフォルト値を返す
func predictTime(times []string, weeks int, defaultTime string) string {
	if len(times) == 0 {
		return defaultTime
	}
	minutes, err := prediction.GetMostLikelyTime(times, weeks)
	if err != nil {
		return defaultTime
	}
	return lib.MinutesToTime(minutes)
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

// extractStartDatetimes は "start" ステータスのログからJST日時文字列を抽出する
func extractStartDatetimes(logs []model.Log, loc *time.Location) []string {
	var datetimeStrings []string
	for _, log := range logs {
		if log.Status.Name == "start" {
			datetimeStrings = append(datetimeStrings, log.CreatedAt.In(loc).Format("2006-01-02 15:04"))
		}
	}
	return datetimeStrings
}

// calcHourlyProbabilities は各時間帯（JST 0〜23時）の確率を計算する
// H時 = CDF(H:30) - CDF((H-1):30) で (H-1):30〜H:30 の確率密度合計を求める
func calcHourlyProbabilities(datetimeStrings []string, weeks int) []float64 {
	probabilities := make([]float64, 24)
	for hour := 0; hour < 24; hour++ {
		probabilities[hour] = calcHourProbability(datetimeStrings, hour, weeks)
	}
	return probabilities
}

// calcHourProbability は指定時間帯の確率を計算する
func calcHourProbability(datetimeStrings []string, hour int, weeks int) float64 {
	endTimeJST := fmt.Sprintf("%02d:30", hour)
	startTimeJST := fmt.Sprintf("%02d:30", (hour-1+24)%24)

	cdfEnd, err := prediction.GetProbabilityByUniqueDate(datetimeStrings, endTimeJST, weeks)
	if err != nil {
		return 0.0
	}
	cdfStart, err := prediction.GetProbabilityByUniqueDate(datetimeStrings, startTimeJST, weeks)
	if err != nil {
		return 0.0
	}

	prob := cdfEnd - cdfStart
	if prob < 0 {
		return 0.0
	}
	if prob > 1.0 {
		return 1.0
	}
	return prob
}

// calcEventProbability はイベント1件分の活動確率を計算する
func calcEventProbability(ev model.Event, dayOfWeek time.Weekday) ActivityProbability {
	logs, err := model.ReadLogsByEventIDAndDayOfWeek(ev.ID, dayOfWeek)
	if err != nil || len(logs) == 0 {
		return ActivityProbability{ActivityName: ev.Name, Probabilities: make([]float64, 24)}
	}

	weeks := calculateWeeks(logs)
	datetimeStrings := extractStartDatetimes(logs, lib.JST)
	if len(datetimeStrings) == 0 {
		return ActivityProbability{ActivityName: ev.Name, Probabilities: make([]float64, 24)}
	}

	return ActivityProbability{
		ActivityName:  ev.Name,
		Probabilities: calcHourlyProbabilities(datetimeStrings, weeks),
	}
}

// GetAllActivityProbabilities は全活動の1時間ごとの発生確率を取得する
// 各時間帯（JST H時）について、(H-1):30〜H:30 の範囲の確率密度合計を計算する
// 例: 12時の場合、CDF(12:30) - CDF(11:30) で 11:30〜12:30 の確率を求める
func GetAllActivityProbabilities(dayOfWeek time.Weekday) ([]ActivityProbability, error) {
	event := model.Event{}
	events, err := event.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	var results []ActivityProbability
	for _, ev := range events {
		results = append(results, calcEventProbability(ev, dayOfWeek))
	}
	return results, nil
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
