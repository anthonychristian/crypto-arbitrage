package trade

import (
	"fmt"

	"github.com/anthonychristian/crypto-arbitrage/orderbook"
)

type Worker struct {
	ObUpdated chan orderbook.ExchangeKey
	halt      chan bool
	pairs     []Pair
}

var TradeUpdate chan string
var TradeHedge chan string

// InitWorker instances
func InitWorker(pairs []Pair) *Worker {
	TradeUpdate = make(chan string)
	TradeHedge = make(chan string)
	newWorker := &Worker{
		ObUpdated: make(chan orderbook.ExchangeKey),
		halt:      make(chan bool),
		pairs:     pairs,
	}
	return newWorker
}

var TradeWorkers map[orderbook.ExchangeKey][]Worker

func (w *Worker) Start() {
	go w.work()
}

func (w *Worker) Stop() {
	w.halt <- true
}

func (w *Worker) work() {
	for {
		select {
		case _ = <-w.ObUpdated:
			exchanges := copy(orderbook.Exchanges)
			tradeWorker(w.pairs, exchanges)
		case <-w.halt:
			return
		}
	}
}

func tradeWorker(pairs []Pair, exchanges orderbook.ExchangeMap) {
	var totalQty float64
	var totalEndQty float64
	for {
		start, end, notProfitable, err := detectArbitrage(pairs, exchanges)
		if err != nil {
			if notProfitable && totalQty > 0 {
				TradeHedge <- fmt.Sprintln("HEDGE NOW start:", totalQty, "end:", totalEndQty, "totalProfit:", totalEndQty-totalQty)
				TradeUpdate <- fmt.Sprintln("HEDGE NOW start:", totalQty, "end:", totalEndQty, "totalProfit:", totalEndQty-totalQty)
				return
			} else {
				totalQty = 0
				TradeUpdate <- fmt.Sprintln(err)
				return
			}
		} else {
			totalQty += start
			totalEndQty += end
		}
	}
}

func InitEthUsdtIdr() {
	pair1 := Pair{
		symbol:   orderbook.ETH_USDT,
		exchange: orderbook.Binance,
		side:     Buy,
	}
	pair2 := Pair{
		symbol:   orderbook.ETH_IDR,
		exchange: orderbook.Indodax,
		side:     Sell,
	}
	pair3 := Pair{
		symbol:   orderbook.USDT_IDR,
		exchange: orderbook.Indodax,
		side:     Buy,
	}
	worker := InitWorker([]Pair{pair1, pair2, pair3})
	RegisterWorker(worker, orderbook.Binance)
	RegisterWorker(worker, orderbook.Indodax)
	worker.Start()
	fmt.Println("worker started")
}

func RegisterWorker(worker *Worker, exchange orderbook.ExchangeKey) {
	if TradeWorkers == nil {
		TradeWorkers = make(map[orderbook.ExchangeKey][]Worker)
	}
	if TradeWorkers[exchange] == nil {
		TradeWorkers[exchange] = []Worker{}
	}
	TradeWorkers[exchange] = append(TradeWorkers[exchange], *worker)
}

func copy(originalMap orderbook.ExchangeMap) orderbook.ExchangeMap {
	newMap := make(orderbook.ExchangeMap)
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}
