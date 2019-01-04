package orderbook

import (
	"github.com/anthonychristian/crypto-arbitrage/skiplist"
	"github.com/shopspring/decimal"
)

type Order struct {
	Price float64 // Actual price on exchange
	Qty   float64
}

type OrderBookMap map[Symbol]*OrderBook // Key is the currency pair, e.g. BTC/USDC

type OrderBook struct {
	buyside, sellside *skiplist.SkipList
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		skiplist.NewDecimalMapReverse(),
		skiplist.NewDecimalMap(),
	}
}

func (ob *OrderBook) AddBuy(order Order) {
	add(order, ob.buyside)
}

func (ob *OrderBook) AddSell(order Order) {
	add(order, ob.sellside)
}

func add(order Order, book *skiplist.SkipList) {
	priceKey := decimal.NewFromFloat(order.Price)
	if _, ok := book.Get(priceKey); ok { // Existing price level, append order
		if order.Qty == 0 {
			book.Delete(priceKey)
			return
		}
		// ol := val.(Order)
		// order.Qty = ol.Qty + order.Qty
		book.Set(priceKey, order)
	} else if order.Qty != 0 {
		book.Set(priceKey, order) // New price level
	}
}

func (ob *OrderBook) IteratorBuySide() skiplist.Iterator {
	return iterator(ob.buyside)
}

func (ob *OrderBook) IteratorSellSide() skiplist.Iterator {
	return iterator(ob.sellside)
}

func iterator(book *skiplist.SkipList) skiplist.Iterator {
	return book.Iterator()
}

func (ob *OrderBook) TopPriceSellSide() Order {
	return lowPrice(ob.sellside)
}

func (ob *OrderBook) TopPriceBuySide() Order {
	return topPrice(ob.buyside)
}

func (ob *OrderBook) Empty() bool {
	iterBuy := ob.buyside.Iterator()
	okBuy := iterBuy.Next()
	iterSell := ob.sellside.Iterator()
	okSell := iterSell.Next()
	return !okBuy || !okSell
}

func topPrice(book *skiplist.SkipList) Order {
	iter := book.Iterator()
	iter.Next()
	return iter.Value().(Order)
}

func (ob *OrderBook) LowPriceSellSide() Order {
	return topPrice(ob.sellside)
}

func (ob *OrderBook) LowPriceBuySide() Order {
	return lowPrice(ob.buyside)
}

func lowPrice(book *skiplist.SkipList) Order {
	iter := book.SeekToLast()
	iter.Next()
	return iter.Value().(Order)
}

func (ob *OrderBook) GetTopTenPrices(side string) []Order {
	arr := make([]Order, 10)
	var it skiplist.Iterator
	if side == "buy" {
		it = ob.IteratorBuySide()
	} else { // sell
		it = ob.IteratorSellSide()
	}
	it.Next()
	for i := 0; i < 10; i++ {
		arr[i] = it.Value().(Order)
		it.Next()
	}
	// for testing purposes
	// var str string
	// if side == "buy" {
	// 	str = "HIGHEST 10 BIDS: "
	// } else {
	// 	str = "HIGHEST 10 ASKS: "
	// }
	// for i := 0; i < len(arr); i++ {
	// 	log.Info(str, "Index", i,
	// 		"Price", arr[i].Price,
	// 		"Quantity", arr[i].Qty,
	// 		"FillCost", arr[i].FillCost,
	// 		"ExchangeKey", arr[i].ExchangeKey,
	// 	)
	// }
	return arr
}
