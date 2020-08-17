package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stevenwilkin/treasury/alert"
	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/state"
	"github.com/stevenwilkin/treasury/venue"
)

type TestNotifier struct{}

func (t *TestNotifier) Notify(s string) bool { return true }

func TestSetHandlerInvalidVenue(t *testing.T) {
	params := url.Values{}
	params.Set("venue", "fake")
	params.Set("asset", "btc")
	params.Set("quantity", "1.23")
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/set", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(setHandler)
	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
}

func TestSetHandlerInvalidAsset(t *testing.T) {
	params := url.Values{}
	params.Set("venue", "nexo")
	params.Set("asset", "fake")
	params.Set("quantity", "1.23")
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/set", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(setHandler)
	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
}

func TestSetHandlerSetsAsset(t *testing.T) {
	statum = state.NewState()

	params := url.Values{}
	params.Set("venue", "nexo")
	params.Set("asset", "btc")
	params.Set("quantity", "1.23")
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/set", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(setHandler)
	handler.ServeHTTP(w, r)

	if statum.Asset(venue.Nexo, asset.BTC) != 1.23 {
		t.Errorf("Unexpected asset value %f", statum.Assets[venue.Nexo][asset.BTC])
	}
}

func TestCostHandler(t *testing.T) {
	params := url.Values{"cost": {"123.45"}}
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/cost", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(costHandler)
	handler.ServeHTTP(w, r)

	if statum.Cost != 123.45 {
		t.Errorf("Unexpected cost %f", statum.Cost)
	}
}

func TestPriceAlertsHandler(t *testing.T) {
	alerter = alert.NewAlerter(&TestNotifier{})

	params := url.Values{"value": {"20000"}}
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/alerts/price", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(priceAlertsHandler)
	handler.ServeHTTP(w, r)

	if len(alerter.Alerts()) != 1 {
		t.Error("Should set an alert")
	}

	alert := alerter.Alerts()[0]
	expected := "Price alert at BTCUSDT 20000.00"

	if alert.Description() != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, alert.Description())
	}
}
