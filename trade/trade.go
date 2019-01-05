package trade

import (
	"errors"
	"math"
	"strconv"

	"github.com/anthonychristian/crypto-arbitrage/orderbook"
)

type Pair struct {
	symbol   orderbook.Symbol
	exchange orderbook.ExchangeKey
	side     OrderSide
}

type OwnedCurr struct {
	curr orderbook.Currency
	qty  float64
}

func trade(pairs []Pair) (qty float64, err error) {
	minQty := math.Inf(1)
	// starting currency set owned = 1
	// if sell = left curr
	// if buy = right curr
	owned := OwnedCurr{
		curr: orderbook.GetLeftCurrency(pairs[0].symbol),
		qty:  1,
	}
	startCurr := orderbook.GetLeftCurrency(pairs[0].symbol)
	if pairs[0].side == Buy {
		owned.curr = orderbook.GetRightCurrency(pairs[0].symbol)
		startCurr = orderbook.GetRightCurrency(pairs[0].symbol)
	}

	for index, pair := range pairs {
		ob := orderbook.Exchanges[pair.exchange].Books[pair.symbol]
		leftCurr := orderbook.GetLeftCurrency(pair.symbol)
		rightCurr := orderbook.GetRightCurrency(pair.symbol)
		var bestPrice orderbook.Order
		if pair.side == Buy {
			if owned.curr != rightCurr {
				return 0, errors.New("wrong starting currency " + string(owned.curr))
			}
			bestPrice = ob.LowPriceSellSide()
			// if buy -> now own the left curr
			owned.curr = leftCurr
			owned.qty = owned.qty / bestPrice.Price
		} else {
			if owned.curr != leftCurr {
				return 0, errors.New("wrong starting currency " + string(owned.curr))
			}
			bestPrice = ob.TopPriceBuySide()
			// if sell -> now own the right curr
			owned.curr = rightCurr
			owned.qty = owned.qty * bestPrice.Price
		}
		minQty = math.Min(minQty, bestPrice.Qty)

		// add taker price
		owned.qty = owned.qty * (1 - orderbook.ExFeeMap[pair.exchange])

		// add withdraw fee if transfering to next exchange
		nextPair := pairs[0]
		if index < len(pairs) {
			nextPair = pairs[index+1]
		}
		if nextPair.exchange != pair.exchange {
			withdrawFee, err := orderbook.WithdrawFee(pair.exchange, owned.curr, owned.qty)
			if err != nil {
				return 0, err
			}
			owned.qty -= withdrawFee
		}
	}
	if owned.curr != startCurr {
		return 0, errors.New("Cycle incomplete! Started with " + string(startCurr) + " end with " + string(owned.curr))
	}
	if owned.qty > 1 {
		return minQty, nil
	}
	return -1, errors.New("Not profitable, final qty = " + strconv.FormatFloat(owned.qty, 'f', -1, 64))
}
