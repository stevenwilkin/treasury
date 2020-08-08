package bybit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type Bybit struct {
	ApiKey    string
	ApiSecret string
}

type equityResponse struct {
	Result struct {
		BTC struct {
			Equity float64
		}
	}
}

func getSignature(params map[string]string, key string) string {
	keys := make([]string, len(params))
	i := 0
	_val := ""
	for k, _ := range params {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		_val += k + "=" + params[k] + "&"
	}
	_val = _val[0 : len(_val)-1]
	h := hmac.New(sha256.New, []byte(key))
	io.WriteString(h, _val)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (b *Bybit) GetEquity() float64 {
	timestamp := strconv.FormatInt((time.Now().UnixNano() / int64(time.Millisecond)), 10)

	params := map[string]string{
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
		log.Panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}

	var response equityResponse
	json.Unmarshal(body, &response)

	return response.Result.BTC.Equity
}

func (b *Bybit) Equity() chan float64 {
	log.WithField("venue", "bybit").Info("Polling equity")

	ch := make(chan float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			equity := b.GetEquity()
			log.WithFields(log.Fields{
				"venue": "bybit",
				"value": equity,
			}).Debug("Received equity")

			ch <- equity
			<-ticker.C
		}
	}()

	return ch
}
