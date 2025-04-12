package router

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/controller"
)

func Router() {
	gin.DisableConsoleColor()
	f, _ := os.Create("../log/server.log")
	gin.DefaultWriter = io.MultiWriter(f)

	r := gin.Default()
	r.SetTrustedProxies([]string{
		"103.21.244.0/22",
		"2400:cb00:2048:1::/64",
	})

	r.POST("/slack/events", controller.PostSlackEvents)
	r.POST("/slack/interaction", controller.PostSlackInteraction)
	r.POST("/slack/command/test", controller.PostSlackCommandTest)
	r.POST("/slack/command/add_user", controller.PostRegisterUserCommand)
	r.POST("/slack/command/add_tag", controller.PostRegisterTagCommand)
	r.POST("slack/command/add_correspond", controller.PostRegisterCorrespondCommand)
	r.GET("/notification", controller.SendDM)

	r.Run(":8085")
}
