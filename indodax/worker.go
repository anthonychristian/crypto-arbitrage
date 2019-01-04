// Package indodax currently support only 1 coin type(eth)
package indodax

import (
	"strconv"

	"github.com/albertputrapurnama/arbitrage/orderbook"
)

// Worker is the main engine for making order decisions
// and continuing arbitrage loop ETH -> IDR -> USDT
type Worker struct {
	depth chan Depth
	halt  bool
}

var WorkerInstance *Worker

// InitWorker instances
func InitWorker() *Worker {
	newWorker := &Worker{
		depth: make(chan Depth),
	}
	WorkerInstance = newWorker
	go WorkerInstance.work()
	return newWorker
}

// Halt to halt the worker from doing actions
func (w *Worker) Halt() {
	w.halt = true
}

// Start to start the worker to do actions
func (w *Worker) Start() {
	w.halt = false
}

func (w *Worker) PushDepthUpdate(d Depth) {
	w.depth <- d
}

func (w *Worker) work() {
	// infinite loop to keep doing actions
	for {
		select {
		case d := <-w.depth:
			// add depth to orderbook
			updateDepth(d)
		}
	}
}

func updateDepth(d Depth) {
	// TODO update the indodax's orderbook (ETHEREUM)
	for _, elem := range d.Buy {
		q, err := strconv.ParseFloat(elem[1].(string), 64)
		if err != nil {
			panic(err)
		}
		newOrder := orderbook.Order{
			Price:       elem[0].(float64),
			Qty:         q,
			ExchangeKey: orderbook.Indodax,
		}
		idxOrderBook.AddBuy(newOrder)
	}
	for _, elem := range d.Sell {
		q, err := strconv.ParseFloat(elem[1].(string), 64)
		if err != nil {
			panic(err)
		}
		newOrder := orderbook.Order{
			Price:       elem[0].(float64),
			Qty:         q,
			ExchangeKey: orderbook.Indodax,
		}
		idxOrderBook.AddSell(newOrder)
	}
}
