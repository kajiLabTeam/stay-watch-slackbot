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
	payload := c.PostForm("payload")

	var interaction slack.InteractionCallback
	if err := json.Unmarshal([]byte(payload), &interaction); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	// --- Block Actionの処理 ---
	if len(interaction.ActionCallback.BlockActions) > 0 {
		action := interaction.ActionCallback.BlockActions[0]

		switch action.ActionID {
		case "select_user":
			userID, err := strconv.Atoi(action.SelectedOption.Value)
			if err != nil {
				respondError(c, http.StatusBadRequest, "invalid user id")
				return
			}

			probability, time, err := service.GetProbability(userID)
			if err != nil {
				_, _, _, _ = api.SendMessage(
					"",
					slack.MsgOptionReplaceOriginal(interaction.ResponseURL),
					slack.MsgOptionText("Sorry, I can't get the data.", false),
				)
				return
			}

			p := probability.Probability * 100
			_, _, _, _ = api.SendMessage(
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
			return
		}
	}

	// --- モーダルSubmitの処理 ---
	if interaction.Type == slack.InteractionTypeViewSubmission && interaction.View.CallbackID == "register_event" {
		values := interaction.View.State.Values
		responseURL := interaction.View.PrivateMetadata
		name := values["name_block"]["name_input"].Value
		numStr := values["number_block"]["number_input"].Value

		numInt, err := strconv.Atoi(numStr)
		if err != nil {
			respondError(c, http.StatusBadRequest, "number must be an integer")
			return
		}

		// DB登録などの処理
		if _, err := service.RegisterEvent(name, numInt); err != nil {
			if err.Error() == "event already exists" {
				api.PostMessage(
					"",
					slack.MsgOptionReplaceOriginal(responseURL),
					slack.MsgOptionText("登録済みのイベントです", false),
				)
				return
			}
			api.SendMessage(
				"",
				slack.MsgOptionReplaceOriginal(interaction.ResponseURL),
				slack.MsgOptionText("Error: "+err.Error(), false),
			)
			return
		}

		api.PostMessage("", slack.MsgOptionReplaceOriginal(responseURL), slack.MsgOptionText("登録が完了しました。", false))

		// 成功レスポンス
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	if interaction.Type == slack.InteractionTypeViewSubmission && interaction.View.CallbackID == "select_events" {
		slackUserID := interaction.User.ID
		responseURL := interaction.View.PrivateMetadata
		options := interaction.View.State.Values["event_select_block"]["event_checkbox"].SelectedOptions

		for _, opt := range options {
			eventName := opt.Text.Text
			service.RegisterCorrespond(eventName, slackUserID)
		}

		api.PostMessage("", slack.MsgOptionReplaceOriginal(responseURL), slack.MsgOptionText("登録が完了しました。", false))

		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// どの処理にも該当しない場合
	c.JSON(http.StatusOK, gin.H{})
}
