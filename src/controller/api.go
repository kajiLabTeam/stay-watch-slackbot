package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
)

// RegisterTypeRequest はType登録のリクエストボディ
type RegisterTypeRequest struct {
	Name string `json:"name" binding:"required"`
}

// RegisterToolRequest はTool登録のリクエストボディ
type RegisterToolRequest struct {
	Name string `json:"name" binding:"required"`
}

// RegisterStatusRequest はStatus登録のリクエストボディ
type RegisterStatusRequest struct {
	Name string `json:"name" binding:"required"`
}

// PostRegisterType はTypeを登録するAPIハンドラー
func PostRegisterType(c *gin.Context) {
	var req RegisterTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	typeObj, err := service.RegisterType(req.Name)
	if err != nil {
		if err.Error() == "type already exists" {
			respondError(c, http.StatusConflict, "type already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "type registered successfully",
		"data":    typeObj,
	})
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

// PostRegisterTool はToolを登録するAPIハンドラー
func PostRegisterTool(c *gin.Context) {
	var req RegisterToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	tool, err := service.RegisterTool(req.Name)
	if err != nil {
		if err.Error() == "tool already exists" {
			respondError(c, http.StatusConflict, "tool already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "tool registered successfully",
		"data":    tool,
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

// PostRegisterStatus はStatusを登録するAPIハンドラー
func PostRegisterStatus(c *gin.Context) {
	var req RegisterStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	status, err := service.RegisterStatus(req.Name)
	if err != nil {
		if err.Error() == "status already exists" {
			respondError(c, http.StatusConflict, "status already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "status registered successfully",
		"data":    status,
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
