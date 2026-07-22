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

// PostListUsersCommand は登録済みユーザの一覧をテキストで返す
func PostListUsersCommand(c *gin.Context) {
	users, err := service.ListAllUsers()
	if err != nil {
		respondSlackError(c, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	if len(users) == 0 {
		respondSlackSuccess(c, "登録されているユーザはいません。")
		return
	}

	message := "登録ユーザ一覧:\n"
	for _, u := range users {
		message += fmt.Sprintf("- %s (SlackID: %s)\n", u.Name, u.SlackID)
	}
	respondSlackSuccess(c, message)
}

// PostDeleteUserCommand はコマンドのtextで指定したユーザ名のユーザを削除する
func PostDeleteUserCommand(c *gin.Context) {
	name := c.PostForm("text")
	if name == "" {
		respondSlackError(c, "削除するユーザ名を指定してください。例: /delete_user 山田太郎")
		return
	}

	if err := service.DeleteUserByName(name); err != nil {
		if err.Error() == "user not found" {
			respondSlackError(c, fmt.Sprintf("User %s not found.", name))
			return
		}
		respondSlackError(c, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	respondSlackSuccess(c, fmt.Sprintf("User %s deleted successfully.", name))
}

// PostDeleteOBUsersCommand はStayWatch側でOBタグ（id:13, name:"OB"）が
// 付与されているユーザーを一括削除する
func PostDeleteOBUsersCommand(c *gin.Context) {
	deleted, err := service.DeleteOBUsers()
	if err != nil {
		respondSlackError(c, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	if len(deleted) == 0 {
		respondSlackSuccess(c, "OBタグが付与されたユーザはいませんでした。")
		return
	}

	message := fmt.Sprintf("以下の%d名のOBユーザを削除しました:\n", len(deleted))
	for _, name := range deleted {
		message += fmt.Sprintf("- %s\n", name)
	}
	respondSlackSuccess(c, message)
}

func PostRegisterEventCommand(c *gin.Context) {
	s, err := slack.SlashCommandParse(c.Request)
	if err != nil {
		log.Printf("Error parsing slash command: %v", err)
		respondError(c, http.StatusBadRequest, "bad request")
		return
	}

	blocks := []slack.Block{
		// 名前
		slack.NewInputBlock(
			"name_block",
			slack.NewTextBlockObject("plain_text", "話題を入力してください", false, false),
			slack.NewTextBlockObject("plain_text", "名前", false, false),
			slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", "例：スマブラ、Android", false, false), "name_input"),
		),
		// code
		slack.NewInputBlock(
			"code_block",
			slack.NewTextBlockObject("plain_text", "イベントを一意に定める識別子を入力してください", false, false),
			slack.NewTextBlockObject("plain_text", "Code", false, false),
			slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", "例：1, 0437ac48be2a81", false, false), "code_input"),
		),
		// 人数
		slack.NewInputBlock(
			"number_block",
			slack.NewTextBlockObject("plain_text", "最低限必要な人数を入力してください", false, false),
			slack.NewTextBlockObject("plain_text", "人数", false, false),
			slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", "例：2", false, false), "number_input"),
		),
	}

	modalRequest := slack.ModalViewRequest{
		Type:            slack.ViewType("modal"),
		Title:           slack.NewTextBlockObject("plain_text", "登録フォーム", false, false),
		Submit:          slack.NewTextBlockObject("plain_text", "送信", false, false),
		PrivateMetadata: s.ResponseURL,
		Close:           slack.NewTextBlockObject("plain_text", "閉じる", false, false),
		CallbackID:      "register_event",
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}
	_, err = api.OpenView(s.TriggerID, modalRequest)
	if err != nil {
		log.Printf("Error opening view: %v", err)
		respondError(c, http.StatusInternalServerError, msgInternalServerError)
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
		respondError(c, http.StatusInternalServerError, msgInternalServerError)
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
		respondError(c, http.StatusInternalServerError, msgInternalServerError)
		return
	}
	respondSlackSuccess(c, "モーダルを開きました。")
}
