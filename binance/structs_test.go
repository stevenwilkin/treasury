package binance

import "testing"

func TestAssetBalanceTotal(t *testing.T) {
	ab := assetBalance{Asset: "foo", Free: "10", Locked: "5"}

	if ab.Total() != 15.0 {
		t.Errorf("Expected 15.0 got %f", ab.Total())
	}
}
