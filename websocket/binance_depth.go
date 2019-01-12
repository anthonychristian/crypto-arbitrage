package websocket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/anthonychristian/crypto-arbitrage/trade"

	binance "github.com/adshao/go-binance"
	"github.com/anthonychristian/crypto-arbitrage/orderbook"
)

const (
	// Bitcoin Symbol
	binanceSymbol = "ETHUSDT"
)

// BinanceDepthResponse is the type retrieved from the first orderbook snapshot
type BinanceDepthResponse struct {
	LastUpdateID int64           `json:"lastUpdateId"`
	Bids         [][]interface{} `json:"bids"`
	Asks         [][]interface{} `json:"asks"`
}

// BinanceDepthEvent is the type retrieved from websocket events
type BinanceDepthEvent struct {
	Event         string        `json:"e"`
	Time          int64         `json:"E"`
	Symbol        string        `json:"s"`
	FirstUpdateID int64         `json:"U"`
	FinalUpdateID int64         `json:"u"`
	Bids          []binance.Bid `json:"b"`
	Asks          []binance.Ask `json:"a"`
}

// BinanceOrderBook is used for temporary orderbook struct
type BinanceOrderBook struct {
	Bids map[int64]float64
	Asks map[int64]float64
}

// OrderEvent used for the channel. Need to keep FirstUpdateID and FinalUpdateID of each event
type OrderEvent struct {
	Bids          []binance.Bid
	Asks          []binance.Ask
	Side          string
	FirstUpdateID int64
	FinalUpdateID int64
}

// Channel to hold the events you receive from WebSocket Stream
var queueBinChan = make(chan *BinanceDepthEvent)
var quitBinQueueChan = make(chan int)

// Keep track of lastUpdateID of the first snapshot, needed to correctly add websocket events
var lastUpdateID int64 = -1

// Keep track of the last event's final update, needed to correctly add websocket events
var prevu int64 = -1

// AddBinOrderBookToSkipList is used to parse binance Bids and Asks to add into the BinanceOrderBook for the Skiplist Orderbook
func AddBinOrderBookToSkipList(sl *orderbook.OrderBook, bids []binance.Bid, asks []binance.Ask) {
	for _, elem := range bids {
		fQty, _ := strconv.ParseFloat(elem.Quantity, 64)
		fPrice, _ := strconv.ParseFloat(elem.Price, 64)
		newOrder := orderbook.Order{
			Price: fPrice,
			Qty:   fQty,
		}
		sl.AddBuy(newOrder)
	}
	for _, elem := range asks {
		fQty, _ := strconv.ParseFloat(elem.Quantity, 64)
		fPrice, _ := strconv.ParseFloat(elem.Price, 64)
		newOrder := orderbook.Order{
			Price: fPrice,
			Qty:   fQty,
		}
		sl.AddSell(newOrder)
	}
	for _, worker := range trade.TradeWorkers[orderbook.Binance] {
		worker.ObUpdated <- orderbook.Binance
	}
}

// AddBinanceBidEventToSkipList is used to add the Bid Event to the Skiplist, with restrictions
func AddBinanceBidEventToSkipList(sl *orderbook.OrderBook, v *binance.Bid) {
	fQty, _ := strconv.ParseFloat(v.Quantity, 64)
	fPrice, _ := strconv.ParseFloat(v.Price, 64)
	orderToAdd := orderbook.Order{
		Price: fPrice,
		Qty:   fQty,
	}
	sl.AddBuy(orderToAdd)
	for _, worker := range trade.TradeWorkers[orderbook.Binance] {
		worker.ObUpdated <- orderbook.Binance
	}
}

// AddBinanceAskEventToSkipList is used to add the Ask Event to the Skiplist, with restrictions
func AddBinanceAskEventToSkipList(sl *orderbook.OrderBook, v *binance.Ask) {
	fQty, _ := strconv.ParseFloat(v.Quantity, 64)
	fPrice, _ := strconv.ParseFloat(v.Price, 64)
	orderToAdd := orderbook.Order{
		Price: fPrice,
		Qty:   fQty,
	}
	sl.AddSell(orderToAdd)
	for _, worker := range trade.TradeWorkers[orderbook.Binance] {
		worker.ObUpdated <- orderbook.Binance
	}
}

// Functions to manage local order book

// InitBinanceHandler is used to initialize orderbook and websocket handler
func InitBinanceHandler() {
	if lastUpdateID != -1 {
		lastUpdateID = -1
		prevu = -1
		quitBinQueueChan <- 0
	}
	go GetDepthFromBinance()
	time.Sleep(1000 * time.Millisecond)
	manageBinanceOrderBook()
	go manageBinanceQueue()
}

