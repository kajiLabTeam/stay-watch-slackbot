package service

import (
	"fmt"
	"math"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
	"github.com/kajiLabTeam/stay-watch-slackbot/prediction"
)

// ActivityProbability は活動名と1時間ごとの発生確率を表す
type ActivityProbability struct {
	EventID       uint      `json:"event_id"`
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
		if log.EventTime.Before(oldestLog.EventTime) {
			oldestLog = log
		}
	}
	now := lib.NowJST()
	days := int(now.Sub(oldestLog.EventTime).Hours() / 24)
	return (days / 7) + 1
}

// GetActivityProbability イベントごとの活動確率を取得する
// 「targetTime 時点で活動が進行中である確率」= P_start(targetTime) - P_end(targetTime) を返す。
// 開始時刻の分布だけでなく終了時刻の分布も見ることで、継続時間の長い活動が
// 開始直後のクラスタ付近でしか高確率にならない問題（進行中でももう始まってから
// 時間が経っていると過小評価される）を避ける。
func GetActivityProbability(eventID uint, dayOfWeek time.Weekday, targetTime string) (float64, error) {
	// 1. ログを取得
	logs, err := model.ReadLogsByEventIDAndDayOfWeek(eventID, dayOfWeek)
	if err != nil || len(logs) == 0 {
		return 0.0, nil // データ不足時は 0.0 を返す
	}

	// 2. 週数を計算
	weeks := calculateWeeks(logs)

	// 3. Status が "start"/"end" のログをそれぞれ抽出（日付重複排除は
	//    NewProbabilityModelByUniqueDate が内部で行う）
	startDatetimes := extractStartDatetimes(logs)
	if len(startDatetimes) == 0 {
		return 0.0, nil
	}
	endDatetimes := extractEndDatetimes(logs)

	// 4. 開始・終了それぞれの分布モデルを構築し、進行中確率を計算
	startModel, err := prediction.NewProbabilityModelByUniqueDate(startDatetimes)
	if err != nil {
		return 0.0, err
	}
	endModel, err := prediction.NewProbabilityModelByUniqueDate(endDatetimes)
	if err != nil {
		return 0.0, err
	}

	return calcOngoingProbability(startModel, endModel, targetTime, weeks)
}

// calcOngoingProbability は startModel/endModel から「時刻 t に活動が進行中である確率」
// P_ongoing(t) = P_start(t) - P_end(t) を計算する（0〜1 にクランプ）
func calcOngoingProbability(startModel, endModel *prediction.ProbabilityModel, timeStr string, weeks int) (float64, error) {
	startProb, err := startModel.Probability(timeStr, weeks)
	if err != nil {
		return 0, err
	}
	endProb, err := endModel.Probability(timeStr, weeks)
	if err != nil {
		return 0, err
	}
	return clampHourProbability(startProb - endProb), nil
}

// getActivityTimeRange イベントの活動予測時刻範囲を取得する
func getActivityTimeRange(eventID uint, dayOfWeek time.Weekday) (ActivityTimeRange, error) {
	logs, err := model.ReadLogsByEventIDAndDayOfWeek(eventID, dayOfWeek)
	if err != nil || len(logs) == 0 {
		return ActivityTimeRange{Start: "00:00", End: "23:59"}, nil
	}

	weeks := calculateWeeks(logs)

	// start と end のログを分離（DBはJSTなのでそのまま使用）
	var startTimes []string
	var endTimes []string
	for _, log := range logs {
		timeStr := lib.FormatTime(log.EventTime)
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
	eventUser := model.EventUser{UserID: userID}
	eventUsers, err := eventUser.ReadByUserID()
	if err != nil {
		return make(map[uint]bool)
	}

	eventIDs := make(map[uint]bool)
	for _, eu := range eventUsers {
		eventIDs[eu.EventID] = true
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
func extractStartDatetimes(logs []model.Log) []string {
	return extractDatetimesByStatus(logs, "start")
}

// extractEndDatetimes は "end" ステータスのログからJST日時文字列を抽出する
func extractEndDatetimes(logs []model.Log) []string {
	return extractDatetimesByStatus(logs, "end")
}

func extractDatetimesByStatus(logs []model.Log, statusName string) []string {
	var datetimeStrings []string
	for _, log := range logs {
		if log.Status.Name == statusName {
			datetimeStrings = append(datetimeStrings, log.EventTime.Format("2006-01-02 15:04"))
		}
	}
	return datetimeStrings
}

// calcHourlyProbabilities は各時間帯（JST 0〜23時）の「活動が進行中である確率」を計算する。
// その時間帯の代表時刻 H:30 における P_ongoing(H:30) = P_start(H:30) - P_end(H:30) を使う。
// 開始ログだけでなく終了ログの分布も見ることで、継続時間の長い活動が開始クラスタ付近
// でしか高確率にならない問題（進行中でも開始から時間が経つと過小評価される）を避ける。
// 終了ログが無いイベントでは P_end が常に0になり、P_start のみの挙動にフォールバックする。
//
// クラスタリングは重い処理なのでモデルを1度だけ構築し、24時刻ぶんの評価に使い回す
func calcHourlyProbabilities(startDatetimes, endDatetimes []string, weeks int) []float64 {
	startModel, err := prediction.NewProbabilityModelFromDatetimes(startDatetimes)
	if err != nil {
		return make([]float64, 24)
	}
	endModel, err := prediction.NewProbabilityModelFromDatetimes(endDatetimes)
	if err != nil {
		return make([]float64, 24)
	}

	probabilities := make([]float64, 24)
	for hour := 0; hour < 24; hour++ {
		prob, err := calcOngoingProbability(startModel, endModel, fmt.Sprintf("%02d:30", hour), weeks)
		if err != nil {
			return make([]float64, 24)
		}
		probabilities[hour] = prob
	}
	return probabilities
}

// clampHourProbability は時間帯確率を 0.0〜1.0 の範囲に丸める
func clampHourProbability(prob float64) float64 {
	if math.IsNaN(prob) || math.IsInf(prob, 0) || prob < 0 {
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
		return ActivityProbability{EventID: ev.ID, ActivityName: ev.Name, Probabilities: make([]float64, 24)}
	}

	weeks := calculateWeeks(logs)
	startDatetimes := extractStartDatetimes(logs)
	if len(startDatetimes) == 0 {
		return ActivityProbability{EventID: ev.ID, ActivityName: ev.Name, Probabilities: make([]float64, 24)}
	}
	endDatetimes := extractEndDatetimes(logs)

	return ActivityProbability{
		EventID:       ev.ID,
		ActivityName:  ev.Name,
		Probabilities: calcHourlyProbabilities(startDatetimes, endDatetimes, weeks),
	}
}

// GetAllActivityProbabilities は全活動の1時間ごとの発生確率を取得する
// 各時間帯（JST H時）について、その代表時刻 H:30 時点で活動が進行中である確率
// P_ongoing(H:30) = P_start(H:30) - P_end(H:30) を計算する
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
	eventUser := model.EventUser{UserID: userID}
	eventUsers, err := eventUser.ReadByUserID()
	if err != nil {
		// エラー時は空配列を返して処理を継続
		return []string{}, nil
	}

	var tags []string
	for _, eu := range eventUsers {
		tags = append(tags, eu.Event.Name)
	}

	return tags, nil
}
