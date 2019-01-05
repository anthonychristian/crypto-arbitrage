package orderbook

import (
	"errors"
	"fmt"
	"math"
)

type ExchangeKey string

const (
	Binance ExchangeKey = "Binance"
	Indodax ExchangeKey = "Indodax"
)

type Exchange struct {
	Books OrderBookMap // The key is the trading pair, e.g. "BTC/USDC"
}

type ExchangeMap map[ExchangeKey]Exchange
type Symbols map[Symbol][]ExchangeKey

var (
	Exchanges = make(ExchangeMap)
	ExFeeMap  = map[ExchangeKey]float64{ // Key: Exchange, Val: Fee
		Binance: 0.001,
		Indodax: 0.003, // only for market order
	}
	ExSymbolMap = map[ExchangeKey][]Symbol{
		Binance: []Symbol{ETH_USDT},
		Indodax: []Symbol{ETH_IDR, USDT_IDR},
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

func WithdrawFee(exchange ExchangeKey, currency Currency, qty float64) (fee float64, err error) {
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
	for ex, symbols := range ExSymbolMap {
		Exchanges[ex] = Exchange{Books: make(OrderBookMap)}
		for _, symbol := range symbols {
			Exchanges[ex].Books[symbol] = NewOrderBook()
		}
	}
}
