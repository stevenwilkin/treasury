package main

import (
	"log"
	"net/http"

	"github.com/stevenwilkin/treasury/bitkub"
	"github.com/stevenwilkin/treasury/symbol"

	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
)

type pricesMessage struct {
	Prices map[string]float64 `json:"prices"`
}

var (
	bitkubExchange = &bitkub.BitKub{}
	conns          = map[*websocket.Conn]bool{}
	prices         = symbol.Prices{}
	upgrader       = websocket.Upgrader{}
)

func sendState(c *websocket.Conn) {
	log.Println("Sending initial state")

	pm := pricesMessage{Prices: map[string]float64{}}
	for s, p := range prices {
		//log.Printf("%s - %f\n", s, p)
		pm.Prices[s.String()] = p
	}
	log.Printf("%v\n", pm)

	err := c.WriteJSON(pm)
	if err != nil {
		log.Println("write:", err)
	}
}

func updatePrice(s symbol.Symbol, price float64) {
	log.Printf("updatePrice - %s - %f\n", s, price)
	prices[s] = price

	for c, _ := range conns {
		pm := pricesMessage{Prices: map[string]float64{s.String(): price}}

		err := c.WriteJSON(pm)
		if err != nil {
			log.Println("write:", err)
			delete(conns, c)
		}
	}
}

func initPriceFeeds() {
	log.Println("Initialising price feeds")

	btcThbPrices := make(chan float64, 1)
	usdtThbPrices := make(chan float64, 1)
	go bitkubExchange.Price(symbol.BTCTHB, btcThbPrices)
	go bitkubExchange.Price(symbol.USDTTHB, usdtThbPrices)

	for {
		select {
		case btcThb := <-btcThbPrices:
			updatePrice(symbol.BTCTHB, btcThb)
		case usdtThb := <-usdtThbPrices:
			updatePrice(symbol.USDTTHB, usdtThb)
		}
	}
}

func initPrices() {
	log.Println("Initialising prices")
	prices[symbol.BTCTHB] = 0
	prices[symbol.USDTTHB] = 0
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.Println("Accepting connection")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	//defer c.Close()
	conns[c] = true
	sendState(c)
}

func initWeb() {
	fs := http.FileServer(http.Dir("./www"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", serveWs)
	log.Println("Listening on 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	initPrices()
	go initPriceFeeds()
	initWeb()
}
