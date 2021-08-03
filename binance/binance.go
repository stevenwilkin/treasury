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

func (b *Binance) doRequest(method, path string, values url.Values, sign bool) ([]byte, error) {
	var params string

	if sign {
		timestamp := time.Now().UnixNano() / int64(time.Millisecond)
		values.Set("timestamp", fmt.Sprintf("%d", timestamp))
		input := values.Encode()
		params = fmt.Sprintf("%s&signature=%s", input, b.sign(input))
	} else {
		params = values.Encode()
	}

	u := url.URL{
		Scheme:   "https",
		Host:     b.hostname(),
		Path:     path,
		RawQuery: params}

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("X-MBX-APIKEY", b.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}

func (b *Binance) GetBalances() (float64, float64, float64, error) {
	body, err := b.doRequest("GET", "/api/v3/account", url.Values{}, true)
	if err != nil {
		return 0, 0, 0, err
	}

	var response accountResponse
	json.Unmarshal(body, &response)

	var btc, usdt, usdc float64

	for _, asset := range response.Balances {
		switch asset.Asset {
		case "BTC":
			btc = asset.Total()
		case "USDT":
			usdt = asset.Total()
		case "USDC":
			usdc = asset.Total()
		default:
			continue
		}
	}

	return btc, usdt, usdc, nil
}

func (b *Binance) Balances() chan [3]float64 {
	log.WithField("venue", "binance").Info("Polling balances")

	ch := make(chan [3]float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			btc, usdt, usdc, err := b.GetBalances()
			if err != nil {
				log.WithField("venue", "binance").Warn(err.Error())
				close(ch)
				return
			}

			log.WithFields(log.Fields{
				"venue": "binance",
				"btc":   btc,
				"usdt":  usdt,
				"usdc":  usdc,
			}).Debug("Received balances")

			ch <- [3]float64{btc, usdt, usdc}
			<-ticker.C
		}
	}()

	return ch
}

func (b *Binance) subscribe(stream string) (*websocket.Conn, error) {
	u := url.URL{
		Scheme: "wss",
		Host:   b.wsHostname(),
		Path:   fmt.Sprintf("/ws/%s", stream)}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return &websocket.Conn{}, err
	}

	return c, nil
}

func (b *Binance) listenKey() (string, error) {
	body, err := b.doRequest("POST", "/api/v3/userDataStream", url.Values{}, false)
	if err != nil {
		return "", err
	}

	var response listenKeyResponse
	json.Unmarshal(body, &response)

	return response.ListenKey, nil
}

func (b *Binance) Price() chan float64 {
	log.WithFields(log.Fields{
		"venue":  "binance",
		"symbol": "BTCUSDT",
	}).Info("Subscribing to price")

	ch := make(chan float64)

	c, err := b.subscribe("btcusdt@aggTrade")
	if err != nil {
		log.WithField("venue", "binance").Warn(err.Error())
		close(ch)
		return ch
	}

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.WithField("venue", "binance").Warn(err.Error())
				c.Close()
				close(ch)
				return
			}

			var ticker tickerMessage
			json.Unmarshal(message, &ticker)

			price, _ := strconv.ParseFloat(ticker.P, 64)
			ch <- price
		}
	}()

	return ch
}
