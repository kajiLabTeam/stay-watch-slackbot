package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
)

const (
	msgInvalidRequestBody         = "invalid request body"
	msgBatchRegistrationCompleted = "batch registration completed"
)

// RegisterStatusesRequest はStatus一括登録のリクエストボディ
type RegisterStatusesRequest struct {
	// 登録するStatus名の配列（1件以上、重複不可）
	Names []string `json:"names" binding:"required,min=1" example:"作業中,休憩中"`
}

// LogEntry はログ登録リクエストの1エントリを表す
type LogEntry struct {
	// 対象のEventのID（文字列で指定）
	EventID string `json:"event_id" binding:"required" example:"1"`
	// 対象のStatusのID
	StatusID uint `json:"status_id" binding:"required" example:"2"`
	// イベント発生日時。RFC3339形式のJST
	EventTime string `json:"event_time" binding:"required" example:"2006-01-02T15:04:05+09:00"`
	// 参加メンバのstay_watch_idリスト（空可）
	ParticipateUsers []int64 `json:"participate_users" example:"1,2"`
	// 在室メンバのstay_watch_idリスト（空可）
	RoomUsers []int64 `json:"room_users" example:"3,4"`
}

// RegisterLogsRequest はログ一括登録のリクエストボディ
type RegisterLogsRequest struct {
	// 登録するログエントリの配列（1件以上）
	Logs []LogEntry `json:"logs" binding:"required,min=1"`
}

// GetEvents はEvent一覧を取得するAPIハンドラー
// id を指定すると該当するEvent1件を、game_type（"digital" | "analog"）を指定すると
// その分類に属するEventの一覧を返す。いずれも未指定の場合は全件を返す
// @Summary Event一覧を取得
// @Tags events
// @Produce json
// @Param id query int false "取得するEventのID"
// @Param game_type query string false "絞り込む分類（digital または analog）"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/events [get]
func GetEvents(c *gin.Context) {
	if idStr := c.Query("id"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			respondError(c, http.StatusBadRequest, "id must be a positive integer")
			return
		}

		event, err := service.GetEventByID(uint(id))
		if err != nil {
			respondError(c, http.StatusNotFound, "event not found")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": event,
		})
		return
	}

	if gameType := c.Query("game_type"); gameType != "" {
		if gameType != "digital" && gameType != "analog" {
			respondError(c, http.StatusBadRequest, "game_type must be 'digital' or 'analog'")
			return
		}

		events, err := service.GetEventsByGameType(gameType)
		if err != nil {
			respondError(c, http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": events,
		})
		return
	}

	events, err := service.GetEvents()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": events,
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
	targetTimeJST := c.DefaultQuery("time", lib.NowJST().Format("15:04"))

	// JSTの時刻をそのまま使用（DB・内部処理すべてJST統一）
	if _, err := time.ParseInLocation("15:04", targetTimeJST, lib.JST); err != nil {
		respondError(c, http.StatusBadRequest, "invalid time format (expected HH:MM)")
		return
	}

	// 確率を取得
	probability, err := service.GetActivityProbability(uint(eventID), weekday, targetTimeJST)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event_id":    eventID,
		"weekday":     weekdayInt,
		"time":        targetTimeJST,
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
	var weekday time.Weekday

	weekdayStr := c.Query("weekday")
	if weekdayStr == "" {
		// デフォルト: 今日の曜日（JST）
		weekday = lib.NowJST().Weekday()
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
		eventID, err := strconv.ParseUint(entry.EventID, 10, 32)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid event_id")
			return
		}
		inputs[i] = service.LogEntryInput{
			EventID:                 uint(eventID),
			StatusID:                entry.StatusID,
			EventTime:               entry.EventTime,
			ParticipateStayWatchIDs: entry.ParticipateUsers,
			RoomStayWatchIDs:        entry.RoomUsers,
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

// GetBoard は共有モニター(moment-board)用の集約表示データを取得するAPIハンドラー
// @Summary 共有モニター用の表示データを取得
// @Tags board
// @Produce json
// @Success 200 {object} service.BoardData
// @Failure 500 {object} map[string]interface{}
// @Router /api/board [get]
func GetBoard(c *gin.Context) {
	board, err := service.GetBoardData()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, board)
}

// PostRefreshUserIcons は全ユーザのアイコンURLをSlackから取得し直してDBを更新するAPIハンドラー
func PostRefreshUserIcons(c *gin.Context) {
	updated, err := service.RefreshAllUserIcons()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "icon URLs refreshed",
		"updated": updated,
	})
}
