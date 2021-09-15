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

func (b *Bybit) GetFundingRate() ([2]float64, error) {
	var err error
	defer func() {
		if err != nil {
			log.WithField("venue", "bybit").Warn(err.Error())
		}
	}()

	url := "https://api.bybit.com/v2/public/tickers?symbol=BTCUSD"
	resp, err := http.Get(url)
	if err != nil {
		return [2]float64{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return [2]float64{}, err
	}

	var response fundingResponse
	json.Unmarshal(body, &response)

	if len(response.Result) != 1 {
		err = errors.New("Empty funding rate response")
		return [2]float64{}, err
	}

	funding, _ := strconv.ParseFloat(response.Result[0].FundingRate, 64)
	predicted, _ := strconv.ParseFloat(response.Result[0].PredictedFundingRate, 64)

	return [2]float64{funding, predicted}, nil
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

func (b *Bybit) GetEquityAndLeverage() ([2]float64, error) {
	var response positionResponse

	err := b.get("/v2/private/position/list",
		map[string]string{"symbol": "BTCUSD"}, &response)

	if err != nil {
		log.WithField("venue", "bybit").Warn(err.Error())
		return [2]float64{0, 0}, err
	}

	walletBalance, _ := strconv.ParseFloat(response.Result.WalletBalance, 64)
	positionValue, _ := strconv.ParseFloat(response.Result.PositionValue, 64)

	equity := walletBalance + response.Result.UnrealisedPnl

	if equity == 0 {
		return [2]float64{0, 0}, nil
	}

	return [2]float64{equity, (positionValue / equity)}, nil
}
