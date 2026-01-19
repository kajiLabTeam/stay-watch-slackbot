// Package lib provides common utility functions used across the application.
package lib

import (
	"fmt"
	"strconv"
	"strings"
)

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

// ExtractTimeFromDatetime "YYYY-MM-DD HH:MM"形式の日時文字列から時刻部分を抽出する
func ExtractTimeFromDatetime(datetimeStr string) (string, error) {
	parts := strings.SplitN(datetimeStr, " ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid datetime format: %s", datetimeStr)
	}
	return parts[1], nil
}
