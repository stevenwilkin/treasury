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

func TestFromStringInvalidSymbol(t *testing.T) {
	_, err := FromString("fake")
	if err == nil {
		t.Errorf("Should return an error")
	}
}

func TestFromStringValidSymbol(t *testing.T) {
	a, err := FromString("btcthb")
	if err != nil {
		t.Errorf("Should not return an error")
	}
	if a != BTCTHB {
		t.Errorf("Unexpected symbol %s", a)
	}
}
