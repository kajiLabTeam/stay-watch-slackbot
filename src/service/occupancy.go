package service

import (
	"sort"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

// TimeRange 時間範囲
type TimeRange struct {
	Start string
	End   string
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
		overlapStart := lib.Max(activityStartMinutes, occupancyStartMinutes)
		overlapEnd := lib.Min(activityEndMinutes, occupancyEndMinutes)

		if overlapStart < overlapEnd {
			recommendedRanges = append(recommendedRanges, TimeRange{
				Start: lib.MinutesToTime(overlapStart),
				End:   lib.MinutesToTime(overlapEnd),
			})
		}
	}

	return recommendedRanges
}

// findOverlappingRanges 規定人数在室時間を計算する
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
