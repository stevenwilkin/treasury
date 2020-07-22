package venue

import "strings"

type Venue int

const (
	FTX Venue = iota
	Nexo
	Blockfi
	Hodlnaut
)

func venues() []string {
	return []string{"FTX", "Nexo", "Blockfi", "Hodlnaut"}
}

func (v Venue) String() string {
	return venues()[v]
}

func Exists(name string) bool {
	for _, venue := range venues() {
		if strings.ToLower(name) == strings.ToLower(venue) {
			return true
		}
	}
	return false
}
