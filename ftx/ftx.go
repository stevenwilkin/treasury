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

	log "github.com/sirupsen/logrus"
)

type FTX struct {
	ApiKey    string
	ApiSecret string
}

func (f *FTX) GetBalances() [2]float64 {
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

	var btc, usdt float64

	for _, account := range response.Result {
		for _, coin := range account {
			switch coin.Coin {
			case "BTC":
				btc += coin.Total
			case "USDT":
				usdt += coin.Total
			}
		}
	}

	return [2]float64{btc, usdt}
}

func (f *FTX) Balances() chan [2]float64 {
	log.WithField("venue", "ftx").Info("Polling balances")

	ch := make(chan [2]float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			balances := f.GetBalances()
			log.WithFields(log.Fields{
				"venue": "ftx",
				"btc":   balances[0],
				"usdt":  balances[1],
			}).Debug("Received balances")

			ch <- balances
			<-ticker.C
		}
	}()

	return ch
}
