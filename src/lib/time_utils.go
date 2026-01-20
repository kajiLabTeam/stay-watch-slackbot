// Package lib provides common utility functions used across the application.
package lib

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// JST は日本標準時（表示用、固定）
var JST = time.FixedZone("JST", 9*60*60)

// StayWatchTimezone はStayWatch APIが使用するタイムゾーン
// TODO: StayWatchがUTCに移行したら time.UTC に変更する
var StayWatchTimezone = time.FixedZone("JST", 9*60*60)

// ToJST はUTC時刻をJSTに変換する（Slack表示用）
func ToJST(t time.Time) time.Time {
	return t.In(JST)
}

// FormatTimeJST はUTC時刻をJSTの"HH:MM"形式に変換する（Slack表示用）
func FormatTimeJST(t time.Time) string {
	return ToJST(t).Format("15:04")
}

// ToStayWatchTime はUTC時刻をStayWatchのタイムゾーンに変換する
func ToStayWatchTime(t time.Time) time.Time {
	return t.In(StayWatchTimezone)
}

// FormatForStayWatch はUTC時刻をStayWatch用の"HH:MM"形式に変換する
func FormatForStayWatch(t time.Time) string {
	return ToStayWatchTime(t).Format("15:04")
}

// FormatDateTimeForStayWatch はUTC時刻をStayWatch用の"2006-01-02 15:04"形式に変換する
func FormatDateTimeForStayWatch(t time.Time) string {
	return ToStayWatchTime(t).Format("2006-01-02 15:04")
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

// ErrUnsupportedTimezone は対応していないタイムゾーンの場合のエラー
var ErrUnsupportedTimezone = fmt.Errorf("timezone must be JST (+09:00) or UTC (+00:00/Z)")

// ErrNoTimezone は入力時刻にタイムゾーン情報がない場合のエラー
var ErrNoTimezone = fmt.Errorf("timezone information is required (use RFC3339 format, e.g., 2006-01-02T15:04:05+09:00)")

// TimezoneType は検出されたタイムゾーンの種類を表す
type TimezoneType string

const (
	TimezoneJST TimezoneType = "JST"
	TimezoneUTC TimezoneType = "UTC"
)

// ParseResult はパース結果を表す
type ParseResult struct {
	Time     time.Time    // UTC変換後の時刻
	Timezone TimezoneType // 検出されたタイムゾーン
}

// ParseWithTimezoneDetection はRFC3339形式の時刻文字列をパースし、タイムゾーンを判別してUTCに変換する
// 入力形式: "2006-01-02T15:04:05+09:00" (JST) または "2006-01-02T15:04:05Z" / "2006-01-02T15:04:05+00:00" (UTC)
// JSTまたはUTC以外のタイムゾーン、またはタイムゾーン情報がない場合はエラーを返す
func ParseWithTimezoneDetection(timeStr string) (ParseResult, error) {
	// RFC3339形式でパース
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return ParseResult{}, ErrNoTimezone
	}

	// タイムゾーンオフセットを確認
	_, offset := t.Zone()

	switch offset {
	case 9 * 60 * 60: // JST (+09:00)
		return ParseResult{
			Time:     t.UTC(),
			Timezone: TimezoneJST,
		}, nil
	case 0: // UTC (+00:00 or Z)
		return ParseResult{
			Time:     t.UTC(),
			Timezone: TimezoneUTC,
		}, nil
	default:
		return ParseResult{}, ErrUnsupportedTimezone
	}
}
