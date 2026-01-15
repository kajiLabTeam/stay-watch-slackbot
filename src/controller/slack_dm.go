package controller

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
	"github.com/slack-go/slack"
)

func SendDM(c *gin.Context) {
	// ログファイルを開く
	logFile, err := os.OpenFile("../log/dm_send.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to open log file")
		return
	}
	defer logFile.Close()
	logger := log.New(logFile, "", 0)

	// クエリパラメータから曜日を取得（デフォルトは翌日）
	weekdayParam := c.DefaultQuery("weekday", "")

	var targetWeekday time.Weekday
	if weekdayParam == "" {
		// パラメータが指定されていない場合は翌日
		loc, _ := time.LoadLocation("Asia/Tokyo")
		tomorrow := time.Now().In(loc).AddDate(0, 0, 1)
		targetWeekday = tomorrow.Weekday()
	} else {
		// パラメータから曜日を解析（MySQL WEEKDAY形式: 月=0, 日=6）
		weekdayInt, err := strconv.Atoi(weekdayParam)
		if err != nil || weekdayInt < 0 || weekdayInt > 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weekday parameter. Use integer 0-6 (0=Monday, ..., 6=Sunday)"})
			return
		}
		// MySQL WEEKDAY形式(月=0)からGoのtime.Weekday形式(日=0)に変換
		targetWeekday = time.Weekday((weekdayInt + 1) % 7)
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
			respondError(c, http.StatusInternalServerError, "failed to open conversation")
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
			respondError(c, http.StatusInternalServerError, "failed to send message")
		} else {
			// 送信成功時にログを記録
			loc, _ := time.LoadLocation("Asia/Tokyo")
			now := time.Now().In(loc)
			logger.Printf("[%s] 送信先: %s (SlackID: %s)\n推奨活動内容:\n%s\n---\n",
				now.Format("2006-01-02 15:04:05"),
				user.Name,
				user.SlackID,
				message)
		}
	}
}
