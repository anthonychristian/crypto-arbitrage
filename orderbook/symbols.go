package orderbook

import (
	"strings"
)

type Currency string

const (
	ETH  Currency = "ETH"
	USDT Currency = "USDT"
	IDR  Currency = "IDR"
	USD  Currency = "USD"
)

type Symbol string

const (
	ETH_USDT Symbol = "ETH/USDT"
	ETH_IDR  Symbol = "ETH/IDR"
	USDT_IDR Symbol = "USDT/IDR"
)

// GetLeftCurrency returns the left currency in the symbol
func GetLeftCurrency(symbol Symbol) Currency {
	currencies := strings.Split(string(symbol), "/")
	return Currency(currencies[0])
}

// GetRightCurrency returns the right currency in the symbol
func GetRightCurrency(symbol Symbol) Currency {
	currencies := strings.Split(string(symbol), "/")
	return Currency(currencies[1])
}
