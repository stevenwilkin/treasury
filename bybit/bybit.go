package bybit

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

func (b *Bybit) GetEquity() float64 {
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
				"asset": "BTC",
				"value": equity,
			}).Debug("Received equity")

			ch <- equity
			<-ticker.C
		}
	}()

	return ch
}

func (b *Bybit) GetFundingRate() (float64, float64) {
	url := "https://api.bybit.com/v2/public/tickers?symbol=BTCUSD"
	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}

	var response fundingResponse
	json.Unmarshal(body, &response)

	if len(response.Result) != 1 {
		log.WithField("venue", "bybit").Error("Empty funding rate response")
		return 0, 0
	}

	funding, _ := strconv.ParseFloat(response.Result[0].FundingRate, 64)
	predicted, _ := strconv.ParseFloat(response.Result[0].PredictedFundingRate, 64)

	return funding, predicted
}

func (b *Bybit) FundingRate() chan [2]float64 {
	log.WithField("venue", "bybit").Info("Polling funding rate")

	ch := make(chan [2]float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			current, predicted := b.GetFundingRate()
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
		log.Panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}

	var response positionResponse
	json.Unmarshal(body, &response)

	return response.Result.Size
}

func (b *Bybit) subscribe(channels []string) *websocket.Conn {
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
		log.Panic(err.Error())
	}

	command := wsCommand{Op: "subscribe", Args: channels}
	if err = c.WriteJSON(command); err != nil {
		log.Panic(err.Error())
	}

	return c
}

func (b *Bybit) orderRequest(params map[string]interface{}, path string) string {
	params["api_key"] = b.ApiKey
	params["timestamp"] = b.timestamp()
	params["sign"] = getSignature(params, b.ApiSecret)

	jsonRequest, err := json.Marshal(params)
	if err != nil {
		log.Panic(err.Error())
	}

	u := url.URL{
		Scheme: "https",
		Host:   b.hostname(),
		Path:   path}

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonRequest))
	if err != nil {
		log.Panic(err.Error())
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}

	var response orderResponse
	json.Unmarshal(body, &response)

	return response.Result.OrderId
}

func (b *Bybit) PlaceOrder(amount int, price float64, buy, reduce bool) string {
	log.WithFields(log.Fields{
		"venue":  "bybit",
		"amount": amount,
		"price":  price,
	}).Info("Placing order")

	params := map[string]interface{}{
		"symbol":        "BTCUSD",
		"order_type":    "Limit",
		"qty":           strconv.Itoa(amount),
		"price":         strconv.FormatFloat(price, 'f', 2, 64),
		"time_in_force": "PostOnly"}

	if buy {
		params["side"] = "Buy"
	} else {
		params["side"] = "Sell"
	}

	if reduce {
		params["reduce_only"] = true
	}

	return b.orderRequest(params, "/v2/private/order/create")
}

func (b *Bybit) EditOrder(id string, price float64) string {
	log.WithFields(log.Fields{
		"venue": "bybit",
		"order": id,
		"price": price,
	}).Debug("Updating order")

	params := map[string]interface{}{
		"order_id":  id,
		"symbol":    "BTCUSD",
		"p_r_price": strconv.FormatFloat(price, 'f', 2, 64)}

	return b.orderRequest(params, "/v2/private/order/replace")
}

func highest(orders map[int64]float64) float64 {
	var result float64

	for _, x := range orders {
		if x > result {
			result = x
		}
	}

	return result
}

func lowest(orders map[int64]float64) float64 {
	var result float64

	for _, x := range orders {
		if result == 0.0 {
			result = x
		} else if x < result {
			result = x
		}
	}

	return result
}

func (b *Bybit) Trade(contracts int, buy, reduce bool) {
	var bestPrice, price float64
	var orderId string
	var pendingInitialOrder bool

	bids := map[int64]float64{}
	asks := map[int64]float64{}

	orderBookTopic := "orderBookL2_25.BTCUSD"
	orderTopic := "order"

	c := b.subscribe([]string{orderBookTopic, orderTopic})
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Error(err)
			return
		}

		var response wsResponse
		json.Unmarshal(message, &response)

		switch response.Topic {
		case orderTopic:
			var orders []orderTopicData
			json.Unmarshal(response.Data, &orders)
			order := orders[0]

			switch order.OrderStatus {
			case "New":
				orderId = order.OrderId
			case "PartiallyFilled":
				log.WithFields(log.Fields{
					"venue":        "bybit",
					"order":        orderId,
					"cum_quantity": order.CumExecQty,
				}).Debug("Fill")
			case "Filled":
				log.WithFields(log.Fields{
					"venue": "bybit",
					"order": orderId,
				}).Info("Order filled")
				return
			case "Cancelled":
				log.WithFields(log.Fields{
					"venue":        "bybit",
					"order":        orderId,
					"quantity":     order.Qty,
					"cum_quantity": order.CumExecQty,
				}).Debug("Order cancelled")
				orderId = ""
				contracts = order.Qty - order.CumExecQty
				pendingInitialOrder = false
			default:
				pendingInitialOrder = false
			}
		case orderBookTopic:
			switch response.Type {
			case "snapshot":
				var snapshot snapshotData
				json.Unmarshal(response.Data, &snapshot)

				for _, order := range snapshot {
					p, _ := strconv.ParseFloat(order.Price, 64)

					if order.Side == "Buy" {
						bids[order.Id] = p
					} else {
						asks[order.Id] = p
					}
				}
			case "delta":
				var updates updateData
				json.Unmarshal(response.Data, &updates)

				for _, order := range updates.Delete {
					if order.Side == "Buy" {
						delete(bids, order.Id)
					} else {
						delete(asks, order.Id)
					}
				}

				for _, order := range updates.Insert {
					p, _ := strconv.ParseFloat(order.Price, 64)

					if order.Side == "Buy" {
						bids[order.Id] = p
					} else {
						asks[order.Id] = p
					}
				}
			}

			if buy {
				bestPrice = highest(bids)
			} else {
				bestPrice = lowest(asks)
			}

			if orderId == "" {
				if !pendingInitialOrder {
					price = bestPrice
					pendingInitialOrder = true
					b.PlaceOrder(contracts, price, buy, reduce)
				}
			} else if price != bestPrice {
				price = bestPrice
				b.EditOrder(orderId, price)
			}
		}
	}
}
