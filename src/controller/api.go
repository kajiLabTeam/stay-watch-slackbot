package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
)

// BatchRegisterTypeRequest はType一括登録のリクエストボディ
type RegisterTypesRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
}

// BatchRegisterToolRequest はTool一括登録のリクエストボディ
type RegisterToolsRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
}

// BatchRegisterStatusRequest はStatus一括登録のリクエストボディ
type RegisterStatusesRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
}

// GetTypes はType一覧を取得するAPIハンドラー
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
func PostRegisterTypes(c *gin.Context) {
	var req RegisterTypesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	types, errors, err := service.BatchRegisterTypes(req.Names)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "batch registration completed",
		"data":    types,
		"errors":  errors,
	})
}

// Post	RegisterTools はToolを一括登録するAPIハンドラー
func PostRegisterTools(c *gin.Context) {
	var req RegisterToolsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	tools, errors, err := service.BatchRegisterTools(req.Names)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "batch registration completed",
		"data":    tools,
		"errors":  errors,
	})
}

// PostRegisterStatuses はStatusを一括登録するAPIハンドラー
func PostRegisterStatuses(c *gin.Context) {
	var req RegisterStatusesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	statuses, errors, err := service.BatchRegisterStatuses(req.Names)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "batch registration completed",
		"data":    statuses,
		"errors":  errors,
	})
}
