package symbol

import "testing"

func TestSymbolToString(t *testing.T) {
	tests := map[Symbol]string{BTCTHB: "BTCTHB", USDTTHB: "USDTTHB"}
	for symbol, symbolString := range tests {
		if symbol.String() != symbolString {
			t.Errorf("ToString() should equal '%s'", symbolString)
		}
	}
}
