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
	curr  orderbook.Currency
	qty   float64 // to calculate percent up
	funds float64 // to calculate how much to buy
}

func detectArbitrage(pairs []Pair) (start float64, end float64, qty float64, err error) {
	// starting currency set owned = 1
	// if sell = left curr
	// if buy = right curr
	owned := OwnedCurr{
		curr:  orderbook.GetLeftCurrency(pairs[0].symbol),
		qty:   1000,
		funds: 10000,
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
				return 0, 0, 0, errors.New("wrong starting currency " + string(owned.curr))
			}
			bestPrice = ob.LowPriceSellSide()
			// if buy -> now own the left curr
			owned.curr = leftCurr
			owned.qty = owned.qty / bestPrice.Price
			// max qty you can buy = max(available quantity at the price, funds you have)
			owned.funds = math.Min(bestPrice.Qty, owned.funds/bestPrice.Price)
		} else {
			if owned.curr != leftCurr {
				return 0, 0, 0, errors.New("wrong starting currency " + string(owned.curr))
			}
			bestPrice = ob.TopPriceBuySide()
			// if sell -> now own the right curr
			owned.curr = rightCurr
			owned.qty = owned.qty * bestPrice.Price
			// max qty you can sell = max(available quantity at the price, funds you have)
			owned.funds = math.Min(bestPrice.Qty, owned.funds) * bestPrice.Price
		}

		// add taker price
		owned.qty = owned.qty * (1 - orderbook.ExFeeMap[pair.exchange])
		owned.funds = owned.funds * (1 - orderbook.ExFeeMap[pair.exchange])

		// add withdraw fee if transfering to next exchange
		nextPair := pairs[0]
		if index < len(pairs)-1 {
			nextPair = pairs[index+1]
		}
		if nextPair.exchange != pair.exchange {
			withdrawFee, err := orderbook.WithdrawFee(pair.exchange, owned.curr, owned.qty)
			if err != nil {
				return 0, 0, 0, err
			}
			owned.qty -= withdrawFee
			// for funds calculation
			withdrawFee, err = orderbook.WithdrawFee(pair.exchange, owned.curr, owned.funds)
			if err != nil {
				return 0, 0, 0, err
			}
			owned.funds -= withdrawFee
		}
	}
	if owned.curr != startCurr {
		return 0, owned.qty, 0, errors.New("Cycle incomplete! Started with " + string(startCurr) + " end with " + string(owned.curr))
	}

	endingFund := owned.funds

	// if profitable
	if owned.qty > 1000 && owned.funds > 0 {
		// trace back needed funds from ending fund
		for i := len(pairs) - 1; i >= 0; i-- {
			pair := pairs[i]
			ob := orderbook.Exchanges[pair.exchange].Books[pair.symbol]
			leftCurr := orderbook.GetLeftCurrency(pair.symbol)
			rightCurr := orderbook.GetRightCurrency(pair.symbol)
			var bestPrice orderbook.Order
			nextPair := pairs[0]
			if i < len(pairs)-1 {
				nextPair = pairs[i+1]
			}
			if nextPair.exchange != pair.exchange {
				// for funds calculation
				withdrawFee, err := orderbook.WithdrawFee(pair.exchange, owned.curr, owned.funds)
				if err != nil {
					return 0, 0, 0, err
				}
				// get original fund before withdraw fee
				owned.funds += withdrawFee
			}
			// get original fund before taker fee
			owned.funds = owned.funds / (1 - orderbook.ExFeeMap[pair.exchange])

			if pair.side == Buy {
				if owned.curr != leftCurr {
					return 0, 0, 0, errors.New("traceback: wrong starting currency " + string(owned.curr))
				}
				bestPrice = ob.LowPriceSellSide()
				// if buy -> originally had right Curr
				owned.curr = rightCurr
				// original fund of right Curr = left Curr * price
				owned.funds = owned.funds * bestPrice.Price
			} else {
				if owned.curr != rightCurr {
					return 0, 0, 0, errors.New("traceback: wrong starting currency " + string(owned.curr))
				}
				bestPrice = ob.TopPriceBuySide()
				// if sell -> orifinally had left curr
				owned.curr = leftCurr
				// original funds of left Curr = right Curr / price
				owned.funds = owned.funds / bestPrice.Price
			}

		}
		if owned.curr != startCurr {
			return owned.funds, endingFund, owned.qty, errors.New("Traceback cycle wrong! Started with " + string(startCurr) + " end with " + string(owned.curr))
		}

		if endingFund < owned.funds {
			return owned.funds, endingFund, owned.qty, errors.New("Not profitable for available qty, starting fund = " + strconv.FormatFloat(owned.funds, 'f', -1, 64) + " ending fund = " + strconv.FormatFloat(endingFund, 'f', -1, 64))
		}

		return owned.funds, endingFund, owned.qty, nil
	}
	return owned.funds, endingFund, owned.qty, errors.New("Not profitable, final qty = " + strconv.FormatFloat(owned.qty, 'f', -1, 64))
}
