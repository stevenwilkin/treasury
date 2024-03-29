package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stevenwilkin/treasury/alert"
	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/feed"
	"github.com/stevenwilkin/treasury/state"
	"github.com/stevenwilkin/treasury/venue"
)

type TestNotifier struct{}

func (t *TestNotifier) Notify(_ alert.Alert) error { return nil }

var (
	s = state.NewState()
	h = NewHandler(s,
		alert.NewAlerter(s, &TestNotifier{}),
		feed.NewHandler(),
		venue.Venues{})
)

func TestSetAssetInvalidVenue(t *testing.T) {
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
	handler := http.HandlerFunc(h.SetAsset)
	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
}

func TestSetAssetInvalidAsset(t *testing.T) {
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
	handler := http.HandlerFunc(h.SetAsset)
	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
}

func TestSetAsset(t *testing.T) {
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
	handler := http.HandlerFunc(h.SetAsset)
	handler.ServeHTTP(w, r)

	if h.s.GetAsset(venue.Nexo, asset.BTC) != 1.23 {
		t.Errorf("Unexpected asset value %f", h.s.Assets[venue.Nexo][asset.BTC])
	}
}

func TestSetCost(t *testing.T) {
	params := url.Values{"cost": {"123.45"}}
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/cost", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(h.SetCost)
	handler.ServeHTTP(w, r)

	if h.s.Cost != 123.45 {
		t.Errorf("Unexpected cost %f", h.s.Cost)
	}
}

func TestAddPriceAlert(t *testing.T) {
	h.a = alert.NewAlerter(s, &TestNotifier{})

	params := url.Values{"value": {"20000"}}
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/alerts/price", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(h.AddPriceAlert)
	handler.ServeHTTP(w, r)

	if len(h.a.Alerts()) != 1 {
		t.Error("Should set an alert")
	}

	alert := h.a.Alerts()[0]
	expected := "Price alert at BTCUSDT 20000.00"

	if alert.Description() != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, alert.Description())
	}
}

func TestAddFundingAlert(t *testing.T) {
	h.a = alert.NewAlerter(s, &TestNotifier{})

	r, err := http.NewRequest("POST", "/alerts/funding", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(h.AddFundingAlert)
	handler.ServeHTTP(w, r)

	if len(h.a.Alerts()) != 1 {
		t.Error("Should set an alert")
	}

	alert := h.a.Alerts()[0]
	expected := "Negative funding alert"

	if alert.Description() != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, alert.Description())
	}
}

func TestAddLeverageAlert(t *testing.T) {
	h.a = alert.NewAlerter(s, &TestNotifier{})

	params := url.Values{"value": {"4"}}
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/alerts/leverage", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(h.AddLeverageAlert)
	handler.ServeHTTP(w, r)

	if len(h.a.Alerts()) != 1 {
		t.Fatal("Should set an alert")
	}

	alert := h.a.Alerts()[0]
	expected := "Leverage alert at 4.00"

	if alert.Description() != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, alert.Description())
	}
}

func TestReactivateFeedInvalidFeed(t *testing.T) {
	params := url.Values{}
	params.Set("feed", "fake")
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/feeds/reactivate", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(h.ReactivateFeed)
	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
}

func TestReactivateFeed(t *testing.T) {
	params := url.Values{}
	params.Set("feed", feed.BTCUSDT.String())
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/feeds/reactivate", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(h.ReactivateFeed)
	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
}

func TestSetLoan(t *testing.T) {
	params := url.Values{"loan": {"123.45"}}
	body := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", "/loan/set", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(h.SetLoan)
	handler.ServeHTTP(w, r)

	if h.s.GetLoan() != 123.45 {
		t.Errorf("Unexpected loan %f", h.s.GetLoan())
	}
}
