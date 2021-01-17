package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/stevenwilkin/treasury/alert"

	log "github.com/sirupsen/logrus"
)

type Telegram struct {
	ApiToken string
	ChatId   int
}

type sendMessageParams struct {
	ChatId int    `json:"chat_id"`
	Text   string `json:"text"`
}

type sendMessageResponse struct {
	Ok bool
}

func (t *Telegram) Notify(text string) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.ApiToken)

	params := sendMessageParams{ChatId: t.ChatId, Text: text}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonParams))
	if err != nil {
		log.Error(err.Error())
		return false
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	var response sendMessageResponse
	json.Unmarshal(body, &response)

	return response.Ok
}

var _ alert.Notifier = &Telegram{}
