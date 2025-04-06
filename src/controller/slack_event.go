package controller

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func PostSlackEvents(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	sv, err := slack.NewSecretsVerifier(c.Request.Header, signingSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	if _, err := sv.Write(body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if err := sv.Ensure(); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.Header("Content-Type", "text/plain")
		c.JSON(http.StatusOK, r.Challenge)
		return
	}

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		log.Default().Println(eventsAPIEvent.InnerEvent.Data)
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			obo, err := service.GetUsers()
			if err != nil {
				if _, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText("Sorry, I can't get the data.", false)); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				}
				log.Default().Println(err)
				return
			}

			_, _, err = api.PostMessage(ev.Channel, slack.MsgOptionBlocks(
				slack.SectionBlock{
					Type: slack.MBTSection,
					Text: &slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: "だれの確率を調べますか？",
					},
					Accessory: &slack.Accessory{
						SelectElement: &slack.SelectBlockElement{
							ActionID: "select_user",
							Type:     slack.OptTypeStatic,
							Placeholder: &slack.TextBlockObject{
								Type: slack.PlainTextType,
								Text: "ユーザーを選択",
							},
							Options: obo,
						},
					},
				},
			))
			if err != nil {
				log.Default().Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
			return
		}
		return
	}
}
