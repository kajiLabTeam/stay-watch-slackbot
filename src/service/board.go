package service

import (
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/config"
	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

// moment-board の src/types.ts を契約とするDTO群

// BoardActivity は時間帯内の活動1件を表す
type BoardActivity struct {
	// 活動（イベント）名
	Name string `json:"name" example:"人狼"`
	// 発生確率の高さ
	Likelihood string `json:"likelihood" example:"high" enums:"high,mid,low"`
	// 想定人数
	Headcount int `json:"headcount" example:"4"`
}

// BoardPerson は時間帯内に来そうな人1件を表す
type BoardPerson struct {
	// ユーザー名
	Name string `json:"name" example:"山田太郎"`
	// アイコン画像URL
	AvatarURL string `json:"avatarUrl" example:"https://example.com/avatar.png"`
	// 来訪見込み度
	Arrival string `json:"arrival" example:"likely" enums:"likely,maybe"`
}

// BoardPresentMember は現在在室している人を表す
type BoardPresentMember struct {
	// ユーザー名
	Name string `json:"name" example:"山田太郎"`
	// アイコン画像URL
	AvatarURL string `json:"avatarUrl" example:"https://example.com/avatar.png"`
}

// BoardPresence は在室情報を表す（在室はフロントがStayWatchから直接取得するため常に空）
type BoardPresence struct {
	// 在室中メンバーの配列（常に空配列）
	Members []BoardPresentMember `json:"members"`
}

// BoardTimeBlock は1時間帯（昼/夕方/夜）を表す
type BoardTimeBlock struct {
	// 時間帯の識別子
	ID string `json:"id" example:"noon" enums:"noon,evening,night"`
	// 時間帯の表示名
	Label string `json:"label" example:"昼"`
	// 時間帯の範囲（表示用文字列）
	Range string `json:"range" example:"12:00-17:00"`
	// 現在時刻がこの時間帯に含まれるか
	IsNow bool `json:"isNow" example:"true"`
	// この時間帯に発生しうる活動一覧
	Activities []BoardActivity `json:"activities"`
	// この時間帯に来訪見込みのある人一覧
	People []BoardPerson `json:"people"`
}

// BoardData は共有モニター画面全体の表示データを表す
type BoardData struct {
	// データ生成時点の現在時刻（JST）
	CurrentTime string `json:"currentTime" example:"2006-01-02T15:04:05+09:00"`
	// 在室情報
	Presence BoardPresence `json:"presence"`
	// 時間帯（昼/夕方/夜）ごとの表示データ
	TimeBlocks []BoardTimeBlock `json:"timeBlocks"`
}

// timeBlockDef は時間帯の定義（分単位、JST）
type timeBlockDef = config.BoardTimeBlockConfig

// boardTimeBlocks は時間帯の定義一覧（環境変数で調整可能。config.TimeBlocks参照）
var boardTimeBlocks = config.TimeBlocks

// boardPersonAssign はユーザーの時間帯割当に必要な情報を保持する
type boardPersonAssign struct {
	user         model.User
	arrival      string
	visitMin     int // -1 = 予測なし
	departureMin int // -1 = 予測なし
}

// GetBoardData は共有モニター用の表示データを集約して返す
func GetBoardData() (BoardData, error) {
	now := lib.NowJST()
	weekday := now.Weekday()
	nowMin := now.Hour()*60 + now.Minute()

	assigns := collectBoardPeople(weekday)

	var e model.Event
	events, err := e.ReadAllWithUsers()
	if err != nil {
		return BoardData{}, err
	}

	activityProbs, err := GetAllActivityProbabilities(weekday)
	if err != nil {
		return BoardData{}, err
	}

	// イベント名 → 所属ユーザーID集合
	eventMembers := make(map[string]map[uint]bool)
	for _, ev := range events {
		members := make(map[uint]bool)
		for _, eu := range ev.EventUsers {
			members[eu.UserID] = true
		}
		eventMembers[ev.Name] = members
	}

	blocks := make([]BoardTimeBlock, 0, len(boardTimeBlocks))
	for _, def := range boardTimeBlocks {
		people := assignPeopleToBlock(assigns, def)

		// この時間帯に来そうな人のユーザーID集合（headcount計算用）
		blockUserIDs := make(map[uint]bool)
		for _, a := range assigns {
			if isAssignedToBlock(a, def) {
				blockUserIDs[a.user.ID] = true
			}
		}

		activities := buildBlockActivities(activityProbs, eventMembers, blockUserIDs, def)

		blocks = append(blocks, BoardTimeBlock{
			ID:         def.ID,
			Label:      def.Label,
			Range:      def.RangeLabel,
			IsNow:      nowMin >= def.StartMin && nowMin < def.EndMin,
			Activities: activities,
			People:     people,
		})
	}

	return BoardData{
		CurrentTime: now.Format("15:04"),
		Presence:    BoardPresence{Members: []BoardPresentMember{}},
		TimeBlocks:  blocks,
	}, nil
}

// collectBoardPeople は全ユーザーの来訪確率・予測時刻を取得し、時間帯割当用の情報を作る
func collectBoardPeople(weekday time.Weekday) []boardPersonAssign {
	var u model.User
	users, err := u.ReadAll()
	if err != nil || len(users) == 0 {
		return nil
	}

	probs := GetStayWatchProbability(users, weekday)

	// 来訪確率が maybe 閾値以上のユーザーのみ対象
	arrivalByStayWatchID := make(map[int64]string)
	var candidates []model.User
	userByStayWatchID := make(map[int64]model.User)
	for _, user := range users {
		userByStayWatchID[user.StayWatchID] = user
	}
	for _, p := range probs {
		if p.Probability < config.Board.ArrivalMaybe {
			continue
		}
		user, ok := userByStayWatchID[int64(p.UserID)]
		if !ok {
			continue
		}
		arrival := "maybe"
		if p.Probability >= config.Board.ArrivalLikely {
			arrival = "likely"
		}
		arrivalByStayWatchID[user.StayWatchID] = arrival
		candidates = append(candidates, user)
	}
	if len(candidates) == 0 {
		return nil
	}

	visitTimes := fetchPredictionTime(candidates, weekday, "visit")
	departureTimes := fetchPredictionTime(candidates, weekday, "departure")

	visitByID := predictionMinutesByUserID(visitTimes)
	departureByID := predictionMinutesByUserID(departureTimes)

	var assigns []boardPersonAssign
	for _, user := range candidates {
		visitMin, hasVisit := visitByID[user.StayWatchID]
		departureMin, hasDeparture := departureByID[user.StayWatchID]
		if !hasVisit {
			visitMin = -1
		}
		if !hasDeparture {
			departureMin = -1
		}
		assigns = append(assigns, boardPersonAssign{
			user:         user,
			arrival:      arrivalByStayWatchID[user.StayWatchID],
			visitMin:     visitMin,
			departureMin: departureMin,
		})
	}
	return assigns
}

// predictionMinutesByUserID は予測結果をユーザーID→分のマップに変換する。パース不能は除外
func predictionMinutesByUserID(results []Result) map[int64]int {
	m := make(map[int64]int)
	for _, r := range results {
		min, err := lib.TimeToMinutes(r.PredictionTime)
		if err != nil {
			continue
		}
		m[r.UserID] = min
	}
	return m
}

// isAssignedToBlock は予測時刻に基づきユーザーを時間帯に割り当てるか判定する
// - visit/departure 両方あり: 滞在区間と時間帯が重なれば表示
// - visit のみ: visit 以降のすべての時間帯に表示
// - departure のみ: departure までのすべての時間帯に表示
// - 両方なし: 非表示
func isAssignedToBlock(a boardPersonAssign, def timeBlockDef) bool {
	switch {
	case a.visitMin >= 0 && a.departureMin >= 0:
		return a.visitMin < def.EndMin && a.departureMin > def.StartMin
	case a.visitMin >= 0:
		return a.visitMin < def.EndMin
	case a.departureMin >= 0:
		return a.departureMin > def.StartMin
	default:
		return false
	}
}

// assignPeopleToBlock は時間帯に表示する人のリストを作る
func assignPeopleToBlock(assigns []boardPersonAssign, def timeBlockDef) []BoardPerson {
	people := []BoardPerson{}
	for _, a := range assigns {
		if !isAssignedToBlock(a, def) {
			continue
		}
		people = append(people, BoardPerson{
			Name:      a.user.Name,
			AvatarURL: a.user.IconURL,
			Arrival:   a.arrival,
		})
	}
	return people
}

// buildBlockActivities は時間帯内の活動リストを作る
// likelihood は時間帯内の時間別確率の最大値を2閾値で段階化し、最小閾値未満は表示しない
// headcount はその活動のメンバーのうち、この時間帯に来そうな人の数
func buildBlockActivities(probs []ActivityProbability, eventMembers map[string]map[uint]bool, blockUserIDs map[uint]bool, def timeBlockDef) []BoardActivity {
	activities := []BoardActivity{}
	for _, ap := range probs {
		maxProb := 0.0
		for hour := def.StartMin / 60; hour < (def.EndMin+59)/60 && hour < 24; hour++ {
			if ap.Probabilities[hour] > maxProb {
				maxProb = ap.Probabilities[hour]
			}
		}
		if maxProb < config.Board.LikelihoodMin {
			continue
		}

		likelihood := "low"
		if maxProb >= config.Board.LikelihoodHigh {
			likelihood = "high"
		} else if maxProb >= config.Board.LikelihoodMid {
			likelihood = "mid"
		}

		headcount := 0
		for userID := range eventMembers[ap.ActivityName] {
			if blockUserIDs[userID] {
				headcount++
			}
		}

		activities = append(activities, BoardActivity{
			Name:       ap.ActivityName,
			Likelihood: likelihood,
			Headcount:  headcount,
		})
	}
	return activities
}
