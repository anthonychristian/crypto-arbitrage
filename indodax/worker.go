// Package indodax currently support only 1 coin type(eth)
package indodax

import (
	"strconv"
	"strings"
	"time"

	"github.com/anthonychristian/crypto-arbitrage/orderbook"
)

// Worker is the main engine for making order decisions
// and continuing arbitrage loop ETH -> IDR -> USDT
type Worker struct {
	depth  chan Depth
	halt   bool
	symbol orderbook.Symbol
}

var IndodaxWorkers map[orderbook.Symbol]*Worker

func InitAllWorkers() {
	IndodaxWorkers = make(map[orderbook.Symbol]*Worker)
	for _, symbol := range orderbook.ExSymbolMap[orderbook.Indodax] {
		worker := InitWorker(symbol)
		go updateDepthToWorker(worker)
	}
}

// InitWorker instances
func InitWorker(symbol orderbook.Symbol) *Worker {
	newWorker := &Worker{
		depth:  make(chan Depth),
		symbol: symbol,
	}
	IndodaxWorkers[symbol] = newWorker
	go IndodaxWorkers[symbol].work()
	return IndodaxWorkers[symbol]
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
			updateDepth(d, w.symbol)
		}
	}
}

func updateDepth(d Depth, symbol orderbook.Symbol) {
	// TODO update the indodax's orderbook (ETHEREUM)
	for _, elem := range d.Buy {
		q, err := strconv.ParseFloat(elem[1].(string), 64)
		if err != nil {
			panic(err)
		}
		newOrder := orderbook.Order{
			Price: elem[0].(float64),
			Qty:   q,
		}
		orderbook.Exchanges[orderbook.Indodax].Books[symbol].AddBuy(newOrder)
	}
	for _, elem := range d.Sell {
		q, err := strconv.ParseFloat(elem[1].(string), 64)
		if err != nil {
			panic(err)
		}
		newOrder := orderbook.Order{
			Price: elem[0].(float64),
			Qty:   q,
		}
		orderbook.Exchanges[orderbook.Indodax].Books[symbol].AddSell(newOrder)
	}
}

func updateDepthToWorker(worker *Worker) {
	symbolString := strings.ToLower(string(orderbook.GetLeftCurrency(worker.symbol)) + "_" + string(orderbook.GetRightCurrency(worker.symbol)))
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			d := IndodaxInstance.GetDepth(symbolString)
			worker.PushDepthUpdate(d)
		}
	}()
}
