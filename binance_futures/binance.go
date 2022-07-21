package binance_futures

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
)

type BinanceFutures struct {
	ApiKey    string
	ApiSecret string
	Testnet   bool
}

func (b *BinanceFutures) hostname() string {
	if b.Testnet {
		return "testnet.binancefuture.com"
	} else {
		return "dapi.binance.com"
	}
}

func (b *BinanceFutures) sign(s string) string {
	h := hmac.New(sha256.New, []byte(b.ApiSecret))
	io.WriteString(h, string(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (b *BinanceFutures) doRequest(method, path string, values url.Values, sign bool) ([]byte, error) {
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

	if resp.StatusCode != http.StatusOK {
		var response errorResponse
		json.Unmarshal(body, &response)
		return []byte{}, errors.New(response.Msg)
	}

	return body, nil
}

func (b *BinanceFutures) GetBalance() (float64, error) {
	body, err := b.doRequest("GET", "/dapi/v1/balance", url.Values{}, true)
	if err != nil {
		return 0, err
	}

	var response balanceResponse
	json.Unmarshal(body, &response)

	for _, asset := range response {
		if asset.Asset == "BTC" {
			balance, _ := strconv.ParseFloat(asset.Balance, 64)
			return balance, nil
		}
	}

	return 0, nil
}
