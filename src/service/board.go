package service

import (
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

// moment-board の src/types.ts を契約とするDTO群

// BoardActivity は時間帯内の活動1件を表す
type BoardActivity struct {
	Name       string `json:"name"`
	Likelihood string `json:"likelihood"` // "high" | "mid" | "low"
	Headcount  int    `json:"headcount"`
}

// BoardPerson は時間帯内に来そうな人1件を表す
type BoardPerson struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	Arrival   string `json:"arrival"` // "likely" | "maybe"
}

// BoardPresentMember は現在在室している人を表す
type BoardPresentMember struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

// BoardPresence は在室情報を表す（在室はフロントがStayWatchから直接取得するため常に空）
type BoardPresence struct {
	Members []BoardPresentMember `json:"members"`
}

// BoardTimeBlock は1時間帯（昼/夕方/夜）を表す
type BoardTimeBlock struct {
	ID         string          `json:"id"` // "noon" | "evening" | "night"
	Label      string          `json:"label"`
	Range      string          `json:"range"`
	IsNow      bool            `json:"isNow"`
	Activities []BoardActivity `json:"activities"`
	People     []BoardPerson   `json:"people"`
}

// BoardData は共有モニター画面全体の表示データを表す
type BoardData struct {
	CurrentTime string           `json:"currentTime"`
	Presence    BoardPresence    `json:"presence"`
	TimeBlocks  []BoardTimeBlock `json:"timeBlocks"`
}

// timeBlockDef は時間帯の定義（分単位、JST）。範囲は結合テスト時に調整する仮置き
type timeBlockDef struct {
	id         string
	label      string
	rangeLabel string
	startMin   int // 含む
	endMin     int // 含まない
}

var boardTimeBlocks = []timeBlockDef{
	{id: "noon", label: "昼", rangeLabel: "12〜15時", startMin: 12 * 60, endMin: 15 * 60},
	{id: "evening", label: "夕方", rangeLabel: "15〜19時", startMin: 15 * 60, endMin: 19 * 60},
	{id: "night", label: "夜", rangeLabel: "19時〜", startMin: 19 * 60, endMin: 22 * 60},
}

// 段階化の閾値（結合テスト時に調整する仮置き）
const (
	likelihoodHighThreshold = 0.5
	likelihoodMidThreshold  = 0.3
	likelihoodMinThreshold  = 0.1 // これ未満の活動は表示しない
	arrivalLikelyThreshold  = 0.5
	arrivalMaybeThreshold   = 0.3
)

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
			ID:         def.id,
			Label:      def.label,
			Range:      def.rangeLabel,
			IsNow:      nowMin >= def.startMin && nowMin < def.endMin,
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
		if p.Probability < arrivalMaybeThreshold {
			continue
		}
		user, ok := userByStayWatchID[int64(p.UserID)]
		if !ok {
			continue
		}
		arrival := "maybe"
		if p.Probability >= arrivalLikelyThreshold {
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
		return a.visitMin < def.endMin && a.departureMin > def.startMin
	case a.visitMin >= 0:
		return a.visitMin < def.endMin
	case a.departureMin >= 0:
		return a.departureMin > def.startMin
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
		for hour := def.startMin / 60; hour < (def.endMin+59)/60 && hour < 24; hour++ {
			if ap.Probabilities[hour] > maxProb {
				maxProb = ap.Probabilities[hour]
			}
		}
		if maxProb < likelihoodMinThreshold {
			continue
		}

		likelihood := "low"
		if maxProb >= likelihoodHighThreshold {
			likelihood = "high"
		} else if maxProb >= likelihoodMidThreshold {
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
