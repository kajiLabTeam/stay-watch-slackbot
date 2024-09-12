package controller

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

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
			message := strings.Split(ev.Text, " ")
			if len(message) < 2 {
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
				return
			}

			command := message[1]
			switch command {
			case "addme":
				user := slack.User{ID: ev.User}
				if err := service.SetUser(user); err != nil {
					if _, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText("登録に失敗しました", false)); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
						return
					}
					return
				}
				if _, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText("登録しました", false)); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
					return
				}
				return
			case "tag":
				tags, _ := service.GetAllTagsObject()
				// optionObject = append(optionObject, &slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "新しいタグを追加"}, Value: "new"})
				_, _, err = api.PostMessage(ev.Channel, slack.MsgOptionBlocks(
					slack.SectionBlock{
						Type: slack.MBTSection,
						Text: &slack.TextBlockObject{
							Type: slack.PlainTextType,
							Text: "自分に当てはまるタグは？",
						},
						Accessory: &slack.Accessory{
							MultiSelectElement: &slack.MultiSelectBlockElement{
								ActionID: "select_tag",
								Type:     slack.MultiOptTypeStatic,
								Placeholder: &slack.TextBlockObject{
									Type: slack.PlainTextType,
									Text: "タグを選択",
								},
								Options: tags,
							},
						},
					},
				))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
					return
				}
			}
			return
		}
	}
}
