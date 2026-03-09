package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
)

const (
	msgInvalidRequestBody         = "invalid request body"
	msgBatchRegistrationCompleted = "batch registration completed"
)

// RegisterTypesRequest はType一括登録のリクエストボディ
type RegisterTypesRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
}

// RegisterToolsRequest はTool一括登録のリクエストボディ
type RegisterToolsRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
}

// RegisterStatusesRequest はStatus一括登録のリクエストボディ
type RegisterStatusesRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
}

// LogEntry はログ登録リクエストの1エントリを表す
type LogEntry struct {
	EventID   uint   `json:"event_id" binding:"required"`
	StatusID  uint   `json:"status_id" binding:"required"`
	CreatedAt string `json:"created_at" binding:"required"` // RFC3339形式 JST or UTC (例: "2006-01-02T15:04:05+09:00" or "2006-01-02T15:04:05Z")
}

// RegisterLogsRequest はログ一括登録のリクエストボディ
type RegisterLogsRequest struct {
	Logs []LogEntry `json:"logs" binding:"required,min=1"`
}

// GetTypes はType一覧を取得するAPIハンドラー
// @Summary Type一覧を取得
// @Tags types
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/types [get]
func GetTypes(c *gin.Context) {
	types, err := service.GetTypes()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": types,
	})
}

// GetTools はTool一覧を取得するAPIハンドラー
// @Summary Tool一覧を取得
// @Tags tools
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/tools [get]
func GetTools(c *gin.Context) {
	tools, err := service.GetTools()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tools,
	})
}

// GetStatuses はStatus一覧を取得するAPIハンドラー
// @Summary Status一覧を取得
// @Tags statuses
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/statuses [get]
func GetStatuses(c *gin.Context) {
	statuses, err := service.GetStatuses()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": statuses,
	})
}

// PostRegisterTypes はTypeを一括登録するAPIハンドラー
// @Summary Typeを一括登録
// @Tags types
// @Accept json
// @Produce json
// @Param request body RegisterTypesRequest true "登録するType名のリスト"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/types [post]
func PostRegisterTypes(c *gin.Context) {
	var req RegisterTypesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, msgInvalidRequestBody)
		return
	}

	types, errors, err := service.BatchRegisterTypes(req.Names)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": msgBatchRegistrationCompleted,
		"data":    types,
		"errors":  errors,
	})
}

// PostRegisterTools はToolを一括登録するAPIハンドラー
// @Summary Toolを一括登録
// @Tags tools
// @Accept json
// @Produce json
// @Param request body RegisterToolsRequest true "登録するTool名のリスト"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/tools [post]
func PostRegisterTools(c *gin.Context) {
	var req RegisterToolsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, msgInvalidRequestBody)
		return
	}

	tools, errors, err := service.BatchRegisterTools(req.Names)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": msgBatchRegistrationCompleted,
		"data":    tools,
		"errors":  errors,
	})
}

// PostRegisterStatuses はStatusを一括登録するAPIハンドラー
// @Summary Statusを一括登録
// @Tags statuses
// @Accept json
// @Produce json
// @Param request body RegisterStatusesRequest true "登録するStatus名のリスト"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/statuses [post]
func PostRegisterStatuses(c *gin.Context) {
	var req RegisterStatusesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, msgInvalidRequestBody)
		return
	}

	statuses, errors, err := service.BatchRegisterStatuses(req.Names)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": msgBatchRegistrationCompleted,
		"data":    statuses,
		"errors":  errors,
	})
}

// GetEventProbability は指定したイベントと曜日の発生確率を取得するAPIハンドラー
// @Summary イベントの発生確率を取得
// @Tags events
// @Produce json
// @Param id path int true "イベントID"
// @Param weekday query int true "曜日 (MySQL WEEKDAY形式: 0=月, 6=日)"
// @Param time query string false "時刻 (HH:MM形式, JST。デフォルト: 現在時刻)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/events/{id}/probability [get]
func GetEventProbability(c *gin.Context) {
	// パスパラメータからイベントIDを取得
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid event id")
		return
	}

	// クエリパラメータから曜日を取得（必須）
	weekdayStr := c.Query("weekday")
	if weekdayStr == "" {
		respondError(c, http.StatusBadRequest, "weekday parameter is required")
		return
	}
	weekdayInt, err := strconv.Atoi(weekdayStr)
	if err != nil || weekdayInt < 0 || weekdayInt > 6 {
		respondError(c, http.StatusBadRequest, "weekday must be 0-6 (Monday=0, Sunday=6)")
		return
	}
	// MySQL WEEKDAY形式(月=0)からGoのtime.Weekday形式(日=0)に変換
	weekday := time.Weekday((weekdayInt + 1) % 7)

	// クエリパラメータから時刻を取得（オプション、デフォルトは現在時刻JST）
	jst := time.FixedZone("JST", 9*60*60)
	inputTimeJST := c.DefaultQuery("time", time.Now().In(jst).Format("15:04"))

	// JST入力をUTCに変換
	parsedJST, err := time.ParseInLocation("15:04", inputTimeJST, jst)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid time format (expected HH:MM)")
		return
	}
	targetTimeUTC := parsedJST.UTC().Format("15:04")

	// 確率を取得
	probability, err := service.GetActivityProbability(uint(eventID), weekday, targetTimeUTC)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event_id":    eventID,
		"weekday":     weekdayInt,
		"time":        inputTimeJST,
		"probability": probability,
	})
}

// GetAllActivityProbabilities は全活動の1時間ごとの発生確率を取得するAPIハンドラー
// @Summary 全活動の時間帯別発生確率を取得
// @Tags activities
// @Produce json
// @Param weekday query int false "曜日 (MySQL WEEKDAY形式: 0=月, 6=日)。省略時は今日の曜日"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/activities/probabilities [get]
func GetAllActivityProbabilities(c *gin.Context) {
	jst := time.FixedZone("JST", 9*60*60)
	var weekday time.Weekday

	weekdayStr := c.Query("weekday")
	if weekdayStr == "" {
		// デフォルト: 今日の曜日（JST）
		weekday = time.Now().In(jst).Weekday()
	} else {
		weekdayInt, err := strconv.Atoi(weekdayStr)
		if err != nil || weekdayInt < 0 || weekdayInt > 6 {
			respondError(c, http.StatusBadRequest, "weekday must be 0-6 (Monday=0, Sunday=6)")
			return
		}
		// MySQL WEEKDAY形式(月=0)からGoのtime.Weekday形式(日=0)に変換
		weekday = time.Weekday((weekdayInt + 1) % 7)
	}

	results, err := service.GetAllActivityProbabilities(weekday)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}

// PostRegisterLogs はログを一括登録するAPIハンドラー
// @Summary ログを一括登録
// @Tags logs
// @Accept json
// @Produce json
// @Param request body RegisterLogsRequest true "登録するログエントリのリスト"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/logs [post]
func PostRegisterLogs(c *gin.Context) {
	var req RegisterLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, msgInvalidRequestBody)
		return
	}

	// リクエストをサービス層の入力形式に変換
	inputs := make([]service.LogEntryInput, len(req.Logs))
	for i, entry := range req.Logs {
		inputs[i] = service.LogEntryInput{
			EventID:   entry.EventID,
			StatusID:  entry.StatusID,
			CreatedAt: entry.CreatedAt,
		}
	}

	logs, errors, err := service.BatchRegisterLogs(inputs)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": msgBatchRegistrationCompleted,
		"data":    logs,
		"errors":  errors,
	})
}
