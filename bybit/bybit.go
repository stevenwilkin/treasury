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

func (b *Bybit) get(path string, params url.Values, result interface{}) error {
	query := params.Encode()
	timestamp := b.timestamp()

	u := url.URL{
		Scheme:   "https",
		Host:     b.hostname(),
		Path:     path,
		RawQuery: query}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	h := hmac.New(sha256.New, []byte(b.ApiSecret))
	io.WriteString(h, timestamp+b.ApiKey+query)
	signature := fmt.Sprintf("%x", h.Sum(nil))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BAPI-API-KEY", b.ApiKey)
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)

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

	json.Unmarshal(body, result)

	return nil
}

func (b *Bybit) GetFundingRate() (float64, error) {
	var err error
	defer func() {
		if err != nil {
			log.WithField("venue", "bybit").Warn(err.Error())
		}
	}()

	var response fundingResponse

	err = b.get("/v5/market/tickers",
		url.Values{"category": {"inverse"}, "symbol": {"BTCUSD"}}, &response)
	if err != nil {
		return 0, err
	}

	if len(response.Result.List) != 1 {
		err = errors.New("Empty funding rate response")
		return 0, err
	}

	funding, _ := strconv.ParseFloat(response.Result.List[0].FundingRate, 64)

	return funding, nil
}

func (b *Bybit) positionRequest() (positionResponse, error) {
	var response positionResponse

	err := b.get("/v5/position/list",
		url.Values{"category": {"inverse"}, "symbol": {"BTCUSD"}}, &response)

	if err != nil {
		return response, err
	}

	if len(response.Result.List) != 1 {
		return response, errors.New("Unexpected position response")
	}

	return response, nil
}

func (b *Bybit) GetSize() int {
	response, err := b.positionRequest()
	if err != nil {
		return 0
	}

	size, _ := strconv.Atoi(response.Result.List[0].Size)
	return size
}

func (b *Bybit) GetEquityAndLeverage() ([2]float64, error) {
	var response walletResponse

	err := b.get("/v5/account/wallet-balance",
		url.Values{"accountType": {"CONTRACT"}, "coin": {"BTC"}}, &response)

	if err != nil {
		return [2]float64{}, err
	}

	if len(response.Result.List) != 1 {
		return [2]float64{}, errors.New("Unexpected wallet response")
	}

	if len(response.Result.List[0].Coin) != 1 {
		return [2]float64{}, errors.New("Unexpected coin response")
	}

	equity, _ := strconv.ParseFloat(response.Result.List[0].Coin[0].Equity, 64)
	if equity == 0 {
		return [2]float64{}, nil
	}

	resp, err := b.positionRequest()
	if err != nil {
		log.WithField("venue", "bybit").Warn(err.Error())
		return [2]float64{equity, 0}, nil
	}

	positionValue, _ := strconv.ParseFloat(resp.Result.List[0].PositionValue, 64)

	return [2]float64{equity, (positionValue / equity)}, nil
}
