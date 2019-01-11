package indodax

import (
	"github.com/anthonychristian/crypto-arbitrage/orderbook"
)

var idxOrderBooks map[orderbook.Symbol]*orderbook.OrderBook

func InitOrderBook() {
	idxOrderBooks = make(map[orderbook.Symbol]*orderbook.OrderBook)
	idxOrderBooks[orderbook.ETH_IDR] = orderbook.Exchanges[orderbook.Indodax].Books[orderbook.ETH_IDR]
	idxOrderBooks[orderbook.USDT_IDR] = orderbook.Exchanges[orderbook.Indodax].Books[orderbook.USDT_IDR]
}

func GetOB(symbol orderbook.Symbol) *orderbook.OrderBook {
	return idxOrderBooks[symbol]
}
