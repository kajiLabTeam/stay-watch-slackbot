package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
	"github.com/slack-go/slack"
)

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
		userID, _ := strconv.Atoi(action.SelectedOption.Value)
		probability, time, err := service.GetProbability(userID)
		p := probability.Probability * 100
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
						Text: "```\n" + probability.UserName + "が" + time + "までに研究室に来る確率 : " + strconv.FormatFloat(p, 'f', 2, 64) + "% ```",
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
