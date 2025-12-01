package model

import "time"

func (l *Log) Create() error {
	if err := db.Create(l).Error; err != nil {
		return err
	}
	return nil
}

func (l *Log) ReadByID() error {
	if err := db.Preload("Correspondence").Preload("Status").First(l, l.ID).Error; err != nil {
		return err
	}
	return nil
}

func (l *Log) ReadAll() ([]Log, error) {
	var logs []Log
	if err := db.Preload("Correspondence").Preload("Status").Find(&logs).Error; err != nil {
		return logs, err
	}
	return logs, nil
}

// ReadByCorrespondenceID retrieves logs by correspondence ID
func (l *Log) ReadByCorrespondenceID() ([]Log, error) {
	var logs []Log
	if err := db.Preload("Status").Where("correspondence_id = ?", l.CorrespondenceID).Find(&logs).Error; err != nil {
		return logs, err
	}
	return logs, nil
}

// ReadByEventIDAndDateRange retrieves logs by event ID and date range
func ReadLogsByEventIDAndDateRange(eventID uint, startDate, endDate time.Time) ([]Log, error) {
	var logs []Log
	if err := db.Joins("JOIN correspondences ON correspondences.id = logs.correspondence_id").
		Where("correspondences.event_id = ? AND logs.created_at BETWEEN ? AND ?", eventID, startDate, endDate).
		Preload("Correspondence").
		Preload("Status").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// ReadByEventIDAndDayOfWeek retrieves logs by event ID and day of week
func ReadLogsByEventIDAndDayOfWeek(eventID uint, dayOfWeek time.Weekday, weeks int) ([]Log, error) {
	var logs []Log
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7*weeks)

	if err := db.Joins("JOIN correspondences ON correspondences.id = logs.correspondence_id").
		Where("correspondences.event_id = ? AND logs.created_at BETWEEN ? AND ?", eventID, startDate, endDate).
		Preload("Correspondence").
		Preload("Status").
		Find(&logs).Error; err != nil {
		return nil, err
	}

	// Filter by day of week
	filteredLogs := make([]Log, 0)
	for _, log := range logs {
		if log.CreatedAt.Weekday() == dayOfWeek {
			filteredLogs = append(filteredLogs, log)
		}
	}

	return filteredLogs, nil
}
