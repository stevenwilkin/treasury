package binance

import "sync"

type quantity interface {
	remaining(float64) float64
	fill(float64)
	done()
}

type btcQuantity struct {
	btc    float64
	filled float64
	m      sync.Mutex
}

func (q *btcQuantity) remaining(price float64) float64 {
	q.m.Lock()
	defer q.m.Unlock()

	return q.btc - q.filled
}

func (q *btcQuantity) fill(quantity float64) {
	q.m.Lock()
	defer q.m.Unlock()

	q.filled += quantity
}

func (q *btcQuantity) done() {
	q.m.Lock()
	defer q.m.Unlock()

	q.filled = q.btc
}

type usdQuantity struct {
	usd    float64
	filled float64
	isDone bool
	m      sync.Mutex
}

func (q *usdQuantity) remaining(price float64) float64 {
	q.m.Lock()
	defer q.m.Unlock()

	if q.isDone {
		return 0
	} else {
		return (q.usd / price) - q.filled
	}

}
func (q *usdQuantity) fill(quantity float64) {
	q.m.Lock()
	defer q.m.Unlock()

	q.filled += quantity
}

func (q *usdQuantity) done() {
	q.m.Lock()
	defer q.m.Unlock()

	q.isDone = true
}

var _ quantity = &btcQuantity{}
var _ quantity = &usdQuantity{}
