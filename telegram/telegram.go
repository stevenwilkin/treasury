package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

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

func (t *Telegram) Notify(a alert.Alert) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.ApiToken)

	params := sendMessageParams{ChatId: t.ChatId, Text: a.Message()}
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

func NewFromEnv() *Telegram {
	chatId, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	if err != nil {
		log.Fatal(err.Error())
	}

	return &Telegram{
		ApiToken: os.Getenv("TELEGRAM_API_TOKEN"),
		ChatId:   chatId}
}

var _ alert.Notifier = &Telegram{}
