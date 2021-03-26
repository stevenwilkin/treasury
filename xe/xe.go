package xe

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type XE struct{}

type rateResponse struct {
	Rates struct {
		THB float64 `json:"THB"`
	} `json:"rates"`
}

func (x *XE) GetPrice() (float64, error) {
	req, err := http.NewRequest("GET", "https://xe.com/api/protected/midmarket-converter/", nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Basic bG9kZXN0YXI6eDRBZE9MaENEbHQ3TkNLV25sTlhIUXlQTzMzZVo0R00=")

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

	var response rateResponse
	json.Unmarshal(body, &response)

	if response.Rates.THB == 0 {
		return 0, errors.New("Empty rate response")
	}

	return response.Rates.THB, nil
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
