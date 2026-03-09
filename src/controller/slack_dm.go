package controller

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
	"github.com/slack-go/slack"
)

func SendDM(c *gin.Context) {
	logFile, err := os.OpenFile("../log/dm_send.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to open log file")
		return
	}
	defer func() { _ = logFile.Close() }()
	logger := log.New(logFile, "", 0)

	targetWeekday, err := parseTargetWeekday(c.DefaultQuery("weekday", ""))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weekday parameter. Use integer 0-6 (0=Monday, ..., 6=Sunday)"})
		return
	}

	users, userMessages := service.NotifyByEvent(targetWeekday)

	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "No users found"})
		return
	}
	if len(userMessages) == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "No message found"})
		return
	}
	for _, user := range users {
		sendDMToUser(c, logger, user, userMessages)
	}
}

func parseTargetWeekday(param string) (time.Weekday, error) {
	if param == "" {
		jst := time.FixedZone("JST", 9*60*60)
		tomorrow := time.Now().In(jst).AddDate(0, 0, 1)
		return tomorrow.Weekday(), nil
	}
	weekdayInt, err := strconv.Atoi(param)
	if err != nil || weekdayInt < 0 || weekdayInt > 6 {
		return 0, fmt.Errorf("invalid weekday: %s", param)
	}
	return time.Weekday((weekdayInt + 1) % 7), nil
}

func buildMessageForUser(eventMessages map[int][]string, corresponds []model.Correspond) string {
	var b strings.Builder
	for _, co := range corresponds {
		m, ok := eventMessages[int(co.EventID)]
		if !ok {
			continue
		}
		for _, v := range m {
			b.WriteString(v)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func sendDMToUser(c *gin.Context, logger *log.Logger, user model.User, userMessages map[int]map[int][]string) {
	eventMessages, hasMessages := userMessages[int(user.ID)]
	if !hasMessages || len(eventMessages) == 0 {
		return
	}
	if len(user.Corresponds) == 0 {
		return
	}

	channel, _, _, err := api.OpenConversation(&slack.OpenConversationParameters{
		ReturnIM: true,
		Users:    []string{user.SlackID},
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to open conversation")
		return
	}

	message := buildMessageForUser(eventMessages, user.Corresponds)
	if message == "" {
		return
	}

	_, _, err = api.PostMessage(channel.ID, slack.MsgOptionText(message, false))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to send message")
		return
	}

	jst := time.FixedZone("JST", 9*60*60)
	now := time.Now().UTC().In(jst)
	logger.Printf("[%s] 送信先: %s (SlackID: %s)\n推奨活動内容:\n%s\n---\n",
		now.Format("2006-01-02 15:04:05"),
		user.Name,
		user.SlackID,
		message)
}
