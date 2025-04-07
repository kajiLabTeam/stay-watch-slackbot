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

	r.Run(":8085")
}