var wsDepthHandler = func(event *binance.WsDepthEvent) {
	// Put event in BinanceDepth struct
	data := BinanceDepthEvent{
		Event:         event.Event,
		Time:          event.Time,
		Symbol:        event.Symbol,
		FirstUpdateID: event.FirstUpdateID,
		FinalUpdateID: event.UpdateID,
		Bids:          event.Bids,
		Asks:          event.Asks,
	}

	// Put the data received inside the queue
	queueBinChan <- &data
}

var depthErrHandler = func(err error) {
	fmt.Println("error", "err", err.Error())
}

// GetDepthFromBinance is the function used to start websocket connection to binance
func GetDepthFromBinance() {
	doneC, _, err := binance.WsDepthServe(binanceSymbol, wsDepthHandler, depthErrHandler)
	if err != nil {
		return
	}
	<-doneC
}

// Function to get the depth snapshot from API, and insert into the local order book
func manageBinanceOrderBook() {
	// Get the data from the order book
	depth := getBinanceDepth()
	// add bids and asks into the skiplist orderbook
	AddBinOrderBookToSkipList(orderbook.Exchanges[orderbook.Binance].Books[orderbook.ETH_USDT], depth.Bids, depth.Asks)
	// update the lastUpdateID of the snapshot
	lastUpdateID = depth.LastUpdateID
	fmt.Println("Binance Orderbook Initialized, empty: ", orderbook.Exchanges[orderbook.Binance].Books[orderbook.ETH_USDT].Empty())
}

// Function to manage queue channel.
// Each event is sent into the channel,
// once lastUpdateID is updated, start processing events,
// updating the orderbook when appropriate
func manageBinanceQueue() {
	for {
		// Start processing queueChan when lastUpdateID is initialized
		if lastUpdateID != -1 {
			v, ok := <-queueBinChan
			if ok {
				// ignore events where u <= lastUpdateID
				if v.FinalUpdateID <= lastUpdateID {
					continue
				}
				// finding the first event to use
				if prevu == -1 && v.FirstUpdateID <= lastUpdateID+1 && v.FinalUpdateID >= lastUpdateID+1 {
					// Add the bids and asks to the SkipList OrderBook
					for _, elem := range v.Bids {
						AddBinanceBidEventToSkipList(orderbook.Exchanges[orderbook.Binance].Books[orderbook.ETH_USDT], &elem)
					}
					for _, elem := range v.Asks {
						AddBinanceAskEventToSkipList(orderbook.Exchanges[orderbook.Binance].Books[orderbook.ETH_USDT], &elem)
					}
					prevu = v.FinalUpdateID
					// for testing purposes
				} else if prevu != -1 && v.FirstUpdateID == prevu+1 {
					// Add the bids and asks to the SkipList OrderBook
					for _, elem := range v.Bids {
						AddBinanceBidEventToSkipList(orderbook.Exchanges[orderbook.Binance].Books[orderbook.ETH_USDT], &elem)
					}
					for _, elem := range v.Asks {
						AddBinanceAskEventToSkipList(orderbook.Exchanges[orderbook.Binance].Books[orderbook.ETH_USDT], &elem)
					}
					prevu = v.FinalUpdateID
					// for testing purposes
				}
			}
		} else {
			_, ok := <-quitBinQueueChan
			if ok {
				return
			}
		}
	}
}

func getBinanceDepth() binance.DepthResponse {
	response, err := http.Get("https://www.binance.com/api/v1/depth?symbol=ETHUSDT&limit=1000")
	if err != nil {
		return binance.DepthResponse{}
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return binance.DepthResponse{}
	}
	// unmarshal JSON response
	depthResponse := BinanceDepthResponse{}
	jsonErr := json.Unmarshal(contents, &depthResponse)
	if jsonErr != nil {
		return binance.DepthResponse{}
	}

	depthToReturn := binance.DepthResponse{
		LastUpdateID: depthResponse.LastUpdateID,
	}

	for _, e := range depthResponse.Bids {
		depthToReturn.Bids = append(depthToReturn.Bids, binance.Bid{
			Price:    e[0].(string),
			Quantity: e[1].(string),
		})
	}
	for _, e := range depthResponse.Asks {
		depthToReturn.Asks = append(depthToReturn.Asks, binance.Ask{
			Price:    e[0].(string),
			Quantity: e[1].(string),
		})
	}
	return depthToReturn
}
