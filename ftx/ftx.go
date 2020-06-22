package ftx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/stevenwilkin/treasury/asset"
)

type FTX struct {
	ApiKey    string
	ApiSecret string
}

type walletResponse struct {
	Result map[string][]struct {
		Coin  string
		Total float64
	}
}

func (f *FTX) Balances() asset.Balances {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	signatureInput := fmt.Sprintf("%dGET/api/wallet/all_balances", timestamp)

	h := hmac.New(sha256.New, []byte(f.ApiSecret))
	io.WriteString(h, string(signatureInput))
	signature := fmt.Sprintf("%x", h.Sum(nil))

	req, err := http.NewRequest("GET", "https://ftx.com/api/wallet/all_balances", nil)
	if err != nil {
		panic(err.Error())
	}

	req.Header.Set("FTX-KEY", f.ApiKey)
	req.Header.Set("FTX-SIGN", signature)
	req.Header.Set("FTX-TS", fmt.Sprintf("%d", timestamp))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var response walletResponse
	json.Unmarshal(body, &response)

	//fmt.Printf("%v\n", response)

	result := asset.Balances{}

	for _, account := range response.Result {
		for _, coin := range account {
			//fmt.Printf("%v\n", coin)
			switch coin.Coin {
			case "BTC":
				result[asset.BTC] += coin.Total
			case "USD":
				result[asset.USD] += coin.Total
			case "USDT":
				result[asset.USDT] += coin.Total
			}
		}
	}

	//fmt.Printf("BTC: %f USD: %f USDT: %f\n", btc, usd, usdt)
	return result
}
