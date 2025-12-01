package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
)

// GetActivityPredictionHandler handles GET /api/v1/prediction/activity/:event_id
func GetActivityPredictionHandler(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	dayOfWeekStr := c.DefaultQuery("day_of_week", time.Now().Weekday().String())
	targetTime := c.DefaultQuery("target_time", time.Now().Format("15:04"))
	weeksStr := c.DefaultQuery("weeks", "4")

	weeks, err := strconv.Atoi(weeksStr)
	if err != nil || weeks < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "weeks must be a positive integer"})
		return
	}

	// Parse day of week
	dayOfWeek := parseDayOfWeek(dayOfWeekStr)

	prediction, err := service.GetActivityPrediction(uint(eventID), dayOfWeek, targetTime, weeks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prediction)
}

// GetWeeklyActivityPredictionsHandler handles GET /api/v1/prediction/activity/:event_id/weekly
func GetWeeklyActivityPredictionsHandler(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	targetTime := c.DefaultQuery("target_time", time.Now().Format("15:04"))
	weeksStr := c.DefaultQuery("weeks", "4")

	weeks, err := strconv.Atoi(weeksStr)
	if err != nil || weeks < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "weeks must be a positive integer"})
		return
	}

	predictions, err := service.GetWeeklyActivityPredictions(uint(eventID), targetTime, weeks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, predictions)
}

// GetAllEventsPredictionsHandler handles GET /api/v1/prediction/events
func GetAllEventsPredictionsHandler(c *gin.Context) {
	dayOfWeekStr := c.DefaultQuery("day_of_week", time.Now().Weekday().String())
	targetTime := c.DefaultQuery("target_time", time.Now().Format("15:04"))
	weeksStr := c.DefaultQuery("weeks", "4")

	weeks, err := strconv.Atoi(weeksStr)
	if err != nil || weeks < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "weeks must be a positive integer"})
		return
	}

	dayOfWeek := parseDayOfWeek(dayOfWeekStr)

	predictions, err := service.GetAllEventsPredictions(dayOfWeek, targetTime, weeks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, predictions)
}

// parseDayOfWeek converts string to time.Weekday
func parseDayOfWeek(dayStr string) time.Weekday {
	switch dayStr {
	case "Sunday", "sunday", "Sun", "sun", "0":
		return time.Sunday
	case "Monday", "monday", "Mon", "mon", "1":
		return time.Monday
	case "Tuesday", "tuesday", "Tue", "tue", "2":
		return time.Tuesday
	case "Wednesday", "wednesday", "Wed", "wed", "3":
		return time.Wednesday
	case "Thursday", "thursday", "Thu", "thu", "4":
		return time.Thursday
	case "Friday", "friday", "Fri", "fri", "5":
		return time.Friday
	case "Saturday", "saturday", "Sat", "sat", "6":
		return time.Saturday
	default:
		return time.Now().Weekday()
	}
}
