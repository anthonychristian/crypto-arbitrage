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

type CalcParts struct {
	operator  string
	amt       float64
	nextCurr  orderbook.Currency
	prevCurr  orderbook.Currency
	action    string
	orderbook *orderbook.OrderBook
}

func execCalc(a float64, b CalcParts) (float64, error) {
	switch b.operator {
	case "+":
		return a + b.amt, nil
	case "-":
		return a - b.amt, nil
	case "*":
		return a * b.amt, nil
	case "/":
		return a / b.amt, nil
	default:
		return 0, errors.New("operator invalid")
	}
}

func execReverseCalc(a float64, b CalcParts) (float64, error) {
	switch b.operator {
	case "+":
		return a - b.amt, nil
	case "-":
		return a + b.amt, nil
	case "*":
		return a / b.amt, nil
	case "/":
		return a * b.amt, nil
	default:
		return 0, errors.New("operator invalid")
	}
}

func detectArbitrage(pairs []Pair, exchanges orderbook.ExchangeMap) (start float64, end float64, qty float64, notProfitable bool, err error) {
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
	calcs := []CalcParts{}

	for index, pair := range pairs {
		ob := exchanges[pair.exchange].Books[pair.symbol]
		if ob.Empty() {
			return 0, 0, 0, false, errors.New("ob not initialized for" + string(pair.symbol))
		}
		leftCurr := orderbook.GetLeftCurrency(pair.symbol)
		rightCurr := orderbook.GetRightCurrency(pair.symbol)
		var bestPrice orderbook.Order
		if pair.side == Buy {
			if owned.curr != rightCurr {
				return 0, 0, 0, false, errors.New("wrong starting currency " + string(owned.curr))
			}
			bestPrice = ob.LowPriceSellSide()
			// if buy -> now own the left curr
			owned.curr = leftCurr
			owned.qty = owned.qty / bestPrice.Price
			// max qty you can buy = max(available quantity at the price, funds you have)
			owned.funds = math.Min(bestPrice.Qty, owned.funds/bestPrice.Price)
			calcs = append(calcs, CalcParts{
				operator:  "/",
				amt:       bestPrice.Price,
				prevCurr:  rightCurr,
				nextCurr:  leftCurr,
				action:    "buy",
				orderbook: ob,
			})
		} else {
			if owned.curr != leftCurr {
				return 0, 0, 0, false, errors.New("wrong starting currency " + string(owned.curr))
			}
			bestPrice = ob.TopPriceBuySide()
			// if sell -> now own the right curr
			owned.curr = rightCurr
			owned.qty = owned.qty * bestPrice.Price
			// max qty you can sell = max(available quantity at the price, funds you have)
			owned.funds = math.Min(bestPrice.Qty, owned.funds) * bestPrice.Price
			calcs = append(calcs, CalcParts{
				operator:  "*",
				amt:       bestPrice.Price,
				prevCurr:  leftCurr,
				nextCurr:  rightCurr,
				action:    "sell",
				orderbook: ob,
			})
		}

		// add taker price
		exFee := (1 - orderbook.ExFeeMap[pair.exchange])
		owned.qty = owned.qty * exFee
		owned.funds = owned.funds * exFee
		calcs = append(calcs, CalcParts{
			operator:  "*",
			amt:       exFee,
			prevCurr:  owned.curr,
			nextCurr:  owned.curr,
			action:    "exchangeFee",
			orderbook: ob,
		})

		// add withdraw fee if transfering to next exchange
		nextPair := pairs[0]
		if index < len(pairs)-1 {
			nextPair = pairs[index+1]
		}
		if nextPair.exchange != pair.exchange {
			withdrawFee, err := orderbook.WithdrawFee(pair.exchange, owned.curr, owned.qty)
			if err != nil {
				return 0, 0, 0, false, err
			}
			owned.qty -= withdrawFee
			// for funds calculation
			withdrawFee, err = orderbook.WithdrawFee(pair.exchange, owned.curr, owned.funds)
			if err != nil {
				return 0, 0, 0, false, err
			}
			owned.funds -= withdrawFee
			calcs = append(calcs, CalcParts{
				operator:  "-",
				amt:       withdrawFee,
				prevCurr:  owned.curr,
				nextCurr:  owned.curr,
				action:    "withdrawFee",
				orderbook: ob,
			})
		}
	}
	if owned.curr != startCurr {
		return 0, 0, 0, false, errors.New("Cycle incomplete! Started with " + string(startCurr) + " end with " + string(owned.curr))
	}

	endingFund := owned.funds

	// if profitable
	if owned.qty > 1000 && owned.funds > 0 {
		// trace back needed funds from ending fund
		for i := len(calcs) - 1; i >= 0; i-- {
			b := calcs[i]
			endFund := owned.funds
			owned.funds, err = execReverseCalc(owned.funds, b)
			if err != nil {
				return 0, 0, 0, false, err
			}
			owned.curr = b.prevCurr

			// decrease qty in orderbook
			if b.action == "buy" {
				bestPrice := b.orderbook.LowPriceSellSide()
				// buying qty = amount received = endFund
				bestPrice.Qty -= endFund
				b.orderbook.AddSell(bestPrice)
				bestPrice = b.orderbook.LowPriceSellSide()
			} else if b.action == "sell" {
				bestPrice := b.orderbook.TopPriceBuySide()
				// selling qty = amount sold (amount had originally) = owned.funds
				bestPrice.Qty -= owned.funds
				b.orderbook.AddBuy(bestPrice)
				bestPrice = b.orderbook.TopPriceBuySide()
			}
		}
		if owned.curr != startCurr {
			return owned.funds, endingFund, owned.qty, false, errors.New("Traceback cycle wrong! Started with " + string(startCurr) + " end with " + string(owned.curr))
		}

		if endingFund < owned.funds {
			return owned.funds, endingFund, owned.qty, true, errors.New("Not profitable for available qty, starting fund = " + strconv.FormatFloat(owned.funds, 'f', -1, 64) + " ending fund = " + strconv.FormatFloat(endingFund, 'f', -1, 64))
		}

		return owned.funds, endingFund, owned.qty, false, nil
	}
	return owned.funds, endingFund, owned.qty, true, errors.New("Not profitable, final qty = " + strconv.FormatFloat(owned.qty, 'f', -1, 64))
}
