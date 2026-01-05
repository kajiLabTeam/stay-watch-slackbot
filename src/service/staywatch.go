package service

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
	"github.com/kajiLabTeam/stay-watch-slackbot/prediction"
)

func GetStayWatchMember() ([]StaywatchUsers, error) {
	var users []StaywatchUsers
	if err := stayWatchClient.Get(staywatch.BaseURL+staywatch.Users, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func GetStayWatchProbability(users []model.User, weekday time.Weekday) []Probability {
	timeStr := "23:59"
	w := int(weekday)
	u, _ := url.Parse(staywatch.BaseURL + staywatch.Probability + "/visit")
	q := u.Query()
	for _, user := range users {
		q.Add("user-id", strconv.Itoa(int(user.StayWatchID)))
	}
	q.Add("weekday", strconv.Itoa(w))
	q.Add("time", timeStr)
	u.RawQuery = q.Encode()

	var r StayWatchResponse
	if err := stayWatchClient.Get(u.String(), &r); err != nil {
		return []Probability{}
	}

	var probability []Probability
	for _, result := range r.Result {
		probability = append(probability, Probability{
			UserId:      int(result.UserID),
			Probability: result.Probability,
		})
	}
	return probability
}

func filterByThreshold(pro []Probability, threshold float64) []model.User {
	var filtered []model.User
	for _, u := range pro {
		if u.Probability >= threshold {
			user := model.User{
				StayWatchID: int64(u.UserId),
			}
			user.ReadByStayWatchID()
			filtered = append(filtered, user)
		}
	}
	return filtered
}

func fetchPredictionTime(users []model.User, weekday time.Weekday, action string) []Result {
	w := int(weekday)
	u, _ := url.Parse(staywatch.BaseURL + staywatch.Time + "/" + action)
	q := u.Query()
	for _, user := range users {
		q.Add("user-id", strconv.Itoa(int(user.StayWatchID)))
	}
	q.Add("weekday", strconv.Itoa(w))
	u.RawQuery = q.Encode()

	var r StayWatchResponse
	if err := stayWatchClient.Get(u.String(), &r); err != nil {
		return []Result{}
	}
	return r.Result
}

func mergePredictions(vr []Result, dr []Result) []Prediction {
	var predictions []Prediction
	for _, v := range vr {
		for _, d := range dr {
			if v.UserID == d.UserID {
				predictions = append(predictions, Prediction{
					UserID:    v.UserID,
					Visit:     v.PredictionTime,
					Departure: d.PredictionTime,
				})
				break
			}
		}
	}
	return predictions
}

type TimeRange struct {
	Start string
	End   string
}

// ActivityTimeRange 活動の予測時間帯
type ActivityTimeRange struct {
	Start string // "HH:MM"
	End   string // "HH:MM"
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
		if log.Status.Name == "start" {
			startTimes = append(startTimes, timeStr)
		} else if log.Status.Name == "end" {
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

// calculateRecommendedTimeRanges 活動推奨時間を計算する
func calculateRecommendedTimeRanges(activityRange ActivityTimeRange, occupancyRanges []TimeRange) []TimeRange {
	var recommendedRanges []TimeRange

	// activityRange を分単位に変換
	activityStartMinutes, err1 := lib.TimeToMinutes(activityRange.Start)
	activityEndMinutes, err2 := lib.TimeToMinutes(activityRange.End)
	if err1 != nil || err2 != nil {
		return recommendedRanges
	}

	// 各 occupancy range との重なりを計算
	for _, occupancy := range occupancyRanges {
		occupancyStartMinutes, err1 := lib.TimeToMinutes(occupancy.Start)
		occupancyEndMinutes, err2 := lib.TimeToMinutes(occupancy.End)
		if err1 != nil || err2 != nil {
			continue
		}

		// 重なりの計算
		overlapStart := max(activityStartMinutes, occupancyStartMinutes)
		overlapEnd := min(activityEndMinutes, occupancyEndMinutes)

		if overlapStart < overlapEnd {
			recommendedRanges = append(recommendedRanges, TimeRange{
				Start: lib.MinutesToTime(overlapStart),
				End:   lib.MinutesToTime(overlapEnd),
			})
		}
	}

	return recommendedRanges
}

func findOverlappingRanges(predictions []Prediction, users []model.User, minNum int) []TimeRange {
	userSet := make(map[int64]bool)
	for _, u := range users {
		userSet[u.StayWatchID] = true
	}

	type event struct {
		Time  string
		Delta int
	}
	var events []event
	for _, p := range predictions {
		if !userSet[p.UserID] {
			continue
		}
		events = append(events, event{Time: p.Visit, Delta: +1})
		events = append(events, event{Time: p.Departure, Delta: -1})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Time < events[j].Time
	})

	var ranges []TimeRange
	current := 0
	var start string
	for _, e := range events {
		prev := current
		current += e.Delta

		if prev < minNum && current >= minNum {
			start = e.Time
		} else if prev >= minNum && current < minNum {
			ranges = append(ranges, TimeRange{
				Start: start,
				End:   e.Time,
			})
			start = ""
		}
	}

	if start != "" {
		ranges = append(ranges, TimeRange{
			Start: start,
			End:   "23:59",
		})
	}

	return ranges
}

// EventActivity は各イベントの活動情報を保持する
type EventActivity struct {
	EventID           uint
	EventName         string
	RecommendedRanges []TimeRange
	FilteredUsers     []model.User
	Predictions       []Prediction
}

func NotifyByEvent(targetWeekday time.Weekday) ([]model.User, map[int]map[int][]string) {
	// UserID → EventID → messages の構造
	userMessages := make(map[int]map[int][]string)

	// Step 1: 全イベントを Corresponds.User を含めて取得（N+1 クエリ問題を回避）
	var e model.Event
	events, err := e.ReadAllWithUsers()
	if err != nil {
		var u model.User
		users, _ := u.ReadAll()
		return users, userMessages
	}

	// 各ユーザーに対して、参加する全イベントの活動情報を収集
	userEventActivities := make(map[uint][]EventActivity)

	// Step 2: 各イベントに対して処理
	for _, event := range events {
		// Step 2a: 活動確率を取得（閾値チェック）
		probability, err := getActivityProbability(event.ID, targetWeekday, "23:59")
		if err != nil || probability < 0.30 {
			continue // 確率が閾値未満の場合スキップ
		}

		// Step 2b: 活動予測時刻範囲を取得
		activityRange, err := getActivityTimeRange(event.ID, targetWeekday)
		if err != nil {
			continue
		}

		// Step 2c: イベント参加ユーザーを取得（Preload 済みデータを使用）
		var eventUsers []model.User
		for _, c := range event.Corresponds {
			eventUsers = append(eventUsers, c.User)
		}

		if len(eventUsers) == 0 {
			continue
		}

		// Step 2d: ユーザーの来訪確率・時刻を取得
		probs := GetStayWatchProbability(eventUsers, targetWeekday)
		filtered := filterByThreshold(probs, 0.3)

		if len(filtered) == 0 {
			continue
		}

		visitTimes := fetchPredictionTime(filtered, targetWeekday, "visit")
		departureTimes := fetchPredictionTime(filtered, targetWeekday, "departure")
		predictions := mergePredictions(visitTimes, departureTimes)

		// Step 2e: 規定人数在室時間を計算
		occupancyRanges := findOverlappingRanges(predictions, eventUsers, event.MinNumber)

		if len(occupancyRanges) == 0 {
			continue
		}

		// Step 2f: 活動推奨時間を計算
		recommendedRanges := calculateRecommendedTimeRanges(activityRange, occupancyRanges)

		if len(recommendedRanges) == 0 {
			continue
		}

		// Step 2g: 各ユーザーの活動情報を収集
		activity := EventActivity{
			EventID:           event.ID,
			EventName:         event.Name,
			RecommendedRanges: recommendedRanges,
			FilteredUsers:     filtered,
			Predictions:       predictions,
		}

		for _, eventUser := range eventUsers {
			userEventActivities[eventUser.ID] = append(userEventActivities[eventUser.ID], activity)
		}
	}

	// Step 3: ユーザーごとにメッセージを生成
	for userID, activities := range userEventActivities {
		if len(activities) == 0 {
			continue
		}

		// ユーザーの活動イベントIDを取得
		receiverActivityEventIDs := getUserActivityEventIDs(userID)

		// 全アクティビティの時間帯を先にリストアップ
		var activityHeaders []string
		allFilteredUsers := make(map[int64]model.User) // StayWatchID → User
		var allPredictions []Prediction

		for _, activity := range activities {
			for _, r := range activity.RecommendedRanges {
				header := fmt.Sprintf("%s〜%s  `%s`", r.Start, r.End, activity.EventName)
				activityHeaders = append(activityHeaders, header)
			}

			// 全ユーザーと予測情報を収集（重複排除）
			for _, user := range activity.FilteredUsers {
				allFilteredUsers[user.StayWatchID] = user
			}
			allPredictions = append(allPredictions, activity.Predictions...)
		}

		// 収集したユーザーをスライスに変換
		var uniqueFilteredUsers []model.User
		for _, user := range allFilteredUsers {
			uniqueFilteredUsers = append(uniqueFilteredUsers, user)
		}

		// 受信者と共通の活動を持つユーザーだけをフィルタリング
		commonActivityUsers := filterByCommonActivities(uniqueFilteredUsers, receiverActivityEventIDs, userID)

		if len(commonActivityUsers) == 0 {
			continue
		}

		// メッセージを組み立て
		var msgBuilder strings.Builder

		// 1. 全アクティビティの時間帯を記載
		for _, header := range activityHeaders {
			msgBuilder.WriteString(header + "\n")
		}

		// 2. 「来そうな人」セクションを追加
		msgBuilder.WriteString("\n来そうな人↓\n")

		// ユーザーを名前順にソート
		sort.Slice(commonActivityUsers, func(i, j int) bool {
			return commonActivityUsers[i].Name < commonActivityUsers[j].Name
		})

		// ループ前に全ユーザーのタグを一括取得（DB クエリの重複を防ぐ）
		activityTagsCache := make(map[uint][]string)
		for _, user := range commonActivityUsers {
			tags, err := getUserActivityTags(user.ID)
			if err != nil {
				tags = []string{}
			}
			activityTagsCache[user.ID] = tags
		}

		for _, user := range commonActivityUsers {
			// 予測時刻を取得（最初に見つかったものを使用）
			visit, departure, found := findPredictionForUser(user.StayWatchID, allPredictions)
			if !found {
				continue
			}

			// キャッシュからユーザーの活動タグを取得
			activityTags := activityTagsCache[user.ID]

			// ユーザー行を生成
			userLine := formatUserLine(user.Name, activityTags, visit, departure)
			msgBuilder.WriteString(userLine + "\n")
		}

		msg := msgBuilder.String()

		// UserID → EventID → messages の構造に追加
		// 全イベントに対して同じメッセージを送る
		if userMessages[int(userID)] == nil {
			userMessages[int(userID)] = make(map[int][]string)
		}
		// 最初のアクティビティのEventIDをキーとして使用
		userMessages[int(userID)][int(activities[0].EventID)] = append(
			userMessages[int(userID)][int(activities[0].EventID)],
			msg,
		)
	}

	// 全ユーザーを返却（既存の Slack DM 送信との互換性のため）
	var u model.User
	users, _ := u.ReadAll()
	return users, userMessages
}

// findPredictionForUser はユーザーの予測滞在時間を検索する
func findPredictionForUser(stayWatchID int64, predictions []Prediction) (visit, departure string, found bool) {
	for _, p := range predictions {
		if p.UserID == stayWatchID {
			return p.Visit, p.Departure, true
		}
	}
	return "", "", false
}

// formatUserLine はユーザー情報を1行のフォーマットされた文字列にする
func formatUserLine(userName string, activityTags []string, visit, departure string) string {
	var tagsStr string
	if len(activityTags) > 0 {
		for i, tag := range activityTags {
			if i > 0 {
				tagsStr += " "
			}
			tagsStr += fmt.Sprintf("`%s`", tag)
		}
	}

	// タグがある場合とない場合でスペースを調整
	if tagsStr != "" {
		return fmt.Sprintf("%s %s  %s~%s", userName, tagsStr, visit, departure)
	}
	return fmt.Sprintf("%s  %s~%s", userName, visit, departure)
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

// buildUpcomingUsersSection は「来そうな人」セクションを生成する
func buildUpcomingUsersSection(
	filtered []model.User,
	predictions []Prediction,
	currentEventID uint,
	receiverUserID uint,
	receiverActivityEventIDs map[uint]bool,
) string {
	if len(filtered) == 0 {
		return ""
	}

	// 受信者と共通の活動を持つユーザーだけをフィルタリング
	commonActivityUsers := filterByCommonActivities(filtered, receiverActivityEventIDs, receiverUserID)

	if len(commonActivityUsers) == 0 {
		return ""
	}

	// ユーザーを名前順にソート
	sortedUsers := make([]model.User, len(commonActivityUsers))
	copy(sortedUsers, commonActivityUsers)
	sort.Slice(sortedUsers, func(i, j int) bool {
		return sortedUsers[i].Name < sortedUsers[j].Name
	})

	var section strings.Builder
	section.WriteString("\n来そうな人↓\n")

	// ループ前に全ユーザーのタグを一括取得（DB クエリの重複を防ぐ）
	activityTagsCache := make(map[uint][]string)
	for _, user := range sortedUsers {
		tags, err := getUserActivityTags(user.ID)
		if err != nil {
			tags = []string{}
		}
		activityTagsCache[user.ID] = tags
	}

	for _, user := range sortedUsers {
		// 予測時刻を取得
		visit, departure, found := findPredictionForUser(user.StayWatchID, predictions)
		if !found {
			// 予測がない場合はスキップ
			continue
		}

		// キャッシュからユーザーの活動タグを取得
		activityTags := activityTagsCache[user.ID]

		// ユーザー行を生成
		userLine := formatUserLine(user.Name, activityTags, visit, departure)
		section.WriteString(userLine + "\n")
	}

	return section.String()
}
