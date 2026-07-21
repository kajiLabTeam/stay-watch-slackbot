// Package router configures HTTP routes and middleware for the application.
package router

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/controller"
	_ "github.com/kajiLabTeam/stay-watch-slackbot/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// errorOnlyLogger はステータスコードが400以上のリクエストのみをwに記録するミドルウェア
func errorOnlyLogger(w io.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		status := c.Writer.Status()
		if status < http.StatusBadRequest {
			return
		}

		fmt.Fprintf(w, "[GIN] %s | %3d | %13v | %15s | %-7s %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			status,
			time.Since(start),
			c.ClientIP(),
			c.Request.Method,
			path,
		)
	}
}

// labNetwork は研究室LANとして許可するCIDR
var labNetwork = func() *net.IPNet {
	_, network, err := net.ParseCIDR("192.168.100.0/23")
	if err != nil {
		panic(fmt.Sprintf("invalid lab network CIDR: %v", err))
	}
	return network
}()

func Router() {
	gin.DisableConsoleColor()
	f, _ := os.Create("../log/server.log")
	gin.DefaultWriter = io.MultiWriter(f)
	gin.DefaultErrorWriter = io.MultiWriter(f)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(errorOnlyLogger(gin.DefaultWriter))
	_ = r.SetTrustedProxies([]string{"127.0.0.1"})

	r.Use(cors.New(cors.Config{
		// アクセスを許可したいアクセス元
		AllowOrigins: []string{
			"https://staywatch.kajilab.net",
			"http://localhost:3000",
		},
		// 研究室LAN(192.168.100.0/23)とループバックからのアクセスを許可
		AllowOriginFunc: func(origin string) bool {
			u, err := url.Parse(origin)
			if err != nil {
				return false
			}
			host, _, err := net.SplitHostPort(u.Host)
			if err != nil {
				host = u.Host
			}
			ip := net.ParseIP(host)
			if ip == nil {
				return false
			}
			return ip.IsLoopback() || labNetwork.Contains(ip)
		},
		// アクセスを許可したいHTTPメソッド
		AllowMethods: []string{
			"GET",
			"POST",
		},
		// 許可したいHTTPリクエストヘッダ
		AllowHeaders: []string{
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Accept",
			"Authorization",
		},
		// cookieなどの情報を必要とするかどうか
		AllowCredentials: true,
		// preflightリクエストの結果をキャッシュする時間
		MaxAge: 24 * time.Hour,
	}))

	// Slack endpoints
	r.POST("/slack/events", controller.PostSlackEvents)
	r.POST("/slack/interaction", controller.PostSlackInteraction)
	r.POST("/slack/command/add_user", controller.PostRegisterUserCommand)
	r.POST("/slack/command/add_event", controller.PostRegisterEventCommand)
	r.POST("/slack/command/add_correspond", controller.PostRegisterCorrespondCommand)
	r.POST("/slack/command/list_users", controller.PostListUsersCommand)
	r.POST("/slack/command/delete_user", controller.PostDeleteUserCommand)
	r.POST("/slack/command/delete_ob_users", controller.PostDeleteOBUsersCommand)
	r.GET("/notification", controller.SendDM)

	// Swagger
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// REST API endpoints
	r.POST("/api/statuses", controller.PostRegisterStatuses)
	r.GET("/api/statuses", controller.GetStatuses)
	r.GET("/api/events", controller.GetEvents)
	r.GET("/api/events/:id/probability", controller.GetEventProbability)
	r.GET("/api/activities/probabilities", controller.GetAllActivityProbabilities)
	r.POST("/api/logs", controller.PostRegisterLogs)
	r.POST("/api/users/icons/refresh", controller.PostRefreshUserIcons)

	_ = r.Run(":8085")
}
