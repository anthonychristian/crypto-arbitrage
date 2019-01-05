package trade

import (
	"testing"
	"time"

	"github.com/anthonychristian/crypto-arbitrage/orderbook"

	"github.com/stretchr/testify/assert"
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
}

func (s *TradeSuite) TearDownSuite() {}

func (s *TradeSuite) TestTrade() {
	CreateDummyExchangesNotProfitable()
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
	start, end, qty, err := detectArbitrage([]Pair{pair1, pair2, pair3})
	s.T().Log(start)
	s.T().Log(end)
	s.T().Log(qty)
	assert := assert.New(s.T())
	assert.True(err != nil)
	assert.Equal(990.5135725590688, qty)

	CreateDummyExchangesNotProfitableQty()
	start, end, qty, err = detectArbitrage([]Pair{pair1, pair2, pair3})
	s.T().Log(start)
	s.T().Log(end)
	s.T().Log(qty)
	assert.True(err != nil)
	assert.Equal(1021.5982950714285, qty)
	assert.Equal(((((start/150)*(1-0.001))-0.01)*2283000*(1-0.003)/14700*(1-0.003))-5, end)

	CreateDummyExchangesProfitable()
	start, end, qty, err = detectArbitrage([]Pair{pair1, pair2, pair3})
	s.T().Log(start)
	s.T().Log(end)
	s.T().Log(qty)
	assert.False(err != nil)
	assert.Equal(1021.5982950714285, qty)
	assert.Equal(((((start/150)*(1-0.001))-0.01)*2283000*(1-0.003)/14700*(1-0.003))-5, end)
}

func CreateDummyExchangesProfitable() {
	orderbook.InitExchanges()
	binance_eth_usdt := orderbook.Exchanges[orderbook.Binance].Books[orderbook.ETH_USDT]
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 149,
		Qty:   2,
	})
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 148,
		Qty:   1,
	})
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 147,
		Qty:   1,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 152,
		Qty:   3,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 151,
		Qty:   3,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 150,
		Qty:   3,
	})

	idx_eth_idr := orderbook.Exchanges[orderbook.Indodax].Books[orderbook.ETH_IDR]
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2283000,
		Qty:   3,
	})
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2282000,
		Qty:   1,
	})
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2281000,
		Qty:   1,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2286000,
		Qty:   2,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2285000,
		Qty:   1,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2284000,
		Qty:   1,
	})

	idx_usdt_idr := orderbook.Exchanges[orderbook.Indodax].Books[orderbook.USDT_IDR]
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 14680,
		Qty:   153,
	})
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 14670,
		Qty:   145,
	})
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 14660,
		Qty:   153,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 14711,
		Qty:   250,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 14710,
		Qty:   250,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 14700,
		Qty:   250,
	})
}
func CreateDummyExchangesNotProfitableQty() {
	orderbook.InitExchanges()
	binance_eth_usdt := orderbook.Exchanges[orderbook.Binance].Books[orderbook.ETH_USDT]
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 149,
		Qty:   2,
	})
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 148,
		Qty:   1,
	})
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 147,
		Qty:   1,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 152,
		Qty:   2,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 151,
		Qty:   1,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 150,
		Qty:   1,
	})

	idx_eth_idr := orderbook.Exchanges[orderbook.Indodax].Books[orderbook.ETH_IDR]
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2283000,
		Qty:   3,
	})
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2282000,
		Qty:   1,
	})
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2281000,
		Qty:   1,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2286000,
		Qty:   2,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2285000,
		Qty:   1,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2284000,
		Qty:   1,
	})

	idx_usdt_idr := orderbook.Exchanges[orderbook.Indodax].Books[orderbook.USDT_IDR]
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 14680,
		Qty:   153,
	})
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 14670,
		Qty:   145,
	})
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 14660,
		Qty:   153,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 14711,
		Qty:   145,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 14710,
		Qty:   145,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 14700,
		Qty:   145,
	})
}
func CreateDummyExchangesNotProfitable() {
	orderbook.InitExchanges()
	binance_eth_usdt := orderbook.Exchanges[orderbook.Binance].Books[orderbook.ETH_USDT]
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 153,
		Qty:   2,
	})
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 152,
		Qty:   1,
	})
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 151,
		Qty:   1,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 156,
		Qty:   2,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 155,
		Qty:   1,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 154,
		Qty:   1,
	})

	idx_eth_idr := orderbook.Exchanges[orderbook.Indodax].Books[orderbook.ETH_IDR]
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2273000,
		Qty:   3,
	})
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2272000,
		Qty:   1,
	})
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2271000,
		Qty:   1,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2276000,
		Qty:   2,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2275000,
		Qty:   1,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2274000,
		Qty:   1,
	})

	idx_usdt_idr := orderbook.Exchanges[orderbook.Indodax].Books[orderbook.USDT_IDR]
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 14680,
		Qty:   153,
	})
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 14670,
		Qty:   145,
	})
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 14660,
		Qty:   153,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 14711,
		Qty:   145,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 14710,
		Qty:   145,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 14700,
		Qty:   145,
	})
}
