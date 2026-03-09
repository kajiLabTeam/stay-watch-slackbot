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
	userMessages := make(map[int]map[int][]string)

	var e model.Event
	events, err := e.ReadAllWithUsers()
	if err != nil {
		var u model.User
		users, _ := u.ReadAll()
		return users, userMessages
	}

	userEventActivities := collectUserEventActivities(events, targetWeekday)

	for userID, activities := range userEventActivities {
		msg := buildUserNotificationMessage(userID, activities)
		if msg == "" {
			continue
		}
		if userMessages[int(userID)] == nil {
			userMessages[int(userID)] = make(map[int][]string)
		}
		eventID := int(activities[0].EventID)
		userMessages[int(userID)][eventID] = append(userMessages[int(userID)][eventID], msg)
	}

	var u model.User
	users, _ := u.ReadAll()
	return users, userMessages
}

// collectUserEventActivities は各イベントを処理し、ユーザーごとの活動情報を収集する
func collectUserEventActivities(events []model.Event, targetWeekday time.Weekday) map[uint][]EventActivity {
	userEventActivities := make(map[uint][]EventActivity)

	for _, event := range events {
		activity, eventUsers, ok := processEvent(event, targetWeekday)
		if !ok {
			continue
		}
		for _, eventUser := range eventUsers {
			userEventActivities[eventUser.ID] = append(userEventActivities[eventUser.ID], activity)
		}
	}

	return userEventActivities
}

// processEvent は1イベントの活動確率・推奨時間を計算し、EventActivityを返す
func processEvent(event model.Event, targetWeekday time.Weekday) (EventActivity, []model.User, bool) {
	probability, err := GetActivityProbability(event.ID, targetWeekday, "08:59")
	if err != nil || probability < 0.30 {
		return EventActivity{}, nil, false
	}

	activityRange, err := getActivityTimeRange(event.ID, targetWeekday)
	if err != nil {
		return EventActivity{}, nil, false
	}

	var eventUsers []model.User
	for _, c := range event.Corresponds {
		eventUsers = append(eventUsers, c.User)
	}
	if len(eventUsers) == 0 {
		return EventActivity{}, nil, false
	}

	probs := GetStayWatchProbability(eventUsers, targetWeekday)
	filtered := filterByThreshold(probs, 0.3)
	if len(filtered) == 0 {
		return EventActivity{}, nil, false
	}

	visitTimes := fetchPredictionTime(filtered, targetWeekday, "visit")
	departureTimes := fetchPredictionTime(filtered, targetWeekday, "departure")
	predictions := mergePredictions(visitTimes, departureTimes)

	occupancyRanges := findOverlappingRanges(predictions, eventUsers, event.MinNumber)
	if len(occupancyRanges) == 0 {
		return EventActivity{}, nil, false
	}

	recommendedRanges := calculateRecommendedTimeRanges(activityRange, occupancyRanges)
	if len(recommendedRanges) == 0 {
		return EventActivity{}, nil, false
	}

	activity := EventActivity{
		EventID:           event.ID,
		EventName:         event.Name,
		RecommendedRanges: recommendedRanges,
		FilteredUsers:     filtered,
		Predictions:       predictions,
	}
	return activity, eventUsers, true
}

// aggregateActivities は複数のEventActivityからヘッダー・ユーザー・予測情報を集約する
func aggregateActivities(activities []EventActivity) ([]string, []model.User, []Prediction) {
	var headers []string
	allUsers := make(map[int64]model.User)
	var allPredictions []Prediction

	for _, activity := range activities {
		for _, r := range activity.RecommendedRanges {
			headers = append(headers, fmt.Sprintf("%s〜%s  `%s`", r.Start, r.End, activity.EventName))
		}
		for _, user := range activity.FilteredUsers {
			allUsers[user.StayWatchID] = user
		}
		allPredictions = append(allPredictions, activity.Predictions...)
	}

	var uniqueUsers []model.User
	for _, user := range allUsers {
		uniqueUsers = append(uniqueUsers, user)
	}

	return headers, uniqueUsers, allPredictions
}

// buildUserNotificationMessage はユーザー1人分の通知メッセージを生成する
func buildUserNotificationMessage(userID uint, activities []EventActivity) string {
	if len(activities) == 0 {
		return ""
	}

	receiverActivityEventIDs := getUserActivityEventIDs(userID)
	headers, uniqueUsers, allPredictions := aggregateActivities(activities)

	commonActivityUsers := filterByCommonActivities(uniqueUsers, receiverActivityEventIDs, userID)
	if len(commonActivityUsers) == 0 {
		return ""
	}

	sort.Slice(commonActivityUsers, func(i, j int) bool {
		return commonActivityUsers[i].Name < commonActivityUsers[j].Name
	})

	var msgBuilder strings.Builder
	for _, header := range headers {
		msgBuilder.WriteString(header + "\n")
	}
	msgBuilder.WriteString("\n来そうな人↓\n")

	activityTagsCache := buildActivityTagsCache(commonActivityUsers)

	for _, user := range commonActivityUsers {
		visit, departure, found := findPredictionForUser(user.StayWatchID, allPredictions)
		if !found {
			continue
		}
		userLine := formatUserLine(user.Name, activityTagsCache[user.ID], visit, departure)
		msgBuilder.WriteString(userLine + "\n")
	}

	return msgBuilder.String()
}

// buildActivityTagsCache はユーザーリストの活動タグを一括取得してキャッシュを返す
func buildActivityTagsCache(users []model.User) map[uint][]string {
	cache := make(map[uint][]string)
	for _, user := range users {
		tags, err := getUserActivityTags(user.ID)
		if err != nil {
			tags = []string{}
		}
		cache[user.ID] = tags
	}
	return cache
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
