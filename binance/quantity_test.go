package binance

import "testing"

func TestBTCQuantity(t *testing.T) {
	q := &btcQuantity{btc: 2.1}

	if q.remaining(10000) != 2.1 {
		t.Errorf("Expected 2.1 got %f", q.remaining(10000))
	}

	q.fill(0.1)

	if q.remaining(10000) != 2.0 {
		t.Errorf("Expected 2.0 got %f", q.remaining(10000))
	}

	q.done()

	if q.remaining(10000) != 0.0 {
		t.Errorf("Expected 0.0 got %f", q.remaining(10000))
	}
}

func TestUSDQuantity(t *testing.T) {
	q := &usdQuantity{usd: 10000}

	if q.remaining(10000) != 1.0 {
		t.Errorf("Expected 1.0 got %f", q.remaining(10000))
	}

	q.fill(0.1)

	if q.remaining(10000) != 0.9 {
		t.Errorf("Expected 0.9 got %f", q.remaining(10000))
	}

	q.done()

	if q.remaining(10000) != 0.0 {
		t.Errorf("Expected 0.0 got %f", q.remaining(10000))
	}
}
