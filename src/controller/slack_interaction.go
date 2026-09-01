package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
	"github.com/slack-go/slack"
)

// eventImageRegisterTimeout は画像取得〜保存にかける上限時間。
// Slack への3秒応答とは別に、バックグラウンド処理が無限に残らないようにするためのもの。
const eventImageRegisterTimeout = 2 * time.Minute

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
	case "register_event_image":
		handleRegisterEventImage(c, interaction)
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

	if _, err := service.RegisterEvent(name, numInt); err != nil {
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

func handleSelectEvents(c *gin.Context, interaction slack.InteractionCallback) {
	slackUserID := interaction.User.ID
	responseURL := interaction.View.PrivateMetadata
	options := interaction.View.State.Values["event_select_block"]["event_checkbox"].SelectedOptions

	for _, opt := range options {
		_, _ = service.RegisterEventUser(opt.Text.Text, slackUserID)
	}

	_, _, _ = api.PostMessage("", slack.MsgOptionReplaceOriginal(responseURL), slack.MsgOptionText("登録が完了しました。", false))
	c.JSON(http.StatusOK, gin.H{})
}

// handleRegisterEventImage は選択されたイベントに Slack アップロード画像を紐づける。
//
// Slack は view_submission に3秒以内の応答を要求するが、画像のダウンロードと
// オブジェクトストレージへの保存はそれを超えうる。先に 200 を返してモーダルを閉じ、
// 実際の保存はバックグラウンドで行って response_url へ結果を投稿する。
func handleRegisterEventImage(c *gin.Context, interaction slack.InteractionCallback) {
	values := interaction.View.State.Values
	responseURL := interaction.View.PrivateMetadata

	eventID, err := strconv.ParseUint(values["event_select_block"]["event_select"].SelectedOption.Value, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, slack.NewErrorsViewSubmissionResponse(map[string]string{
			"event_select_block": "話題を選択してください。",
		}))
		return
	}

	files := values["image_block"]["image_input"].Files
	if len(files) == 0 {
		c.JSON(http.StatusOK, slack.NewErrorsViewSubmissionResponse(map[string]string{
			"image_block": "画像が選択されていません。",
		}))
		return
	}
	file := files[0]

	go registerEventImageAsync(uint(eventID), file.URLPrivate, file.Mimetype, responseURL)

	// モーダルを閉じる（空の200応答）
	c.JSON(http.StatusOK, gin.H{})
}

// registerEventImageAsync は画像の保存を行い、結果を response_url へ投稿する。
// リクエストのライフサイクルから外れるため、独自のタイムアウト付き context を使う。
func registerEventImageAsync(eventID uint, urlPrivate, mimetype, responseURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), eventImageRegisterTimeout)
	defer cancel()

	event, err := service.RegisterEventImageFromSlack(ctx, eventID, urlPrivate, mimetype)
	if err != nil {
		log.Printf("failed to register event image (event %d): %v", eventID, err)
		postToResponseURL(responseURL, "画像の登録に失敗しました: "+err.Error())
		return
	}

	postToResponseURL(responseURL, fmt.Sprintf("「%s」の画像を登録しました。", event.Name))
}

// postToResponseURL は Slack の response_url へメッセージを投稿する
func postToResponseURL(responseURL, text string) {
	if responseURL == "" {
		return
	}
	if _, _, err := api.PostMessage("", slack.MsgOptionReplaceOriginal(responseURL), slack.MsgOptionText(text, false)); err != nil {
		log.Printf("failed to post to response_url: %v", err)
	}
}
