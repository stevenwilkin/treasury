package venue

import "testing"

func TestVenueToString(t *testing.T) {
	tests := map[Venue]string{Nexo: "Nexo", Binance: "Binance"}

	for venue, venueString := range tests {
		if venue.String() != venueString {
			t.Errorf("ToString() should equal '%s'", venueString)
		}
	}
}

func TestExists(t *testing.T) {
	tests := map[string]bool{"NEXO": true, "nexo": true, "Fake": false}

	for name, exists := range tests {
		if Exists(name) != exists {
			t.Errorf("Exists(\"%s\") should equal %t", name, exists)
		}
	}
}

func TestFromStringInvalidVenue(t *testing.T) {
	_, err := FromString("fake")
	if err == nil {
		t.Errorf("Should return an error")
	}
}

func TestFromStringValidVenue(t *testing.T) {
	v, err := FromString("nexo")
	if err != nil {
		t.Errorf("Should not return an error")
	}
	if v != Nexo {
		t.Errorf("Unexpected venue %s", v)
	}
}
