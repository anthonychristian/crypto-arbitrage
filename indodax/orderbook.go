package indodax

import (
	"github.com/anthonychristian/crypto-arbitrage/orderbook"
)

var idxOrderBook *orderbook.OrderBook

func InitOrderBook() {
	if idxOrderBook == nil {
		idxOrderBook = orderbook.Exchanges[orderbook.Indodax].Books[orderbook.ETH_IDR]
	}
}

func GetOB() *orderbook.OrderBook {
	return idxOrderBook
}
