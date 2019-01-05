package indodax

import (
	"github.com/anthonychristian/crypto-arbitrage/orderbook"
)

var idxOrderBook *orderbook.OrderBook

func InitOrderBook() {
	if idxOrderBook == nil {
		idxOrderBook = orderbook.NewOrderBook()
	}
}

func GetOB() *orderbook.OrderBook {
	return idxOrderBook
}
