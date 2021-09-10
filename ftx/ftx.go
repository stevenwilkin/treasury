package ftx

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
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

func (f *FTX) PlaceOrder(size, price float64, buy bool) (int64, error) {
	log.WithFields(log.Fields{
		"venue":  "ftx",
		"market": "BTC/USDT",
		"size":   size,
		"price":  price,
		"buy":    buy,
	}).Info("Placing order")

	side := "buy"
	if !buy {
		side = "sell"
	}

	or := orderRequest{
		Market:   "BTC/USDT",
		Side:     side,
		Size:     size,
		Price:    price,
		Type:     "limit",
		PostOnly: true}

	jsonRequest, err := json.Marshal(or)
	if err != nil {
		return 0, err
	}

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	signatureInput := fmt.Sprintf("%dPOST/api/orders%s", timestamp, jsonRequest)

	req, err := http.NewRequest(
		"POST", "https://ftx.com/api/orders", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("FTX-KEY", f.ApiKey)
	req.Header.Set("FTX-SIGN", f.sign(signatureInput))
	req.Header.Set("FTX-TS", fmt.Sprintf("%d", timestamp))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response orderResponse
	json.Unmarshal(body, &response)

	if response.Success {
		return response.Result.Id, nil
	}

	return 0, nil
}

func (f *FTX) EditOrder(id int64, size, price float64) (int64, error) {
	log.WithFields(log.Fields{
		"venue": "ftx",
		"order": id,
		"size":  size,
		"price": price,
	}).Info("Updating order")

	or := editOrderRequest{
		Size:  size,
		Price: price}

	jsonRequest, err := json.Marshal(or)
	if err != nil {
		return 0, err
	}

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	path := fmt.Sprintf("/api/orders/%d/modify", id)
	signatureInput := fmt.Sprintf("%dPOST%s%s", timestamp, path, jsonRequest)

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://ftx.com%s", path),
		bytes.NewBuffer(jsonRequest))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("FTX-KEY", f.ApiKey)
	req.Header.Set("FTX-SIGN", f.sign(signatureInput))
	req.Header.Set("FTX-TS", fmt.Sprintf("%d", timestamp))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response orderResponse
	json.Unmarshal(body, &response)

	if response.Success {
		return response.Result.Id, nil
	}

	return 0, nil
}

func (f *FTX) Trade(size float64, buy bool) {
	var bestPrice, price float64
	var orderId int64

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	socketUrl := url.URL{Scheme: "wss", Host: "ftx.com", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(socketUrl.String(), nil)
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer c.Close()

	request := opMessage{
		Args: map[string]interface{}{
			"key":  f.ApiKey,
			"sign": f.sign(fmt.Sprintf("%dwebsocket_login", timestamp)),
			"time": timestamp},
		Op: "login"}
	if err = c.WriteJSON(request); err != nil {
		log.Error(err.Error())
		return
	}

	subscribe := []byte(`{"op":"subscribe","channel":"ticker","market":"BTC/USDT"}`)
	if err = c.WriteMessage(websocket.TextMessage, subscribe); err != nil {
		log.Error(err.Error())
		return
	}

	subscribe = []byte(`{"op":"subscribe","channel":"orders"}`)
	if err = c.WriteMessage(websocket.TextMessage, subscribe); err != nil {
		log.Error(err.Error())
		return
	}

	for {
		var message tradeMessage
		if err = c.ReadJSON(&message); err != nil {
			log.Error(err)
			return
		}

		if message.Type != "update" {
			continue
		}

		if message.Channel == "ticker" {
			if buy {
				bestPrice = message.Data.Bid
			} else {
				bestPrice = message.Data.Ask
			}

			if orderId == 0 {
				price = bestPrice
				orderId, err = f.PlaceOrder(size, price, buy)
				if err != nil {
					log.Error(err.Error())
					return
				}
			} else if price != bestPrice {
				price = bestPrice
				_, err = f.EditOrder(orderId, size, price)
				if err != nil {
					log.Error(err.Error())
					return
				}
			}
		} else if message.Channel == "orders" {
			switch message.Data.Status {
			case "new":
				orderId = message.Data.Id
			case "closed":
				if message.Data.FilledSize == message.Data.Size {
					log.WithFields(log.Fields{
						"venue": "ftx",
						"order": orderId,
					}).Info("Order filled")
					return
				}
			}
		}
	}
}
