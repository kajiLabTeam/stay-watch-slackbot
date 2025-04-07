package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PostGASInteraction(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GAS event received"})
}
