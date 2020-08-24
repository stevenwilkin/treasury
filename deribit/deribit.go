package deribit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

type Deribit struct {
	ApiId        string
	ApiSecret    string
	Test         bool
	_accessToken string
}

func (d *Deribit) accessToken() string {
	if d._accessToken != "" {
		return d._accessToken
	}

	log.WithField("venue", "deribit").Debug("Fetching access token")

	v := url.Values{}
	v.Set("client_id", d.ApiId)
	v.Set("client_secret", d.ApiSecret)
	v.Set("grant_type", "client_credentials")

	u := url.URL{
		Scheme:   "https",
		Host:     d.hostname(),
		Path:     "/api/v2/public/auth",
		RawQuery: v.Encode()}

	req, err := http.NewRequest("GET", u.String(), nil)
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
	d._accessToken = response.Result.AccessToken

	return d._accessToken
}

func (d *Deribit) hostname() string {
	if d.Test {
		return "test.deribit.com"
	} else {
		return "www.deribit.com"
	}
}

func (d *Deribit) subscribe(channels []string) *websocket.Conn {
	socketUrl := url.URL{Scheme: "wss", Host: d.hostname(), Path: "/ws/api/v2"}

	c, _, err := websocket.DefaultDialer.Dial(socketUrl.String(), nil)
	if err != nil {
		log.Panic(err.Error())
	}

	request := requestMessage{
		Method: "/private/subscribe",
		Params: map[string]interface{}{
			"channels":     channels,
			"access_token": d.accessToken()}}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		log.Panic(err.Error())
	}

	if err = c.WriteMessage(websocket.TextMessage, jsonRequest); err != nil {
		log.Panic(err.Error())
	}

	return c
}

func (d *Deribit) Equity() chan float64 {
	log.WithField("venue", "deribit").Info("Subscribing to equity")

	c := d.subscribe([]string{"user.portfolio.BTC"})
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

func (d *Deribit) GetSize() int {
	u := "https://www.deribit.com/api/v2/private/get_positions?currency=BTC&kind=future"
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Panic(err.Error())
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.accessToken()))
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

	var response positionsResponse
	json.Unmarshal(body, &response)

	size := 0.0
	for _, position := range response.Result {
		size += position.Size
	}

	return int(math.Abs(size))
}
