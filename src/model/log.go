package model

import (
	"time"
)

func (l *Log) Create() error {
	if err := db.Create(l).Error; err != nil {
		return err
	}
	return nil
}

func (l *Log) ReadByID() error {
	if err := db.Preload("Event").Preload("Status").First(l, l.ID).Error; err != nil {
		return err
	}
	return nil
}

func (l *Log) ReadAll() ([]Log, error) {
	var logs []Log
	if err := db.Preload("Event").Preload("Status").Find(&logs).Error; err != nil {
		return logs, err
	}
	return logs, nil
}

// ReadByEventID retrieves logs by event ID
func (l *Log) ReadByEventID() ([]Log, error) {
	var logs []Log
	if err := db.Preload("Status").Where("event_id = ?", l.EventID).Find(&logs).Error; err != nil {
		return logs, err
	}
	return logs, nil
}

// ReadLogsByEventIDAndDateRange retrieves logs by event ID and date range
func ReadLogsByEventIDAndDateRange(eventID uint, startDate, endDate time.Time) ([]Log, error) {
	var logs []Log
	if err := db.Where("event_id = ? AND logs.created_at BETWEEN ? AND ?", eventID, startDate, endDate).
		Preload("Event").
		Preload("Status").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// ReadLogsByEventIDAndDayOfWeek retrieves logs by event ID and day of week
func ReadLogsByEventIDAndDayOfWeek(eventID uint, dayOfWeek time.Weekday) ([]Log, error) {
	var logs []Log

	// dayOfWeekをMysqlのWEEKDAY関数に合わせて変換 (0=月曜日, ..., 6=日曜日)
	mysqlDayOfWeek := (int(dayOfWeek) + 6) % 7

	// 指定した曜日のログを全て取得
	if err := db.Where("event_id = ? AND WEEKDAY(logs.created_at) = ?", eventID, mysqlDayOfWeek).
		Preload("Event").
		Preload("Status").
		Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// ReadLogsByEventIDAndDayOfWeekWithWeeks retrieves logs by event ID, day of week, and number of weeks
func ReadLogsByEventIDAndDayOfWeekWithWeeks(eventID uint, dayOfWeek time.Weekday, weeks int) ([]Log, error) {
	var logs []Log

	// dayOfWeekをMysqlのWEEKDAY関数に合わせて変換 (0=月曜日, ..., 6=日曜日)
	mysqlDayOfWeek := (int(dayOfWeek) + 6) % 7

	// 指定した曜日、指定した週数のログを全て取得
	if err := db.Where("event_id = ? AND WEEKDAY(logs.created_at) = ?", eventID, mysqlDayOfWeek).
		Preload("Event").
		Preload("Status").
		Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}
