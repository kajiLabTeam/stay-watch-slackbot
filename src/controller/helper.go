// Package controller provides HTTP request handlers for the Slack bot application.
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondError は統一されたエラーレスポンスを返す
func respondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
}

// respondSlackError はSlackコマンド用のエラーレスポンスを返す
func respondSlackError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"response_type": "in_channel",
		"text":          message,
	})
}

// respondSlackSuccess はSlackコマンド用の成功レスポンスを返す
func respondSlackSuccess(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"response_type": "in_channel",
		"text":          message,
	})
}
