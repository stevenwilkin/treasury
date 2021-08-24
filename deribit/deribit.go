package deribit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

type Deribit struct {
	ApiId        string
	ApiSecret    string
	Test         bool
	_accessToken string
	expiresIn    time.Time
}

func (d *Deribit) accessToken() (string, error) {
	if d._accessToken != "" && d.expiresIn.After(time.Now()) {
		return d._accessToken, nil
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
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response authResponse
	json.Unmarshal(body, &response)
	d._accessToken = response.Result.AccessToken
	expirySecs := time.Second * time.Duration(response.Result.ExpiresIn-10)
	d.expiresIn = time.Now().Add(expirySecs)

	return d._accessToken, nil
}

func (d *Deribit) hostname() string {
	if d.Test {
		return "test.deribit.com"
	} else {
		return "www.deribit.com"
	}
}

func (d *Deribit) subscribe(channels []string) (*websocket.Conn, error) {
	socketUrl := url.URL{Scheme: "wss", Host: d.hostname(), Path: "/ws/api/v2"}

	c, _, err := websocket.DefaultDialer.Dial(socketUrl.String(), nil)
	if err != nil {
		return &websocket.Conn{}, err
	}

	authRequest := requestMessage{
		Method: "/public/auth",
		Params: map[string]interface{}{
			"client_id":     d.ApiId,
			"client_secret": d.ApiSecret,
			"grant_type":    "client_credentials"}}

	if err = c.WriteJSON(authRequest); err != nil {
		return &websocket.Conn{}, err
	}

	request := requestMessage{
		Method: "/private/subscribe",
		Params: map[string]interface{}{
			"channels": channels}}

	if err = c.WriteJSON(request); err != nil {
		return &websocket.Conn{}, err
	}

	ticker := time.NewTicker(10 * time.Second)
	testMessage := requestMessage{Method: "/public/test"}

	go func() {
		for {
			if err = c.WriteJSON(testMessage); err != nil {
				log.WithField("venue", "deribit").Debug("Heartbeat stopping")
				return
			}
			<-ticker.C
		}
	}()

	return c, nil
}

func (d *Deribit) Equity() chan float64 {
	log.WithField("venue", "deribit").Info("Subscribing to equity")

	ch := make(chan float64)
	c, err := d.subscribe([]string{"user.portfolio.BTC"})
	if err != nil {
		log.WithField("venue", "deribit").Warn(err.Error())
		close(ch)
		return ch
	}

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.WithField("venue", "deribit").Warn(err.Error())
				c.Close()
				close(ch)
				return
			}

			var response portfolioResponse
			json.Unmarshal(message, &response)

			if response.Method == "" {
				continue
			}

			log.WithFields(log.Fields{
				"venue": "deribit",
				"asset": "BTC",
				"value": response.Params.Data.Equity,
			}).Debug("Received equity")
			ch <- response.Params.Data.Equity
		}
	}()

	return ch
}

func (d *Deribit) get(path string, params url.Values, result interface{}) error {
	u := url.URL{
		Scheme:   "https",
		Host:     d.hostname(),
		Path:     path,
		RawQuery: params.Encode()}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	accessToken, err := d.accessToken()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	json.Unmarshal(body, result)

	return nil
}

func (d *Deribit) GetSize() int {
	var response positionsResponse

	err := d.get("/api/v2/private/get_positions",
		url.Values{"currency": {"BTC"}, "kind": {"future"}}, &response)

	if err != nil {
		return 0
	}

	size := 0.0
	for _, position := range response.Result {
		size += position.Size
	}

	return int(math.Abs(size))
}

func (d *Deribit) GetLeverage() (float64, error) {
	var response accountSummaryResponse

	err := d.get("/api/v2/private/get_account_summary",
		url.Values{"currency": {"BTC"}}, &response)

	if err != nil {
		return 0, err
	}

	if response.Result.Equity == 0 {
		return 0, nil
	}

	return (response.Result.InitialMargin / response.Result.Equity) * 100, nil
}

func (d *Deribit) Leverage() chan float64 {
	log.WithField("venue", "deribit").Info("Polling leverage")

	ch := make(chan float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			leverage, err := d.GetLeverage()
			if err != nil {
				log.WithField("venue", "deribit").Warn(err.Error())
				close(ch)
				return
			}

			log.WithFields(log.Fields{
				"venue":    "deribit",
				"leverage": leverage,
			}).Debug("Received leverage")

			ch <- leverage
			<-ticker.C
		}
	}()

	return ch
}
