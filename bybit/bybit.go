package bybit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Bybit struct {
	ApiKey    string
	ApiSecret string
	Testnet   bool
}

func (b *Bybit) hostname() string {
	if b.Testnet {
		return "api-testnet.bybit.com"
	} else {
		return "api.bybit.com"
	}
}

func (b *Bybit) websocketHostname() string {
	if b.Testnet {
		return "stream-testnet.bybit.com"
	} else {
		return "stream.bybit.com"
	}
}

func getSignature(params map[string]interface{}, key string) string {
	keys := make([]string, len(params))
	i := 0
	_val := ""
	for k, _ := range params {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		_val += fmt.Sprintf("%s=%v&", k, params[k])
	}
	_val = _val[0 : len(_val)-1]
	h := hmac.New(sha256.New, []byte(key))
	io.WriteString(h, _val)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (b *Bybit) timestamp() string {
	return strconv.FormatInt((time.Now().UnixNano() / int64(time.Millisecond)), 10)
}

func (b *Bybit) GetEquity() (float64, error) {
	timestamp := b.timestamp()

	params := map[string]interface{}{
		"api_key":   b.ApiKey,
		"coin":      "BTC",
		"timestamp": timestamp,
	}

	sign := getSignature(params, b.ApiSecret)

	url := fmt.Sprintf(
		"https://api.bybit.com/v2/private/wallet/balance?api_key=%s&coin=BTC&timestamp=%s&sign=%s",
		b.ApiKey,
		timestamp,
		sign,
	)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response equityResponse
	json.Unmarshal(body, &response)

	return response.Result.BTC.Equity, nil
}

func (b *Bybit) Equity() chan float64 {
	log.WithField("venue", "bybit").Info("Polling equity")

	ch := make(chan float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			equity, err := b.GetEquity()
			if err != nil {
				log.WithField("venue", "bybit").Error(err.Error())
				<-ticker.C
				continue
			}

			log.WithFields(log.Fields{
				"venue": "bybit",
				"asset": "BTC",
				"value": equity,
			}).Debug("Received equity")

			ch <- equity
			<-ticker.C
		}
	}()

	return ch
}

func (b *Bybit) GetFundingRate() (float64, float64, error) {
	url := "https://api.bybit.com/v2/public/tickers?symbol=BTCUSD"
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var response fundingResponse
	json.Unmarshal(body, &response)

	if len(response.Result) != 1 {
		return 0, 0, errors.New("Empty funding rate response")
	}

	funding, _ := strconv.ParseFloat(response.Result[0].FundingRate, 64)
	predicted, _ := strconv.ParseFloat(response.Result[0].PredictedFundingRate, 64)

	return funding, predicted, nil
}

func (b *Bybit) FundingRate() chan [2]float64 {
	log.WithField("venue", "bybit").Info("Polling funding rate")

	ch := make(chan [2]float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			current, predicted, err := b.GetFundingRate()
			if err != nil {
				log.WithField("venue", "bybit").Error(err.Error())
				<-ticker.C
				continue
			}

			log.WithFields(log.Fields{
				"venue":     "bybit",
				"current":   current,
				"predicted": predicted,
			}).Debug("Received funding rate")

			ch <- [2]float64{current, predicted}
			<-ticker.C
		}
	}()

	return ch
}

func (b *Bybit) GetSize() int {
	timestamp := strconv.FormatInt((time.Now().UnixNano() / int64(time.Millisecond)), 10)

	params := map[string]interface{}{
		"api_key":   b.ApiKey,
		"symbol":    "BTCUSD",
		"timestamp": timestamp,
	}

	sign := getSignature(params, b.ApiSecret)

	url := fmt.Sprintf(
		"https://api.bybit.com/v2/private/position/list?api_key=%s&symbol=BTCUSD&timestamp=%s&sign=%s",
		b.ApiKey,
		timestamp,
		sign,
	)

	resp, err := http.Get(url)
	if err != nil {
		log.Error(err.Error())
		return 0
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return 0
	}

	var response positionResponse
	json.Unmarshal(body, &response)

	return response.Result.Size
}

func (b *Bybit) subscribe(channels []string) (*websocket.Conn, error) {
	expires := (time.Now().UnixNano() / int64(time.Millisecond)) + 10000

	signatureInput := fmt.Sprintf("GET/realtime%d", expires)
	h := hmac.New(sha256.New, []byte(b.ApiSecret))
	io.WriteString(h, signatureInput)
	signature := fmt.Sprintf("%x", h.Sum(nil))

	v := url.Values{}
	v.Set("api_key", b.ApiKey)
	v.Set("expires", strconv.FormatInt(expires, 10))
	v.Set("signature", signature)

	u := url.URL{
		Scheme:   "wss",
		Host:     b.websocketHostname(),
		Path:     "/realtime",
		RawQuery: v.Encode()}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return &websocket.Conn{}, err
	}

	command := wsCommand{Op: "subscribe", Args: channels}
	if err = c.WriteJSON(command); err != nil {
		return &websocket.Conn{}, err
	}

	return c, nil
}
