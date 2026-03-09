package main

import (
	"github.com/kajiLabTeam/stay-watch-slackbot/router"
)

// @title Stay Watch Slackbot API
// @version 1.0
// @description 研究室の来訪予測・活動管理のためのAPI
// @host localhost:8085
// @BasePath /
func main() {
	router.Router()
}
