package deribit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

type authResponse struct {
	Result struct {
		AccessToken string `json:"access_token"`
	} `json:"result"`
}

type requestMessage struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

type portfolioResponse struct {
	Method string `json:"method"`
	Params struct {
		Data struct {
			Equity float64 `json:"equity"`
		} `json:"data"`
	} `json:"params"`
}

type Deribit struct {
	ApiId     string
	ApiSecret string
}

func (d *Deribit) Equity() chan float64 {
	log.WithField("venue", "deribit").Info("Subscribing to equity")

	u := fmt.Sprintf(
		"https://www.deribit.com/api/v2/public/auth?client_id=%s&client_secret=%s&grant_type=client_credentials",
		d.ApiId,
		d.ApiSecret)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Panic(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}

	var response authResponse
	json.Unmarshal(body, &response)
	accessToken := response.Result.AccessToken

	socketUrl := url.URL{Scheme: "wss", Host: "www.deribit.com", Path: "/ws/api/v2"}
	c, _, err := websocket.DefaultDialer.Dial(socketUrl.String(), nil)
	if err != nil {
		log.Panic(err.Error())
	}

	request := requestMessage{
		Method: "/private/subscribe",
		Params: map[string]interface{}{
			"channels":     []string{"user.portfolio.BTC"},
			"access_token": accessToken}}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		log.Panic(err.Error())
	}

	err = c.WriteMessage(websocket.TextMessage, jsonRequest)
	if err != nil {
		log.Panic(err.Error())
	}

	ch := make(chan float64)

	go func() {
		defer c.Close()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Error(err)
				return
			}

			var response portfolioResponse
			json.Unmarshal(message, &response)

			if response.Method == "" {
				continue
			}

			log.WithFields(log.Fields{
				"venue": "deribit",
				"value": response.Params.Data.Equity,
			}).Debug("Received equity")
			ch <- response.Params.Data.Equity
		}
	}()

	return ch
}
