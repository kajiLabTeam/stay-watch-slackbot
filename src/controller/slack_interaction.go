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

	if len(interaction.ActionCallback.BlockActions) > 0 {
		handleBlockAction(c, interaction)
		return
	}

	if interaction.Type == slack.InteractionTypeViewSubmission {
		handleViewSubmission(c, interaction)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func handleBlockAction(c *gin.Context, interaction slack.InteractionCallback) {
	action := interaction.ActionCallback.BlockActions[0]

	switch action.ActionID {
	case "select_user":
		handleSelectUser(c, interaction, action)
	}
}

func handleSelectUser(c *gin.Context, interaction slack.InteractionCallback, action *slack.BlockAction) {
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
}

func handleViewSubmission(c *gin.Context, interaction slack.InteractionCallback) {
	switch interaction.View.CallbackID {
	case "register_event":
		handleRegisterEvent(c, interaction)
	case "select_events":
		handleSelectEvents(c, interaction)
	default:
		c.JSON(http.StatusOK, gin.H{})
	}
}

func handleRegisterEvent(c *gin.Context, interaction slack.InteractionCallback) {
	values := interaction.View.State.Values
	responseURL := interaction.View.PrivateMetadata
	name := values["name_block"]["name_input"].Value
	numStr := values["number_block"]["number_input"].Value

	numInt, err := strconv.Atoi(numStr)
	if err != nil {
		respondError(c, http.StatusBadRequest, "number must be an integer")
		return
	}

	typeID := parseTypeID(c, values)
	if typeID == nil {
		return
	}

	toolIDs := parseToolIDs(c, values)
	if toolIDs == nil {
		return
	}

	if _, err := service.RegisterEvent(name, numInt, *typeID, toolIDs); err != nil {
		if err.Error() == "event already exists" {
			_, _, _ = api.PostMessage("", slack.MsgOptionReplaceOriginal(responseURL), slack.MsgOptionText("登録済みのイベントです", false))
			return
		}
		_, _, _, _ = api.SendMessage("", slack.MsgOptionReplaceOriginal(interaction.ResponseURL), slack.MsgOptionText("Error: "+err.Error(), false))
		return
	}

	_, _, _ = api.PostMessage("", slack.MsgOptionReplaceOriginal(responseURL), slack.MsgOptionText("登録が完了しました。", false))
	c.JSON(http.StatusOK, gin.H{})
}

func parseTypeID(c *gin.Context, values map[string]map[string]slack.BlockAction) *uint {
	var typeID uint
	typeBlock, ok := values["type_block"]
	if !ok {
		return &typeID
	}
	typeSelect, ok := typeBlock["type_select"]
	if !ok || typeSelect.SelectedOption.Value == "" {
		return &typeID
	}
	typeIDInt, err := strconv.Atoi(typeSelect.SelectedOption.Value)
	if err != nil {
		respondError(c, http.StatusBadRequest, "type id must be an integer")
		return nil
	}
	typeID = uint(typeIDInt)
	return &typeID
}

func parseToolIDs(c *gin.Context, values map[string]map[string]slack.BlockAction) []uint {
	var toolIDs []uint
	toolBlock, ok := values["tool_block"]
	if !ok {
		return toolIDs
	}
	toolCheckbox, ok := toolBlock["tool_checkbox"]
	if !ok {
		return toolIDs
	}
	for _, opt := range toolCheckbox.SelectedOptions {
		toolIDInt, err := strconv.Atoi(opt.Value)
		if err != nil {
			respondError(c, http.StatusBadRequest, "tool id must be an integer")
			return nil
		}
		toolIDs = append(toolIDs, uint(toolIDInt))
	}
	return toolIDs
}

func handleSelectEvents(c *gin.Context, interaction slack.InteractionCallback) {
	slackUserID := interaction.User.ID
	responseURL := interaction.View.PrivateMetadata
	options := interaction.View.State.Values["event_select_block"]["event_checkbox"].SelectedOptions

	for _, opt := range options {
		_, _ = service.RegisterCorrespond(opt.Text.Text, slackUserID)
	}

	_, _, _ = api.PostMessage("", slack.MsgOptionReplaceOriginal(responseURL), slack.MsgOptionText("登録が完了しました。", false))
	c.JSON(http.StatusOK, gin.H{})
}
