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

func (b *Bybit) signedUrl(path string, addParams map[string]string) string {
	params := map[string]interface{}{
		"api_key":   b.ApiKey,
		"timestamp": b.timestamp()}

	for k, v := range addParams {
		params[k] = v
	}

	keys := make([]string, len(params))
	i := 0
	query := ""
	for k, _ := range params {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		query += fmt.Sprintf("%s=%v&", k, params[k])
	}
	query = query[0 : len(query)-1]
	h := hmac.New(sha256.New, []byte(b.ApiSecret))
	io.WriteString(h, query)
	query += fmt.Sprintf("&sign=%x", h.Sum(nil))

	u := url.URL{
		Scheme:   "https",
		Host:     b.hostname(),
		Path:     path,
		RawQuery: query}

	return u.String()
}

func (b *Bybit) get(path string, params map[string]string, result interface{}) error {
	u := b.signedUrl(path, params)

	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	json.Unmarshal(body, result)

	return nil
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
				log.WithField("venue", "bybit").Warn(err.Error())
				close(ch)
				return
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
	var response positionResponse

	err := b.get("/v2/private/position/list",
		map[string]string{"symbol": "BTCUSD"}, &response)

	if err != nil {
		return 0
	}

	return response.Result.Size
}

func (b *Bybit) GetEquityAndLeverage() (float64, float64, error) {
	var response positionResponse

	err := b.get("/v2/private/position/list",
		map[string]string{"symbol": "BTCUSD"}, &response)

	if err != nil {
		return 0, 0, err
	}

	walletBalance, _ := strconv.ParseFloat(response.Result.WalletBalance, 64)
	positionValue, _ := strconv.ParseFloat(response.Result.PositionValue, 64)

	equity := walletBalance + response.Result.UnrealisedPnl

	if equity == 0 {
		return 0, 0, nil
	}

	return equity, (positionValue / equity), nil
}

func (d *Bybit) EquityAndLeverage() chan [2]float64 {
	log.WithField("venue", "bybit").Info("Polling equity and leverage")

	ch := make(chan [2]float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			equity, leverage, err := d.GetEquityAndLeverage()
			if err != nil {
				log.WithField("venue", "bybit").Warn(err.Error())
				close(ch)
				return
			}

			log.WithFields(log.Fields{
				"venue":    "bybit",
				"equity":   equity,
				"leverage": leverage,
			}).Debug("Received equity and leverage")

			ch <- [2]float64{equity, leverage}
			<-ticker.C
		}
	}()

	return ch
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
