package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Binance struct {
	ApiKey    string
	ApiSecret string
	Testnet   bool
}

func (b *Binance) hostname() string {
	if b.Testnet {
		return "testnet.binance.vision"
	} else {
		return "api.binance.com"
	}
}

func (b *Binance) wsHostname() string {
	if b.Testnet {
		return "testnet.binance.vision"
	} else {
		return "stream.binance.com:9443"
	}
}

func (b *Binance) sign(s string) string {
	h := hmac.New(sha256.New, []byte(b.ApiSecret))
	io.WriteString(h, string(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (b *Binance) GetBalances() (float64, float64, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	params := fmt.Sprintf("timestamp=%d", timestamp)
	signedParams := fmt.Sprintf("%s&signature=%s", params, b.sign(params))

	u := url.URL{
		Scheme:   "https",
		Host:     b.hostname(),
		Path:     "/api/v3/account",
		RawQuery: signedParams}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, 0, err
	}

	req.Header.Set("X-MBX-APIKEY", b.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var response accountResponse
	json.Unmarshal(body, &response)

	var btc, usdt float64

	for _, asset := range response.Balances {
		if asset.Asset != "BTC" && asset.Asset != "USDT" {
			continue
		}

		free, _ := strconv.ParseFloat(asset.Free, 64)
		locked, _ := strconv.ParseFloat(asset.Locked, 64)
		total := free + locked

		if asset.Asset == "BTC" {
			btc = total
		} else {
			usdt = total
		}
	}

	return btc, usdt, nil
}

func (b *Binance) subscribeToPrice() *websocket.Conn {
	u := url.URL{
		Scheme: "wss",
		Host:   b.wsHostname(),
		Path:   "/ws/btcusdt@aggTrade"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Panic(err.Error())
	}

	return c
}

func (b *Binance) Price() chan float64 {
	log.WithFields(log.Fields{
		"venue":  "binance",
		"symbol": "BTCUSDT",
	}).Info("Subscribing to price")

	c := b.subscribeToPrice()
	ch := make(chan float64)

	go func() {
		defer c.Close()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.WithField("venue", "binance").Info("Reconnecting to price subscription")
				c.Close()
				c = b.subscribeToPrice()
				continue
			}

			var ticker tickerMessage
			json.Unmarshal(message, &ticker)

			price, _ := strconv.ParseFloat(ticker.P, 64)
			ch <- price
		}
	}()

	return ch
}
