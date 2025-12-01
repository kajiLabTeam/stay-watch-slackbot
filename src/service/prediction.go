package service

import (
	"fmt"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
	"github.com/kajiLabTeam/stay-watch-slackbot/prediction"
	"gorm.io/gorm"
)

// ActivityPrediction represents prediction results for an activity
type ActivityPrediction struct {
	EventID          uint      `json:"event_id"`
	EventName        string    `json:"event_name"`
	DayOfWeek        string    `json:"day_of_week"`
	Probability      float64   `json:"probability"`
	MostLikelyTime   string    `json:"most_likely_time"`
	ClusterCount     int       `json:"cluster_count"`
	DataPointCount   int       `json:"data_point_count"`
}

// GetActivityPrediction calculates activity prediction for a specific event and day of week
func GetActivityPrediction(eventID uint, dayOfWeek time.Weekday, targetTime string, weeks int) (*ActivityPrediction, error) {
	// Get event details
	event := model.Event{Model: gorm.Model{ID: eventID}}
	if err := event.ReadByID(); err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Get logs for the specific event and day of week
	logs, err := model.ReadLogsByEventIDAndDayOfWeek(eventID, dayOfWeek, weeks)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}

	if len(logs) == 0 {
		return &ActivityPrediction{
			EventID:        eventID,
			EventName:      event.Name,
			DayOfWeek:      dayOfWeek.String(),
			Probability:    0,
			MostLikelyTime: "",
			ClusterCount:   0,
			DataPointCount: 0,
		}, nil
	}

	// Filter logs by status "start" and extract times
	var startTimes []string
	for _, log := range logs {
		// Load status if not preloaded
		if log.Status.Name == "" {
			status := model.Status{Model: gorm.Model{ID: log.StatusID}}
			if err := status.ReadByID(); err != nil {
				continue
			}
			log.Status = status
		}

		if log.Status.Name == "start" {
			// Format time as HH:MM
			timeStr := log.CreatedAt.Format("15:04")
			startTimes = append(startTimes, timeStr)
		}
	}

	if len(startTimes) == 0 {
		return &ActivityPrediction{
			EventID:        eventID,
			EventName:      event.Name,
			DayOfWeek:      dayOfWeek.String(),
			Probability:    0,
			MostLikelyTime: "",
			ClusterCount:   0,
			DataPointCount: 0,
		}, nil
	}

	// Calculate probability using prediction package
	probability, err := prediction.GetProbability(startTimes, targetTime, weeks)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate probability: %w", err)
	}

	// Get most likely time
	mostLikelyMinutes, err := prediction.GetMostLikelyTime(startTimes, weeks)
	if err != nil {
		return nil, fmt.Errorf("failed to get most likely time: %w", err)
	}
	mostLikelyTime := prediction.MinutesToTime(mostLikelyMinutes)

	// Get cluster count
	dataMinutes := make([]int, len(startTimes))
	for i, t := range startTimes {
		minutes, _ := prediction.TimeToMinutes(t)
		dataMinutes[i] = minutes
	}
	clusters := prediction.Clustering(dataMinutes)

	return &ActivityPrediction{
		EventID:        eventID,
		EventName:      event.Name,
		DayOfWeek:      dayOfWeek.String(),
		Probability:    probability,
		MostLikelyTime: mostLikelyTime,
		ClusterCount:   len(clusters),
		DataPointCount: len(startTimes),
	}, nil
}

// GetWeeklyActivityPredictions gets predictions for all days of the week
func GetWeeklyActivityPredictions(eventID uint, targetTime string, weeks int) (map[string]*ActivityPrediction, error) {
	predictions := make(map[string]*ActivityPrediction)

	daysOfWeek := []time.Weekday{
		time.Sunday,
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
	}

	for _, day := range daysOfWeek {
		pred, err := GetActivityPrediction(eventID, day, targetTime, weeks)
		if err != nil {
			return nil, fmt.Errorf("failed to get prediction for %s: %w", day.String(), err)
		}
		predictions[day.String()] = pred
	}

	return predictions, nil
}

// GetAllEventsPredictions gets predictions for all events
func GetAllEventsPredictions(dayOfWeek time.Weekday, targetTime string, weeks int) ([]*ActivityPrediction, error) {
	event := model.Event{}
	events, err := event.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	var predictions []*ActivityPrediction
	for _, e := range events {
		pred, err := GetActivityPrediction(e.ID, dayOfWeek, targetTime, weeks)
		if err != nil {
			// Log error but continue with other events
			continue
		}
		predictions = append(predictions, pred)
	}

	return predictions, nil
}
