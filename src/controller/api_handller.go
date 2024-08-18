package controller

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/conf"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var (
	signingSecret string
	api           *slack.Client
)

func init() {
	s := conf.GetSlackConfig()
	signingSecret = s.GetString("slack.signing_secret")
	api = slack.New(s.GetString("slack.bot_user_oauth_token"))
}

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
		c.Request.Header.Set("Content-Type", "text/plain")
		c.JSON(http.StatusOK, []byte(r.Challenge))
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}
	}
}

func PostSlackInteraction(c *gin.Context) {
	var interaction slack.InteractionCallback
	err := json.Unmarshal([]byte(c.Request.FormValue("payload")), &interaction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	if len(interaction.ActionCallback.BlockActions) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	action := interaction.ActionCallback.BlockActions[0]
	switch action.ActionID {
	case "select_user":
		probability, time, err := service.GetProbability("2")
		if err != nil {
			api.SendMessage("", slack.MsgOptionReplaceOriginal(interaction.ResponseURL), slack.MsgOptionText("Sorry, I can't get the data.", false))
			return
		}

		_, _, _, err = api.SendMessage(
			"",
			slack.MsgOptionReplaceOriginal(interaction.ResponseURL),
			slack.MsgOptionBlocks(
				slack.SectionBlock{
					Type: slack.MBTSection,
					Text: &slack.TextBlockObject{
						Type: slack.MarkdownType,
						Text: "だれの確率を調べますか？: " + action.SelectedOption.Text.Text,
					},
				},
				slack.SectionBlock{
					Type: slack.MBTSection,
					Text: &slack.TextBlockObject{
						Type: slack.MarkdownType,
						Text: "```\n" + probability.UserName + "が" + time + "までに研究室に来る確率 : " + strconv.FormatFloat(probability.Probability, 'f', 2, 64) + "%\n ```",
					},
				},
			),
		)
		if err != nil {
			api.SendMessage("", slack.MsgOptionReplaceOriginal(interaction.ResponseURL), slack.MsgOptionText("Sorry, I can't get the data.", false))
			return
		}
	default:
		return
	}
}
