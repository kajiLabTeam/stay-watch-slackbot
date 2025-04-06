package controller

import (
	"github.com/kajiLabTeam/stay-watch-slackbot/conf"
	"github.com/slack-go/slack"
)

var (
	signingSecret string
	api           *slack.Client
)

func init() {
	s := conf.GetSlackConfig()
	signingSecret = s.GetString("slack.signing_secret")
	// api = slack.New(s.GetString("slack.bot_user_oauth_token"), slack.OptionDebug(true))
	api = slack.New(s.GetString("slack.bot_user_oauth_token"))
}
