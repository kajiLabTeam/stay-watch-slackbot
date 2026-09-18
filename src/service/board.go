package service

import (
	"log"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/config"
	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

// moment-board の src/types.ts を契約とするDTO群

const (
	// boardHourStart はタイムラインの最も早い時刻（この時刻より前は 11 時始まりに丸める）
	boardHourStart = 11
	// boardHourEnd はタイムラインに出す最も遅い時刻。これを超える列は出さない
	boardHourEnd = 19
	// boardHourCount はタイムラインに並べる列数
	boardHourCount = 4
)

// BoardPerson は来訪見込みのある人1件を表す
type BoardPerson struct {
	// ユーザー名
	Name string `json:"name" example:"山田太郎"`
	// アイコン画像URL
	AvatarURL string `json:"avatarUrl" example:"https://example.com/avatar.png"`
}

// BoardHour はタイムライン1時間ぶんの表示データを表す
type BoardHour struct {
	// 時（JST、0〜23）
	Hour int `json:"hour" example:"15"`
	// この時間に在室していそうな人一覧
	People []BoardPerson `json:"people"`
	// この時間に成立しそうな活動一覧
	Activities []BoardActivity `json:"activities"`
}

// BoardActivity はその時間帯に成立しそうな活動1件を表す
type BoardActivity struct {
	// イベントID
	ID uint `json:"id" example:"5"`
	// 活動（イベント）名
	Name string `json:"name" example:"人狼"`
	// 活動の画像URL。未登録の場合は null
	ImageURL *string `json:"imageUrl" example:"https://example.com/daycast/events/5.png"`
	// 活動の成立に必要な最低人数
	MinNumber int `json:"minNumber" example:"3"`
	// この活動に関心があり、かつその時間帯に在室していそうなメンバー
	Members []BoardPerson `json:"members"`
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

// BoardData は共有モニター画面全体の表示データを表す
type BoardData struct {
	// データ生成時点の現在時刻（JST、"HH:MM"形式）
	CurrentTime string `json:"currentTime" example:"15:04"`
	// 在室情報
	Presence BoardPresence `json:"presence"`
	// 現在時刻から最大4時間ぶんのタイムライン（各時間帯に在室予測メンバーと成立しそうな活動を含む）
	Hours []BoardHour `json:"hours"`
}

// boardPersonAssign はユーザーの時間帯割当に必要な情報を保持する
type boardPersonAssign struct {
	user         model.User
	visitMin     int // -1 = 予測なし
	departureMin int // -1 = 予測なし
}

// GetBoardData は共有モニター用の表示データを集約して返す
func GetBoardData() (BoardData, error) {
	log.Println("[board] GetBoardData: start")

	now := lib.NowJST()
	weekday := now.Weekday()
	log.Printf("[board] now=%s weekday=%s", now.Format("2006-01-02 15:04:05"), weekday)

	assigns := collectBoardPeople(weekday)
	log.Printf("[board] collectBoardPeople: assigns=%d", len(assigns))

	var e model.Event
	events, err := e.ReadAllWithUsers()
	if err != nil {
		log.Printf("[board] ReadAllWithUsers error: %v", err)
		return BoardData{}, err
	}
	log.Printf("[board] ReadAllWithUsers: events=%d", len(events))

	activityProbByEventID, err := activityProbabilitiesByEventID(weekday)
	if err != nil {
		log.Printf("[board] activityProbabilitiesByEventID error: %v", err)
		return BoardData{}, err
	}
	log.Printf("[board] activityProbabilitiesByEventID: probabilities=%d", len(activityProbByEventID))

	hours := buildBoardHours(events, assigns, activityProbByEventID, now.Hour())
	log.Printf("[board] buildBoardHours: hours=%d", len(hours))

	log.Println("[board] GetBoardData: done")
	return BoardData{
		CurrentTime: now.Format("15:04"),
		Presence:    BoardPresence{Members: []BoardPresentMember{}},
		Hours:       hours,
	}, nil
}

// activityProbabilitiesByEventID は全活動のGMM時間帯確率を EventID をキーにしたマップにして返す
func activityProbabilitiesByEventID(weekday time.Weekday) (map[uint]ActivityProbability, error) {
	probs, err := GetAllActivityProbabilities(weekday)
	if err != nil {
		return nil, err
	}

	byEventID := make(map[uint]ActivityProbability, len(probs))
	for _, p := range probs {
		byEventID[p.EventID] = p
	}
	return byEventID, nil
}

// boardHourRange は表示する時刻の列を返す
// 現在時刻を先頭に boardHourCount 時間ぶん並べる。boardHourStart より前は boardHourStart 始まりに、
// boardHourEnd を超える列は出さない（夜間は列が減っていき、最終的に空になる）
func boardHourRange(nowHour int) []int {
	start := nowHour
	if start < boardHourStart {
		start = boardHourStart
	}

	hours := make([]int, 0, boardHourCount)
	for hour := start; hour < start+boardHourCount && hour <= boardHourEnd; hour++ {
		hours = append(hours, hour)
	}
	return hours
}

// buildBoardHours はタイムラインの各列に在室予想者と成立しそうな活動を割り当てる
func buildBoardHours(events []model.Event, assigns []boardPersonAssign, activityProbByEventID map[uint]ActivityProbability, nowHour int) []BoardHour {
	hourColumns := boardHourRange(nowHour)

	hours := make([]BoardHour, 0, len(hourColumns))
	for _, hour := range hourColumns {
		people := []BoardPerson{}
		for _, a := range assigns {
			if isPresentAtHour(a, hour) {
				people = append(people, newBoardPerson(a.user))
			}
		}
		activities := buildBoardActivitiesForHour(events, assigns, activityProbByEventID, hour)
		log.Printf("[board] buildBoardHours: hour=%d people=%d activities=%d", hour, len(people), len(activities))
		hours = append(hours, BoardHour{
			Hour:       hour,
			People:     people,
			Activities: activities,
		})
	}
	return hours
}

// isPresentAtHour は予測時刻に基づき、指定の1時間にユーザーが在室していそうかを判定する
// - visit/departure 両方あり: 滞在区間とその時間が重なれば表示
// - visit のみ: visit 以降のすべての時間に表示
// - departure のみ: departure までのすべての時間に表示
// - 両方なし: 非表示
func isPresentAtHour(a boardPersonAssign, hour int) bool {
	startMin := hour * 60
	endMin := (hour + 1) * 60

	switch {
	case a.visitMin >= 0 && a.departureMin >= 0:
		return a.visitMin < endMin && a.departureMin > startMin
	case a.visitMin >= 0:
		return a.visitMin < endMin
	case a.departureMin >= 0:
		return a.departureMin > startMin
	default:
		return false
	}
}

// buildBoardActivitiesForHour は「関心のあるメンバーがその時間帯に在室していそうで、
// かつ最低人数以上そろいそうな活動」を、その時間のGMM活動確率がしきい値以上のものに絞って返す
func buildBoardActivitiesForHour(events []model.Event, assigns []boardPersonAssign, activityProbByEventID map[uint]ActivityProbability, hour int) []BoardActivity {
	// この時間帯に在室していそうな人（ArrivalMaybe 以上で足切り済み）
	presentUsers := make(map[uint]model.User, len(assigns))
	for _, a := range assigns {
		if isPresentAtHour(a, hour) {
			presentUsers[a.user.ID] = a.user
		}
	}

	activities := []BoardActivity{}
	for _, ev := range events {
		prob, ok := activityProbByEventID[ev.ID]
		if !ok || prob.Probabilities[hour] < config.Board.ActivityProbability {
			log.Printf("[board] buildBoardActivitiesForHour: hour=%d event=%q(id=%d) skipped by probability gate (prob=%.3f threshold=%.3f)",
				hour, ev.Name, ev.ID, prob.Probabilities[hour], config.Board.ActivityProbability)
			continue
		}

		members := []BoardPerson{}
		for _, eu := range ev.EventUsers {
			user, ok := presentUsers[eu.UserID]
			if !ok {
				continue
			}
			members = append(members, newBoardPerson(user))
		}

		if len(members) < ev.MinNumber {
			log.Printf("[board] buildBoardActivitiesForHour: hour=%d event=%q(id=%d) skipped by headcount gate (members=%d minNumber=%d)",
				hour, ev.Name, ev.ID, len(members), ev.MinNumber)
			continue
		}

		log.Printf("[board] buildBoardActivitiesForHour: hour=%d event=%q(id=%d) established (prob=%.3f members=%d)",
			hour, ev.Name, ev.ID, prob.Probabilities[hour], len(members))

		activities = append(activities, BoardActivity{
			ID:        ev.ID,
			Name:      ev.Name,
			ImageURL:  EventImageURL(ev.ImageKey),
			MinNumber: ev.MinNumber,
			Members:   members,
		})
	}
	return activities
}

// newBoardPerson は表示用の人物データを作る
func newBoardPerson(user model.User) BoardPerson {
	return BoardPerson{
		Name:      user.Name,
		AvatarURL: user.IconURL,
	}
}

// collectBoardPeople は全ユーザーの来訪確率・予測時刻を取得し、時間帯割当用の情報を作る
func collectBoardPeople(weekday time.Weekday) []boardPersonAssign {
	var u model.User
	users, err := u.ReadAll()
	if err != nil || len(users) == 0 {
		log.Printf("[board] collectBoardPeople: no users (err=%v)", err)
		return nil
	}
	log.Printf("[board] collectBoardPeople: users=%d", len(users))

	probs := GetStayWatchProbability(users, weekday)
	log.Printf("[board] collectBoardPeople: GetStayWatchProbability results=%d", len(probs))
	for _, p := range probs {
		log.Printf("[board] collectBoardPeople: userID=%d name=%s probability=%.3f", p.UserID, p.UserName, p.Probability)
	}

	// 来訪確率が ArrivalMaybe 以上のユーザーのみ対象

	// 来訪確率が maybe 閾値以上のユーザーのみ対象
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
		candidates = append(candidates, user)
	}
	log.Printf("[board] collectBoardPeople: candidates(prob>=%.2f)=%d", config.Board.ArrivalMaybe, len(candidates))
	if len(candidates) == 0 {
		return nil
	}

	visitTimes := fetchPredictionTime(candidates, weekday, "visit")
	departureTimes := fetchPredictionTime(candidates, weekday, "departure")
	log.Printf("[board] collectBoardPeople: visitTimes=%d departureTimes=%d", len(visitTimes), len(departureTimes))

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
