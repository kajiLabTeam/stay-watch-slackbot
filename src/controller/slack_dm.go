package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
	"github.com/slack-go/slack"
)

func SendDM(c *gin.Context) {
	users, userMessages := service.NotifyByEvent()

	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "No users found"})
		return
	}
	if len(userMessages) == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "No message found"})
		return
	}
	for _, user := range users {
		// ユーザー専用のメッセージを取得
		eventMessages, hasMessages := userMessages[int(user.ID)]
		if !hasMessages || len(eventMessages) == 0 {
			continue
		}

		channel, _, _, err := api.OpenConversation(&slack.OpenConversationParameters{
			ReturnIM: true,
			Users:    []string{user.SlackID},
		})
		// log.Default().Println("user", user.SlackID)
		// log.Default().Println("channel", channel.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open conversation"})
		}
		message := ""
		cos := user.Corresponds
		if len(cos) == 0 {
			continue
		}
		for _, co := range cos {
			// ユーザー専用メッセージからイベントIDでメッセージを取得
			m, ok := eventMessages[int(co.EventID)]
			if !ok {
				continue
			}
			for _, v := range m {
				message += v + "\n"
			}
		}
		if message == "" {
			continue
		}
		_, _, err = api.PostMessage(channel.ID, slack.MsgOptionText(message, false))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
		}
	}
}
