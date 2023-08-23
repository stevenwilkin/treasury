package twilio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/stevenwilkin/treasury/alert"
)

type Twilio struct {
	AccountSid string
	AuthToken  string
	From       string
	To         string
}

type errorResponse struct {
	Message string `json:"message"`
}

func (t *Twilio) Notify(_ alert.Alert) error {
	v := url.Values{
		"Twiml": {"<Response><Say>Alert</Say></Response>"},
		"From":  {t.From},
		"To":    {t.To}}

	u := url.URL{
		Scheme: "https",
		Host:   "api.twilio.com",
		Path:   fmt.Sprintf("/2010-04-01/Accounts/%s/Calls.json", t.AccountSid)}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.AccountSid, t.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response errorResponse
	json.Unmarshal(body, &response)

	return errors.New(response.Message)
}

func NewFromEnv() *Twilio {
	return &Twilio{
		AccountSid: os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		From:       os.Getenv("TWILIO_FROM"),
		To:         os.Getenv("TWILIO_TO")}
}

var _ alert.Notifier = &Twilio{}
