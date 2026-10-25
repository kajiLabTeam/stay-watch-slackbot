package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

// GroupingAxis はグループ化に使う軸を表す
type GroupingAxis string

const (
	// GroupByMinNumber は最低必要人数でグループ化する軸
	GroupByMinNumber GroupingAxis = "min_number"
	// GroupByEquipment は使用する機材の組み合わせでグループ化する軸
	GroupByEquipment GroupingAxis = "equipment"
)

// ActivityGroup は指定した軸が一致する活動をまとめたグループを表す
type ActivityGroup struct {
	Key        string
	MinNumber  *int     // GroupByMinNumber軸を使った場合のみセット
	Equipment  []string // GroupByEquipment軸を使った場合のみセット（名前でソート済み）
	EventIDs   []uint
	EventNames []string
}

// ActivityGroupProbability はグループの1時間ごとの発生確率を表す
type ActivityGroupProbability struct {
	GroupKey      string    `json:"groupKey"`
	EventIDs      []uint    `json:"eventIds"`
	EventNames    []string  `json:"eventNames"`
	MinNumber     *int      `json:"minNumber,omitempty"`
	Equipment     []string  `json:"equipment,omitempty"`
	Probabilities []float64 `json:"probabilities"` // length 24, index = hour (0-23 JST), value = 0.0〜1.0
}

// GroupEvents は指定した軸（複数可）が一致する活動を1グループにまとめる。
// 出現順を保持するため、同じ入力に対して常に同じ順序の結果を返す
func GroupEvents(events []model.Event, axes []GroupingAxis) []ActivityGroup {
	byKey := make(map[string]*ActivityGroup)
	var order []string

	for _, ev := range events {
		key, minNumber, equipmentNames := buildGroupKey(ev, axes)

		g, ok := byKey[key]
		if !ok {
			g = &ActivityGroup{Key: key, MinNumber: minNumber, Equipment: equipmentNames}
			byKey[key] = g
			order = append(order, key)
		}
		g.EventIDs = append(g.EventIDs, ev.ID)
		g.EventNames = append(g.EventNames, ev.Name)
	}

	groups := make([]ActivityGroup, 0, len(order))
	for _, key := range order {
		groups = append(groups, *byKey[key])
	}
	return groups
}

// buildGroupKey は1活動分のグループキーと、軸ごとの付帯情報を作る
func buildGroupKey(ev model.Event, axes []GroupingAxis) (key string, minNumber *int, equipmentNames []string) {
	var parts []string
	for _, axis := range axes {
		switch axis {
		case GroupByMinNumber:
			n := ev.MinNumber
			minNumber = &n
			parts = append(parts, fmt.Sprintf("min:%d", n))
		case GroupByEquipment:
			names := make([]string, len(ev.Equipment))
			for i, eq := range ev.Equipment {
				names[i] = eq.Name
			}
			sort.Strings(names)
			equipmentNames = names
			parts = append(parts, "eq:"+strings.Join(names, ","))
		}
	}
	return strings.Join(parts, "|"), minNumber, equipmentNames
}

// GetAllActivityGroupProbabilities は指定した軸でグループ化した活動の時間帯別発生確率を取得する
func GetAllActivityGroupProbabilities(dayOfWeek time.Weekday, axes []GroupingAxis) ([]ActivityGroupProbability, error) {
	event := model.Event{}
	events, err := event.ReadAllWithEquipment()
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	groups := GroupEvents(events, axes)

	results := make([]ActivityGroupProbability, 0, len(groups))
	for _, g := range groups {
		results = append(results, calcGroupProbability(g, dayOfWeek))
	}
	return results, nil
}

// calcGroupProbability はグループ1件分の活動確率を計算する（calcEventProbabilityのグループ版）
func calcGroupProbability(g ActivityGroup, dayOfWeek time.Weekday) ActivityGroupProbability {
	result := ActivityGroupProbability{
		GroupKey:      g.Key,
		EventIDs:      g.EventIDs,
		EventNames:    g.EventNames,
		MinNumber:     g.MinNumber,
		Equipment:     g.Equipment,
		Probabilities: make([]float64, 24),
	}

	logs, err := model.ReadLogsByEventIDsAndDayOfWeek(g.EventIDs, dayOfWeek)
	if err != nil || len(logs) == 0 {
		return result
	}

	weeks := calculateWeeks(logs)
	datetimeStrings := extractStartDatetimes(logs)
	if len(datetimeStrings) == 0 {
		return result
	}

	result.Probabilities = calcHourlyProbabilities(datetimeStrings, weeks)
	return result
}
