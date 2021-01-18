package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/stevenwilkin/treasury/alert"
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

func (t *Telegram) Notify(text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.ApiToken)

	params := sendMessageParams{ChatId: t.ChatId, Text: text}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonParams))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response sendMessageResponse
	json.Unmarshal(body, &response)

	if response.Ok {
		return nil
	} else {
		return errors.New("Error sending message")
	}
}

var _ alert.Notifier = &Telegram{}
