package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kajiLabTeam/stay-watch-slackbot/service"
	"github.com/slack-go/slack"
)

func PostRegisterUserCommand(c *gin.Context) {
	text := c.PostForm("text")
	userID := c.PostForm("user_id")

	_, err := service.RegisterUser(userID, text)
	if err != nil {
		if err.Error() == "user already exists" {
			respondSlackError(c, fmt.Sprintf("User %s already exists.", text))
			return
		}
		respondSlackError(c, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	respondSlackSuccess(c, fmt.Sprintf("User %s registered successfully.", text))
}

func PostRegisterEventCommand(c *gin.Context) {
	s, err := slack.SlashCommandParse(c.Request)
	if err != nil {
		log.Printf("Error parsing slash command: %v", err)
		respondError(c, http.StatusBadRequest, "bad request")
		return
	}
	modalRequest := slack.ModalViewRequest{
		Type:       slack.ViewType("modal"),
		Title:      slack.NewTextBlockObject("plain_text", "登録フォーム", false, false),
		Submit:     slack.NewTextBlockObject("plain_text", "送信", false, false),
		PrivateMetadata: s.ResponseURL,
		Close:      slack.NewTextBlockObject("plain_text", "閉じる", false, false),
		CallbackID: "register_event",
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
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	respondSlackSuccess(c, "モーダルを開きました。")
}

func PostRegisterCorrespondCommand(c *gin.Context) {
	s, err := slack.SlashCommandParse(c.Request)
	if err != nil {
		log.Printf("Error parsing slash command: %v", err)
		respondError(c, http.StatusBadRequest, "bad request")
		return
	}

	events, err := service.GetEvents()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}

	var options []*slack.OptionBlockObject
	for _, event := range events {
		option := slack.OptionBlockObject{
			Text:  slack.NewTextBlockObject("plain_text", event.Name, false, false),
			Value: fmt.Sprintf("%d", event.ID),
		}
		options = append(options, &option)
	}

	modalRequest := slack.ModalViewRequest{
		Type:            slack.VTModal,
		CallbackID:      "select_events",
		Title:           slack.NewTextBlockObject("plain_text", "ユーザ選択", false, false),
		Close:           slack.NewTextBlockObject("plain_text", "閉じる", false, false),
		Submit:          slack.NewTextBlockObject("plain_text", "決定", false, false),
		PrivateMetadata: s.ResponseURL,
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.InputBlock{
					Type:    slack.MBTInput,
					BlockID: "event_select_block",
					Label:   slack.NewTextBlockObject("plain_text", "話題を選択してください", false, false),
					Element: slack.NewCheckboxGroupsBlockElement("event_checkbox", options...),
				},
			},
		},
	}
	_, err = api.OpenView(s.TriggerID, modalRequest)
	if err != nil {
		log.Printf("Error opening view: %v", err)
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	respondSlackSuccess(c, "モーダルを開きました。")
}

func PostRegisterTypeCommand(c *gin.Context) {
	text := c.PostForm("text")

	_, err := service.RegisterType(text)
	if err != nil {
		if err.Error() == "type already exists" {
			respondSlackError(c, fmt.Sprintf("Type %s already exists.", text))
			return
		}
		respondSlackError(c, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	respondSlackSuccess(c, fmt.Sprintf("Type %s registered successfully.", text))
}

func PostRegisterToolCommand(c *gin.Context) {
	text := c.PostForm("text")

	_, err := service.RegisterTool(text)
	if err != nil {
		if err.Error() == "tool already exists" {
			respondSlackError(c, fmt.Sprintf("Tool %s already exists.", text))
			return
		}
		respondSlackError(c, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	respondSlackSuccess(c, fmt.Sprintf("Tool %s registered successfully.", text))
}

func PostRegisterStatusCommand(c *gin.Context) {
	text := c.PostForm("text")

	_, err := service.RegisterStatus(text)
	if err != nil {
		if err.Error() == "status already exists" {
			respondSlackError(c, fmt.Sprintf("Status %s already exists.", text))
			return
		}
		respondSlackError(c, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	respondSlackSuccess(c, fmt.Sprintf("Status %s registered successfully.", text))
}
