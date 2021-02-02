package feed

import "testing"

func TestFeedToString(t *testing.T) {
	tests := map[Feed]string{BTCTHB: "BTCTHB", Bybit: "Bybit"}
	for feed, feedString := range tests {
		if feed.String() != feedString {
			t.Errorf("String() should equal '%s'", feedString)
		}
	}
}

func TestFromStringInvalidFeed(t *testing.T) {
	_, err := FromString("fake")
	if err == nil {
		t.Errorf("Should return an error")
	}
}

func TestFromStringValidFeed(t *testing.T) {
	a, err := FromString("btcthb")
	if err != nil {
		t.Errorf("Should not return an error")
	}
	if a != BTCTHB {
		t.Errorf("Unexpected feed %s", a)
	}
}
