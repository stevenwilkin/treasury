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

func (f *FTX) sign(s string) string {
	h := hmac.New(sha256.New, []byte(f.ApiSecret))
	io.WriteString(h, string(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (f *FTX) GetBalances() ([2]float64, error) {
	var err error
	defer func() {
		if err != nil {
			log.WithField("venue", "ftx").Warn(err.Error())
		}
	}()

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	signatureInput := fmt.Sprintf("%dGET/api/wallet/all_balances", timestamp)

	req, err := http.NewRequest("GET", "https://ftx.com/api/wallet/all_balances", nil)
	if err != nil {
		return [2]float64{}, err
	}

	req.Header.Set("FTX-KEY", f.ApiKey)
	req.Header.Set("FTX-SIGN", f.sign(signatureInput))
	req.Header.Set("FTX-TS", fmt.Sprintf("%d", timestamp))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return [2]float64{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return [2]float64{}, err
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

	return [2]float64{btc, usdt}, nil
}
