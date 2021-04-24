package xe

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

type XE struct {
	_accessToken string
	expiresIn    time.Time
}

type rateResponse struct {
	Rates struct {
		THB float64 `json:"THB"`
	} `json:"rates"`
}

func (x *XE) accessToken() (string, error) {
	if x._accessToken != "" && x.expiresIn.After(time.Now()) {
		return x._accessToken, nil
	}

	resp, err := http.Get("https://xe.com/currencyconverter/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`/_next/\S+/_app-\w+\.js`)
	jsUrl := re.Find(body)
	if jsUrl == nil {
		return "", errors.New("Could not find app js")
	}

	resp, err = http.Get(fmt.Sprintf("https://xe.com%s", jsUrl))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re = regexp.MustCompile(`concat\("(\w+)"\)\)\)`)
	matches := re.FindSubmatch(body)
	if matches == nil {
		return "", errors.New("Could not find secret")
	}

	secret := []byte(fmt.Sprintf("lodestar:%s", matches[1]))
	x._accessToken = base64.StdEncoding.EncodeToString(secret)
	x.expiresIn = time.Now().Add(time.Hour)

	return x._accessToken, nil

}

func (x *XE) GetPrice() (float64, error) {
	accessToken, err := x.accessToken()
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("GET", "https://xe.com/api/protected/midmarket-converter/", nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", accessToken))

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
