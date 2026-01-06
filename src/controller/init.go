package controller

import (
	"os"

	"github.com/slack-go/slack"
)

var (
	signingSecret string
	api           *slack.Client
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func init() {
	signingSecret = getEnv("SLACK_SIGNING_SECRET", "")
	botToken := getEnv("SLACK_BOT_USER_OAUTH_TOKEN", "")
	// api = slack.New(botToken, slack.OptionDebug(true))
	api = slack.New(botToken)
}
