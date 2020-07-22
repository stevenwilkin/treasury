package venue

import "testing"

func TestVenueToString(t *testing.T) {
	tests := map[Venue]string{Nexo: "Nexo", FTX: "FTX"}

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
