package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func GetStayWatchMember() ([]StaywatchUsers, error) {
	var users []StaywatchUsers
	req, err := http.NewRequest("GET", staywatch.BaseURL+staywatch.Users, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", staywatch.ApiKey)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func GetStayWatchProbability(users []model.User) []Probability {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	timeStr := "23:59"
	w := int(now.Weekday())
	u, _ := url.Parse(staywatch.BaseURL + staywatch.Probability + "/visit")
	q := u.Query()
	for _, user := range users {
		q.Add("user-id", strconv.Itoa(int(user.StayWatchID)))
	}
	q.Add("weekday", strconv.Itoa(w))
	q.Add("time", timeStr)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return []Probability{}
	}
	req.Header.Set("X-API-Key", staywatch.ApiKey)
	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return []Probability{}
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	var r StayWatchResponse
	if err := json.Unmarshal(b, &r); err != nil {
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

func fetchPredictionTime(users []model.User, action string) []Result {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	w := int(now.Weekday())
	u, _ := url.Parse(staywatch.BaseURL + staywatch.Time + "/" + action)
	q := u.Query()
	for _, user := range users {
		q.Add("user-id", strconv.Itoa(int(user.StayWatchID)))
	}
	q.Add("weekday", strconv.Itoa(w))
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return []Result{}
	}
	req.Header.Set("X-API-Key", staywatch.ApiKey)
	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return []Result{}
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	var r StayWatchResponse
	if err := json.Unmarshal(b, &r); err != nil {
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

func NotifyByTag() ([]model.User, map[int][]string) {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	weekdays := [...]string{"日", "月", "火", "水", "木", "金", "土"}
	formatted := fmt.Sprintf("%d/%d(%s)", now.Month(), now.Day(), weekdays[now.Weekday()])
	var u model.User
	users, _ := u.ReadAll()

	// Step 1: 来訪確率取得 & フィルタ
	probs := GetStayWatchProbability(users)
	// log.Default().Println("probability", probs)
	filtered := filterByThreshold(probs, 0.05)

	// Step 2: 来訪・退室時刻の予測を取得し、マージ
	visitTimes := fetchPredictionTime(filtered, "visit")
	departureTimes := fetchPredictionTime(filtered, "departure")
	predictions := mergePredictions(visitTimes, departureTimes)

	// Step 3: タグごとにユーザをグループ化
	tagGroups, _ := model.GroupByTag(filtered)
	tagGroupWithMSG := make(map[int][]string)
	// Step 4: タグごとに滞在時間重なりを検出
	for _, group := range tagGroups {
		// log.Default().Println("tag", group.Tag.Name)
		// // log.Default().Println("users", group.Users)
		// for _, u := range group.Users {
		// 	log.Default().Println("user", u.Name)
		// }
		ranges := findOverlappingRanges(predictions, group.Users, group.Tag.MinNumber)
		// log.Default().Println("ranges", ranges)
		if len(ranges) == 0 {
			continue
		}
		for _, r := range ranges {
			msg := fmt.Sprintf("%s%s〜%s に `%s` の仲間が集まりそうです", formatted, r.Start, r.End, group.Tag.Name)
			// Slack送信（またはログ出力等）
			tagGroupWithMSG[int(group.Tag.ID)] = append(tagGroupWithMSG[int(group.Tag.ID)], msg)
			// 例: slack.SendMessageToTag(tagName, msg)
		}
	}
	// log.Default().Println("tagGroupWithMSG", tagGroupWithMSG)
	return users, tagGroupWithMSG
}
