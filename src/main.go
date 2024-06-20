package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Users struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProbabilityStayingResponse struct {
	UserId      int     `json:"userId"`
	UserName    string  `json:"userName"`
	Probability float64 `json:"probability"`
}

// You more than likely want your "Bot User OAuth Access Token" which starts with "xoxb-"
var api = slack.New("")

func main() {
	signingSecret := ""

	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err := sv.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		}
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			log.Default().Println(eventsAPIEvent.InnerEvent.Data)
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				url := ""
				req, _ := http.NewRequest("GET", url, nil)
				client := new(http.Client)
				resp, err := client.Do(req)
				if err != nil {
					if _, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText("Sorry, I can't get the data.", false)); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
					log.Default().Println(err)
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != 200 {
					if _, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText("Sorry, I can't get the data.", false)); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
					log.Default().Println(resp.StatusCode)
					return
				}
				body, _ := io.ReadAll(resp.Body)
				var users []Users
				if err := json.Unmarshal(body, &users); err != nil {
					if _, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText("Sorry, I can't get the data.", false)); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
					log.Default().Println(err)
					return
				}
				var obo []*slack.OptionBlockObject
				for _, user := range users {
					obo = append(obo, &slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: user.Name}, Value: strconv.FormatInt(user.ID, 5)})
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
								Type:	 slack.OptTypeStatic,
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
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
	})

	http.HandleFunc("/slack/interaction", func(w http.ResponseWriter, r *http.Request) {
		var interaction slack.InteractionCallback
		err := json.Unmarshal([]byte(r.FormValue("payload")), &interaction)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(interaction.ActionCallback.BlockActions) != 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		action := interaction.ActionCallback.BlockActions[0]
		switch action.ActionID {
		case "select_user":
			time := time.Now()
			time_str := time.Format("15:04:05")
			// date := time.Format("2006-01-02")
			// url := "https://staywatch-backend.kajilab.net/api/v1/probability/reporting/before?user_id=1&date=" + date + "&time=" + time_str
			// req, _ := http.NewRequest("GET", url, nil)
			// client := new(http.Client)
			// resp, err := client.Do(req)
			// if err != nil {
			// 	if _, _, _, err := api.SendMessage("", slack.MsgOptionReplaceOriginal(interaction.ResponseURL), slack.MsgOptionText("Sorry, I can't get the data.", false)); err != nil {
			// 		w.WriteHeader(http.StatusInternalServerError)
			// 	}
			// 	return
			// }
			// defer resp.Body.Close()
			// if resp.StatusCode != 200 {
			// 	if _, _, _, err := api.SendMessage("", slack.MsgOptionReplaceOriginal(interaction.ResponseURL), slack.MsgOptionText("Sorry, I can't get the data.", false)); err != nil {
			// 		w.WriteHeader(http.StatusInternalServerError)
			// 	}
			// 	return
			// }
			// body, _ := io.ReadAll(resp.Body)
			// var probability ProbabilityStayingResponse
			// if err := json.Unmarshal(body, &probability); err != nil {
			// 	if _, _, _, err := api.SendMessage("", slack.MsgOptionReplaceOriginal(interaction.ResponseURL), slack.MsgOptionText("Sorry, I can't get the data.", false)); err != nil {
			// 		w.WriteHeader(http.StatusInternalServerError)
			// 	}
			// 	return
			// }

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
							// Text: "```\n" + probability.UserName + "が" + time_str + "までに研究室に来る確率 : " + strconv.FormatFloat(probability.Probability, 'f', 2, 64) + "%\n ```",
							Text: "```\n" + action.SelectedOption.Text.Text + "が" + time_str + "までに研究室に来る確率 : 50%```",
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
	})

	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":8080", nil)
}
