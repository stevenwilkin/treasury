package asset

import "testing"

func TestToString(t *testing.T) {
	tests := map[Asset]string{BTC: "BTC", USD: "USD"}

	for asset, assetString := range tests {
		if asset.String() != assetString {
			t.Errorf("ToString() should equal '%s'", assetString)
		}
	}
}

func TestExists(t *testing.T) {
	tests := map[string]bool{"USDT": true, "usdt": true, "Fake": false}

	for name, exists := range tests {
		if Exists(name) != exists {
			t.Errorf("Exists(\"%s\") should equal %t", name, exists)
		}
	}
}

func TestFromStringInvalidAsset(t *testing.T) {
	_, err := FromString("fake")
	if err == nil {
		t.Errorf("Should return an error")
	}
}

func TestFromStringValidAsset(t *testing.T) {
	a, err := FromString("btc")
	if err != nil {
		t.Errorf("Should not return an error")
	}
	if a != BTC {
		t.Errorf("Unexpected asset %s", a)
	}
}
