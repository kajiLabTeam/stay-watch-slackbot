package service

import (
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
}

// BoardActivity は今日成立しそうな活動1件を表す
type BoardActivity struct {
	// イベントID
	ID uint `json:"id" example:"5"`
	// 活動（イベント）名
	Name string `json:"name" example:"人狼"`
	// 活動の画像URL。未登録の場合は null
	ImageURL *string `json:"imageUrl" example:"https://example.com/daycast/events/5.png"`
	// 活動の成立に必要な最低人数
	MinNumber int `json:"minNumber" example:"3"`
	// この活動に関心があり、かつ今日来訪しそうなメンバー
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
	// 現在時刻から最大4時間ぶんのタイムライン
	Hours []BoardHour `json:"hours"`
	// 今日メンバーが揃いそうな活動一覧
	Activities []BoardActivity `json:"activities"`
}

// boardPersonAssign はユーザーの時間帯割当に必要な情報を保持する
type boardPersonAssign struct {
	user         model.User
	visitMin     int // -1 = 予測なし
	departureMin int // -1 = 予測なし
}

// GetBoardData は共有モニター用の表示データを集約して返す
func GetBoardData() (BoardData, error) {
	now := lib.NowJST()
	weekday := now.Weekday()

	assigns := collectBoardPeople(weekday)

	var e model.Event
	events, err := e.ReadAllWithUsers()
	if err != nil {
		return BoardData{}, err
	}

	return BoardData{
		CurrentTime: now.Format("15:04"),
		Presence:    BoardPresence{Members: []BoardPresentMember{}},
		Hours:       buildBoardHours(assigns, now.Hour()),
		Activities:  buildBoardActivities(events, assigns),
	}, nil
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

// buildBoardHours はタイムラインの各列に在室予想者を割り当てる
func buildBoardHours(assigns []boardPersonAssign, nowHour int) []BoardHour {
	hourColumns := boardHourRange(nowHour)

	hours := make([]BoardHour, 0, len(hourColumns))
	for _, hour := range hourColumns {
		people := []BoardPerson{}
		for _, a := range assigns {
			if isPresentAtHour(a, hour) {
				people = append(people, newBoardPerson(a.user))
			}
		}
		hours = append(hours, BoardHour{Hour: hour, People: people})
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

// buildBoardActivities は「関心のあるメンバーが最低人数以上そろいそうな活動」を返す
// 活動の発生確率（GMM由来）は選別に使わず、人が揃うかどうかだけで判断する
func buildBoardActivities(events []model.Event, assigns []boardPersonAssign) []BoardActivity {
	// 今日来訪しそうな人（ArrivalMaybe 以上で足切り済み）
	comingUsers := make(map[uint]model.User, len(assigns))
	for _, a := range assigns {
		comingUsers[a.user.ID] = a.user
	}

	activities := []BoardActivity{}
	for _, ev := range events {
		members := []BoardPerson{}
		for _, eu := range ev.EventUsers {
			user, ok := comingUsers[eu.UserID]
			if !ok {
				continue
			}
			members = append(members, newBoardPerson(user))
		}

		if len(members) < ev.MinNumber {
			continue
		}

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
		return nil
	}

	probs := GetStayWatchProbability(users, weekday)

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
