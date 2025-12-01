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
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// r.Use(cors.New(cors.Config{
	// アクセスを許可したいアクセス元
	// 	AllowOrigins: []string{
	// 		"https://",
	// 		"http://localhost:3000",
	// 	},
	// 	// アクセスを許可したいHTTPメソッド(以下の例だとPUTやDELETEはアクセスできません)
	// 	AllowMethods: []string{
	// 		"POST",
	// 	},
	// 	// 許可したいHTTPリクエストヘッダ
	// 	AllowHeaders: []string{
	// 		"Access-Control-Allow-Credentials",
	// 		"Access-Control-Allow-Headers",
	// 		"Content-Type",
	// 		"Content-Length",
	// 		"Accept-Encoding",
	// 		"Accept",
	// 		"Authorization",
	// 	},
	// 	// cookieなどの情報を必要とするかどうか
	// 	AllowCredentials: true,
	// 	// preflightリクエストの結果をキャッシュする時間
	// 	MaxAge: 24 * time.Hour,
	// }))

	r.POST("/slack/events", controller.PostSlackEvents)
	r.POST("/slack/interaction", controller.PostSlackInteraction)
	r.POST("/slack/command/test", controller.PostSlackCommandTest)
	r.POST("/slack/command/add_user", controller.PostRegisterUserCommand)
	r.POST("/slack/command/add_tag", controller.PostRegisterTagCommand)
	r.POST("slack/command/add_correspond", controller.PostRegisterCorrespondCommand)
	r.GET("/notification", controller.SendDM)

	// Activity prediction endpoints
	r.GET("/api/v1/prediction/activity/:event_id", controller.GetActivityPredictionHandler)
	r.GET("/api/v1/prediction/activity/:event_id/weekly", controller.GetWeeklyActivityPredictionsHandler)
	r.GET("/api/v1/prediction/events", controller.GetAllEventsPredictionsHandler)

	r.Run(":8085")
}
