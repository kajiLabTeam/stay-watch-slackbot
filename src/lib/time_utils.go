// Package lib provides common utility functions used across the application.
package lib

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// JST は日本標準時
var JST = time.FixedZone("JST", 9*60*60)

// FormatTime は時刻を"HH:MM"形式にフォーマットする
func FormatTime(t time.Time) string {
	return t.In(JST).Format("15:04")
}

// FormatDateTime は時刻を"2006-01-02 15:04"形式にフォーマットする
func FormatDateTime(t time.Time) string {
	return t.In(JST).Format("2006-01-02 15:04")
}

// NowJST は現在時刻をJSTで返す
func NowJST() time.Time {
	return time.Now().In(JST)
}

// TimeToMinutes "HH:MM"形式の時刻を分に変換する
func TimeToMinutes(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours: %s", parts[0])
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %s", parts[1])
	}

	return hours*60 + minutes, nil
}

// MinutesToTime 分を"HH:MM"形式に変換する
func MinutesToTime(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}

// ParseJST はRFC3339形式の時刻文字列をパースしてJSTに変換する
// 入力形式: "2006-01-02T15:04:05+09:00" (JST)
// JST以外のタイムゾーンはエラーを返す
func ParseJST(timeStr string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("RFC3339形式で入力してください (例: 2006-01-02T15:04:05+09:00): %w", err)
	}

	_, offset := t.Zone()
	if offset != 9*60*60 {
		return time.Time{}, fmt.Errorf("タイムゾーンはJST (+09:00) のみ対応しています")
	}

	return t.In(JST), nil
}
