package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var api = slack.New(os.Getenv("TOKEN"))

type ChatGPTResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type ChatGPTRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

// ChatGPTにメッセージを送信する
func postChat(message string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	// リクエストボディを作成する
	var chatGPTRequest ChatGPTRequest
	chatGPTRequest.Model = "gpt-3.5-turbo"
	chatGPTRequest.Messages = append(chatGPTRequest.Messages, struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{
		Role:    "user",
		Content: message,
	})
	// jsonに変換する
	reqBody, err := json.Marshal(chatGPTRequest)
	buf := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	// リクエストを送信する
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// レスポンスをパースする
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var chatGPTResponse ChatGPTResponse
	if err := json.Unmarshal(body, &chatGPTResponse); err != nil {
		return "", err
	}
	return chatGPTResponse.Choices[0].Message.Content, nil
}

func handle(w http.ResponseWriter, r *http.Request) {

	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//  Salckからのリクエストであるかどうかを検証する
	sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

	// リクエストを検証したら、イベントをパースする
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// URL Verificationの場合は、Challengeを返す
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

	println(eventsAPIEvent.Type)

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:

			message := ev.Text
			if message == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			log.Println("Message: ", message)
			// userIdを取り除く
			parttern := regexp.MustCompile("<@.+?>")

			reply, err := postChat(parttern.ReplaceAllString(message, ""))
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if _, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText(reply, false)); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	w.Write([]byte("Hello World"))
}

func main() {
	http.HandleFunc("/", handle)

	fmt.Println("Server is running on port 3000")

	// Listen and serve on
	http.ListenAndServe(":3000", nil)
}
