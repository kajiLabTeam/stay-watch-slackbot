package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
	"github.com/slack-go/slack"
)

func PostSlackCommandTest(c *gin.Context) {
	command := c.PostForm("command")
	text := c.PostForm("text")
	// userName := c.PostForm("user_name")
	userID := c.PostForm("user_id")

	c.JSON(http.StatusOK, gin.H{
		"response_type": "in_channel",
		"text":          fmt.Sprintf("Command: %s, Text: %s, User: %s", command, text, userID),
	})
}

func PostRegisterUserCommand(c *gin.Context) {
	text := c.PostForm("text")
	userID := c.PostForm("user_id")

	_, err := service.RegisterUser(userID, text)
	if err != nil {
		if err.Error() == "user already exists" {
			c.JSON(http.StatusOK, gin.H{
				"response_type": "in_channel",
				"text":          fmt.Sprintf("User %s already exists.", text),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"response_type": "in_channel",
			"text":          fmt.Sprintf("Error: %s", err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"response_type": "in_channel",
		"text":          fmt.Sprintf("User %s registered successfully.", text),
	})
}

func PostRegisterTagCommand(c *gin.Context) {
	s, err := slack.SlashCommandParse(c.Request)
	if err != nil {
		log.Printf("Error parsing slash command: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	modalRequest := slack.ModalViewRequest{
		Type:            slack.ViewType("modal"),
		Title:           slack.NewTextBlockObject("plain_text", "登録フォーム", false, false),
		Submit:          slack.NewTextBlockObject("plain_text", "送信", false, false),
		PrivateMetadata: s.ResponseURL,
		Close:           slack.NewTextBlockObject("plain_text", "閉じる", false, false),
		CallbackID:      "register_tag",
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				// 名前
				slack.NewInputBlock(
					"name_block",
					slack.NewTextBlockObject("plain_text", "話題を入力してください", false, false),
					slack.NewTextBlockObject("plain_text", "名前", false, false),
					slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", "例：スマブラ、Android", false, false), "name_input"),
				),
				// 人数
				slack.NewInputBlock(
					"number_block",
					slack.NewTextBlockObject("plain_text", "最低限必要な人数を入力してください", false, false),
					slack.NewTextBlockObject("plain_text", "人数", false, false),
					slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", "例：2", false, false), "number_input"),
				),
			},
		},
	}
	_, err = api.OpenView(s.TriggerID, modalRequest)
	if err != nil {
		log.Printf("Error opening view: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response_type": "in_channel", "text": "モーダルを開きました。"})
}

type Tag struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func PostRegisterCorrespondCommand(c *gin.Context) {
	s, err := slack.SlashCommandParse(c.Request)
	if err != nil {
		log.Printf("Error parsing slash command: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	// tags, err := service.GetTags()
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	// 	return
	// }

	tags := []Tag{
		{ID: 1, Name: "スマブラ"},
		{ID: 2, Name: "Android"},
		{ID: 3, Name: "iOS"},
		{ID: 4, Name: "Python"},
		{ID: 5, Name: "Go"},
		{ID: 6, Name: "Java"},
		{ID: 7, Name: "JavaScript"},
		{ID: 8, Name: "Ruby"},
		{ID: 9, Name: "PHP"},
		{ID: 10, Name: "C++"},
		{ID: 11, Name: "C#"},
		{ID: 12, Name: "Swift"},
		{ID: 13, Name: "Kotlin"},
		{ID: 14, Name: "Rust"},
		{ID: 15, Name: "Dart"},
		{ID: 16, Name: "TypeScript"},
	}

	var options []*slack.OptionBlockObject
	for _, tag := range tags {
		option := slack.OptionBlockObject{
			Text:  slack.NewTextBlockObject("plain_text", tag.Name, false, false),
			Value: fmt.Sprintf("%d", tag.ID),
		}
		options = append(options, &option)
	}

	modalRequest := slack.ModalViewRequest{
		Type:            slack.ViewType("modal"),
		CallbackID:      "select_tags",
		Title:           slack.NewTextBlockObject("plain_text", "ユーザ選択", false, false),
		Close:           slack.NewTextBlockObject("plain_text", "閉じる", false, false),
		Submit:          slack.NewTextBlockObject("plain_text", "決定", false, false),
		PrivateMetadata: s.ResponseURL,
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.InputBlock{
					Type:    slack.MBTInput,
					BlockID: "tag_select_block",
					// Label:   slack.NewTextBlockObject("plain_text", "話題を選択してください", false, false),
					Element: slack.NewOptionsMultiSelectBlockElement(
						slack.OptTypeStatic,
						slack.NewTextBlockObject("plain_text", "話題を選択してください", false, false),
						"tag_checkbox",
						options...,
					),
				},
			},
		},
	}
	_, err = api.OpenView(s.TriggerID, modalRequest)
	if err != nil {
		log.Printf("Error opening view: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response_type": "in_channel", "text": "モーダルを開きました。"})
}
