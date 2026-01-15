package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

// EventActivity は各イベントの活動情報を保持する
type EventActivity struct {
	EventID           uint
	EventName         string
	RecommendedRanges []TimeRange
	FilteredUsers     []model.User
	Predictions       []Prediction
}

// NotifyByEvent はイベントベースの通知を生成する
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
		probability, err := GetActivityProbability(event.ID, targetWeekday, "23:59")
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
