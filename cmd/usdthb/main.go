package main

import (
	"fmt"

	"github.com/stevenwilkin/treasury/xe"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.PanicLevel)

	xe := &xe.XE{}
	usdThb, _ := xe.GetPrice()
	fmt.Println(usdThb)
}
