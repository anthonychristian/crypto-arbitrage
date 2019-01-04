package trade

import (
	"errors"
	"fmt"
	"math"

	"github.com/alpacahq/crypto-poc/orderbook"
)

type ExchangeKey string

const (
	Binance ExchangeKey = "Binance"
	Indodax ExchangeKey = "Indodax"
)

type Exchange struct {
	Books orderbook.OrderBookMap // The key is the trading pair, e.g. "BTC/USDC"
}

type ExchangeMap map[ExchangeKey]Exchange
type Symbols map[orderbook.Symbol][]ExchangeKey

var (
	Exchanges = make(ExchangeMap)
	ExFeeMap  = map[ExchangeKey]float64{ // Key: Exchange, Val: Fee
		Binance: 1.001,
		Indodax: 1.003, // only for market order
	}
	SymbolMap = Symbols{
		ETH_USDT: []ExchangeKey{Binance},
		ETH_IDR:  []ExchangeKey{Indodax},
		USDT_IDR: []ExchangeKey{Indodax},
	}
)

func (s Symbols) ListSymbols() (list []string) {
	for key := range s {
		list = append(list, string(key))
	}
	return list
}

func (s Symbols) IsTradeAble(symbol string) bool {
	if _, ok := SymbolMap[Symbol(symbol)]; !ok {
		return false
	}
	return true
}

func (s Symbols) GetExchangesForSymbol(symbol string) (exchanges []Exchange, err error) {
	if list, ok := s[Symbol(symbol)]; !ok {
		return nil, fmt.Errorf("symbol not in list of tradable symbols")
	} else {
		for _, name := range list {
			exchanges = append(exchanges, Exchanges[name])
		}
	}
	return exchanges, nil
}

func WithdrawFee(exchange ExchangeKey, currency orderbook.Currency, qty float64) (fee float64, err error) {
	if exchange == Indodax {
		switch currency {
		case ETH:
			return 0.005, nil
		case IDR:
			return math.Max(0.01*qty, 25000), nil
		case USDT:
			return 5, nil
		default:
			return 0, errors.New("currency " + string(currency) + " unrecognized in Indodax")
		}
	} else if exchange == Binance {
		switch currency {
		case ETH:
			return 0.01, nil
		case USDT:
			return 2, nil
		default:
			return 0, errors.New("currency " + string(currency) + " unrecognized in Binance")
		}
	} else {
		return 0, errors.New(string(exchange) + " is unrecognized")
	}
}

func InitExchanges() {
	for symbol, exchanges := range SymbolMap {
		for _, exchange := range exchanges {
			if Exchanges[ex] == nil {
				Exchanges[ex] = Exchange{Books: make(orderbook.OrderBookMap)}
			}
			Exchanges[ex].Books[symbol] = orderbook.NewOrderBook()
		}
	}
}

func CreateDummyExchanges() {
	binance_eth_usdt := Exchanges[Binance].Books[orderbook.Symbol("ETH/USDT")]
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 200,
		Qty:   2,
	})
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 150,
		Qty:   1,
	})
	binance_eth_usdt.AddBuy(orderbook.Order{
		Price: 100,
		Qty:   1,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 400,
		Qty:   2,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 350,
		Qty:   1,
	})
	binance_eth_usdt.AddSell(orderbook.Order{
		Price: 300,
		Qty:   1,
	})

	idx_eth_idr := Exchanges[Indodax].Books[orderbook.Symbol("ETH/IDR")]
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2200000,
		Qty:   3,
	})
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2100000,
		Qty:   1,
	})
	idx_eth_idr.AddBuy(orderbook.Order{
		Price: 2000000,
		Qty:   1,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2500000,
		Qty:   2,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2400000,
		Qty:   1,
	})
	idx_eth_idr.AddSell(orderbook.Order{
		Price: 2300000,
		Qty:   1,
	})

	idx_usdt_idr := Exchanges[Indodax].Books[orderbook.Symbol("USDT/IDR")]
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 14000,
		Qty:   3,
	})
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 13500,
		Qty:   1,
	})
	idx_usdt_idr.AddBuy(orderbook.Order{
		Price: 13000,
		Qty:   1,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 16000,
		Qty:   2,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 15500,
		Qty:   1,
	})
	idx_usdt_idr.AddSell(orderbook.Order{
		Price: 15000,
		Qty:   1,
	})
}
