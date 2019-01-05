package trade

import (
	"testing"
	"time"

	"github.com/anthonychristian/crypto-arbitrage/orderbook"

	"github.com/stretchr/testify/suite"
)

var (
	NY, _ = time.LoadLocation("America/New_York")
)

type TradeSuite struct{ suite.Suite }

func TestTradeSuite(t *testing.T) {
	suite.Run(t, new(TradeSuite))
}

func (s *TradeSuite) SetupSuite() {
	orderbook.InitExchanges()
	CreateDummyExchanges()
}

func (s *TradeSuite) TearDownSuite() {}

func (s *TradeSuite) TestTrade() {
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
		side:     Sell,
	}
	qty, err := trade([]Pair{pair1, pair2, pair3})
	s.T().log(qty)
	assert := assert.new(s.T())
	assert.Equal(true, qty > 0)
}

func CreateDummyExchanges() {
	binance_eth_usdt := Exchanges[Binance].Books[Symbol("ETH/USDT")]
	binance_eth_usdt.AddBuy(Order{
		Price: 200,
		Qty:   2,
	})
	binance_eth_usdt.AddBuy(Order{
		Price: 150,
		Qty:   1,
	})
	binance_eth_usdt.AddBuy(Order{
		Price: 100,
		Qty:   1,
	})
	binance_eth_usdt.AddSell(Order{
		Price: 400,
		Qty:   2,
	})
	binance_eth_usdt.AddSell(Order{
		Price: 350,
		Qty:   1,
	})
	binance_eth_usdt.AddSell(Order{
		Price: 300,
		Qty:   1,
	})

	idx_eth_idr := Exchanges[Indodax].Books[Symbol("ETH/IDR")]
	idx_eth_idr.AddBuy(Order{
		Price: 2200000,
		Qty:   3,
	})
	idx_eth_idr.AddBuy(Order{
		Price: 2100000,
		Qty:   1,
	})
	idx_eth_idr.AddBuy(Order{
		Price: 2000000,
		Qty:   1,
	})
	idx_eth_idr.AddSell(Order{
		Price: 2500000,
		Qty:   2,
	})
	idx_eth_idr.AddSell(Order{
		Price: 2400000,
		Qty:   1,
	})
	idx_eth_idr.AddSell(Order{
		Price: 2300000,
		Qty:   1,
	})

	idx_usdt_idr := Exchanges[Indodax].Books[Symbol("USDT/IDR")]
	idx_usdt_idr.AddBuy(Order{
		Price: 14000,
		Qty:   3,
	})
	idx_usdt_idr.AddBuy(Order{
		Price: 13500,
		Qty:   1,
	})
	idx_usdt_idr.AddBuy(Order{
		Price: 13000,
		Qty:   1,
	})
	idx_usdt_idr.AddSell(Order{
		Price: 16000,
		Qty:   2,
	})
	idx_usdt_idr.AddSell(Order{
		Price: 15500,
		Qty:   1,
	})
	idx_usdt_idr.AddSell(Order{
		Price: 15000,
		Qty:   1,
	})
}
