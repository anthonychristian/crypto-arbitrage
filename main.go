package main

import (
	"time"

	"github.com/alpacahq/gopaca/log"
	"github.com/anthonychristian/crypto-arbitrage/indodax"
	"github.com/anthonychristian/crypto-arbitrage/orderbook"
	"github.com/anthonychristian/crypto-arbitrage/websocket"
	"github.com/joho/godotenv"
	"github.com/kataras/iris"
	irisWs "github.com/kataras/iris/websocket"
)

func init() {
	orderbook.Exchanges[orderbook.Binance] = orderbook.Exchange{Books: make(orderbook.OrderBookMap)}
	initExchanges()

	// initialize API gateway
	_ = indodax.InitIndodax()
}

func initExchanges() {
	for k := range orderbook.SymbolMap {
		for _, ex := range orderbook.SymbolMap[k] {
			orderbook.Exchanges[ex].Books[orderbook.Symbol(k)] = orderbook.NewOrderBook()
		}
	}
	indodax.InitOrderBook()
	initOrderbookWebsocket()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := iris.New()
	app.Get("/", func(ctx iris.Context) {
		ctx.ServeFile("view/websockets.html", false)
	})

	go updateDepthToWorker()

	// Using iris websocket to show orderbook updates (for testing purposes)
	// Open Localhost 8080 to start orderbook websocket
	setupWebsocket(app)
	app.Run(iris.Addr(":8080"))
}

func updateDepthToWorker() {
	worker := indodax.InitWorker()
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			d := indodax.IndodaxInstance.GetDepth("eth_idr")
			worker.PushDepthUpdate(d)
		}
	}()
}

func initOrderbookWebsocket() {
	go websocket.InitBinanceHandler()
}

func setupWebsocket(app *iris.Application) {
	// create our echo websocket server
	ws := irisWs.New(irisWs.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})
	ws.OnConnection(handleConnection)
	// register the server on an endpoint.
	// see the inline javascript code in the websockets.html,
	// this endpoint is used to connect to the server.
	app.Get("/echo", ws.Handler())
	// serve the javascript built'n client-side library,
	// see websockets.html script tags, this path is used.
	app.Any("/iris-ws.js", irisWs.ClientHandler())
}

func handleConnection(c irisWs.Connection) {
	ticker := time.NewTicker(1 * time.Second)
	binOrderBook := orderbook.Exchanges[orderbook.Binance].Books[orderbook.BTC_USDC]
	idxOrderBook := indodax.GetOB()
	go func() {
		for range ticker.C {
			if !binOrderBook.Empty() {
				// bids := make(map[int64]float64)
				// bidIter := binOrderBook.IteratorBuySide()
				// asks := make(map[int64]float64)
				// askIter := binOrderBook.IteratorSellSide()

				// okBid := bidIter.Next()
				// for okBid {
				// 	o := bidIter.Value().(orderbook.Order)
				// 	bids[int64(o.Price*1000000)] = o.Qty
				// 	okBid = bidIter.Next()
				// }
				// okAsk := askIter.Next()
				// for okAsk {
				// 	o := askIter.Value().(orderbook.Order)
				// 	asks[int64(o.Price*1000000)] = o.Qty
				// 	okAsk = askIter.Next()
				// }
				// c.Emit("bin_orderbook", map[string]map[int64]float64{
				// 	"Bids": bids,
				// 	"Asks": asks,
				// })
				c.Emit("bin_orderbook_buy", binOrderBook.TopPriceBuySide())
				c.Emit("bin_orderbook_sell", binOrderBook.LowPriceSellSide())
			}
			if !idxOrderBook.Empty() {
				c.Emit("idx_orderbook_buy", idxOrderBook.TopPriceBuySide())
				c.Emit("idx_orderbook_sell", idxOrderBook.LowPriceSellSide())
			}
		}
	}()
}
