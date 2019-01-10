package indodax

import (
	"github.com/anthonychristian/crypto-arbitrage/orderbook"
)

var idxOrderBook *orderbook.OrderBook

func InitOrderBook() {
<<<<<<< HEAD
	if idxOrderBook == nil {
		idxOrderBook = orderbook.Exchanges[orderbook.Indodax].Books[orderbook.ETH_IDR]
	}
=======
	idxOrderBook = orderbook.Exchanges[orderbook.Indodax].Books[orderbook.ETH_IDR]
>>>>>>> 2c685a0978bb4d3d7447eaffb68a840f929e8fd5
}

func GetOB() *orderbook.OrderBook {
	return idxOrderBook
}
