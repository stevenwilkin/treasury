package xe

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type XE struct{}

type rateResponse struct {
	Payload struct {
		Rates struct {
			Rate string `json:"rate"`
		} `json:"rates"`
	} `json:"payload"`
}

func extractDataAndKey(s string) (string, string) {
	var offset int
	var n float64

	for _, c := range s[len(s)-4:] {
		n += float64(c)
	}

	n = math.Mod((float64(len(s)) - 10), n)
	if n > float64(len(s))-14 {
		offset = len(s) - 14
	} else {
		offset = int(n)
	}

	return s[0:offset] + s[offset+10:], s[offset : offset+10]
}

func decodeString(data []byte, key string) string {
	var i, keyPosition float64
	var result, plaintext string

	for o := 0; o < len(data); o += 10 {
		charCode := data[o]

		if math.Mod(float64(i), float64(len(key)))-1 < 0 {
			keyPosition = float64(len(key)) + math.Mod(i, float64(len(key))) - 1
		} else {
			keyPosition = math.Mod(i, float64(len(key))) - 1
		}

		if (o + 10) > len(data) {
			plaintext = string(data[o+1 : len(data)])
		} else {
			plaintext = string(data[o+1 : o+10])
		}

		result += string(charCode-key[int(keyPosition)]) + plaintext
		i++
	}

	return result
}

func decode(s string) (float64, error) {
	dataString, key := extractDataAndKey(s)

	dataUnescaped, err := url.QueryUnescape(dataString)
	if err != nil {
		return 0, err
	}

	data, err := base64.StdEncoding.DecodeString(dataUnescaped)
	if err != nil {
		return 0, err
	}

	resultString := decodeString(data, key)

	if result, err := strconv.ParseFloat(resultString, 64); err != nil {
		return 0, err
	} else {
		return result, nil
	}
}

func (x *XE) GetPrice() (float64, error) {
	resp, err := http.Get("https://www.xe.com/api/page_resources/converter.php?fromCurrency=USD&toCurrency=THB")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response rateResponse
	json.Unmarshal(body, &response)

	return decode(response.Payload.Rates.Rate)
}

func (x *XE) Price() chan float64 {
	log.WithFields(log.Fields{
		"venue":  "xe",
		"symbol": "USDTHB",
	}).Info("Polling price")

	ch := make(chan float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			price, err := x.GetPrice()
			if err != nil {
				log.WithField("venue", "xe").Warn(err.Error())
				close(ch)
				return
			}

			log.WithFields(log.Fields{
				"venue":  "xe",
				"symbol": "USDTHB",
				"value":  price,
			}).Debug("Received price")

			ch <- price
			<-ticker.C
		}
	}()

	return ch
}
