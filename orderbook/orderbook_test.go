package orderbook

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/suite"
)

var (
	NY, _ = time.LoadLocation("America/New_York")
)

type OrderBookSuite struct{ suite.Suite }

func TestOrderBookSuite(t *testing.T) {
	suite.Run(t, new(OrderBookSuite))
}

func (s *OrderBookSuite) SetupSuite() {}

func (s *OrderBookSuite) TearDownSuite() {}

func (s *OrderBookSuite) TestOrderBook() {
	ob := setupInitialBook()

	fmt.Println("Sellside book...")
	iter := ob.IteratorSellSide()
	for iter.Next() {
		fmt.Printf("Key: %v\n", iter.Key())
		fmt.Printf("Order: %v\n", iter.Value())
	}

	tpBuy := ob.TopPriceBuySide()
	lpBuy := ob.LowPriceBuySide()
	tpSell := ob.TopPriceSellSide()
	lpSell := ob.LowPriceSellSide()
	fmt.Printf("Top price buy/sell: %v/%v\n", tpBuy, tpSell)
	fmt.Printf("Low price buy/sell: %v/%v\n", lpBuy, lpSell)
	assert.Equal(s.T(), 108.0, tpBuy.Price)
	assert.Equal(s.T(), 110.0, tpSell.Price)

	// Test price level removal
	ob.AddBuy(Order{
		Price: 108,
		Qty:   0,
	})
	tpBuy = ob.TopPriceBuySide()
	assert.Equal(s.T(), 107.0, tpBuy.Price)
}

func setupInitialBook() *OrderBook {
	ob := NewOrderBook()
	ob.AddBuy(Order{
		Price: 108,
		Qty:   30,
	})
	ob.AddBuy(Order{
		Price: 107,
		Qty:   100,
	})
	ob.AddBuy(Order{
		Price: 106,
		Qty:   50,
	})
	ob.AddSell(Order{
		Price: 110,
		Qty:   10,
	})
	ob.AddSell(Order{
		Price: 109,
		Qty:   20,
	})
	return ob
}
