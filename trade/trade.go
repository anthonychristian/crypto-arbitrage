package trade

import (
	"math"
)

type Pair struct {
	symbol   Symbol
	exchange Exchange
	side     OrderSide
}

type OwnedCurr struct {
	curr  Currency
	owned float64
}

func trade(pairs []Pair) (qty float64, err error) {
	min_qty := math.Inf(1)
	owned := OwnedCurr{}
}
