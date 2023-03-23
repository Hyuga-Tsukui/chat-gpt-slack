package main

import (
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	"net/http"
	// "os"
	// "github.com/slack-go/slack"
	// "github.com/slack-go/slack/slackevents"
)

func handle(w http.ResponseWriter, r *http.Request) {

	// signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// //  Salckからのリクエストであるかどうかを検証する
	// sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	// if _, err := sv.Write(body); err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	// if err := sv.Ensure(); err != nil {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	return
	// }

	// // リクエストを検証したら、イベントをパースする
	// eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// println(eventsAPIEvent.Type)

	w.Write([]byte("Hello World"))
}

func main() {
	http.HandleFunc("/", handle)

	fmt.Println("Server is running on port 3000")

	// Listen and serve on
	http.ListenAndServe(":3000", nil)
}
