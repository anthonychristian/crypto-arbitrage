package orderbook

import (
	"fmt"
	"strings"
)

type Symbol string

const (
	BTC_USDC Symbol = "BTC/USDC"
	BTC_ETH  Symbol = "BTC/ETH"
)

type Symbols map[Symbol][]ExchangeKey // The key is the symbol pair, the []string is a list of exchanges for the pair

var (
	SymbolMap = Symbols{
		BTC_USDC: []ExchangeKey{Binance},
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

// GetLeftCurrency returns the left currency in the symbol
func GetLeftCurrency(symbol string) string {
	currencies := strings.Split(symbol, "/")
	return currencies[0]
}

// GetRightCurrency returns the right currency in the symbol
func GetRightCurrency(symbol string) string {
	currencies := strings.Split(symbol, "/")
	return currencies[1]
}
