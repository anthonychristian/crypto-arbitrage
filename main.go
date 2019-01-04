package main

import (
	"fmt"
	"time"

	"github.com/anthonychristian/crypto-arbitrage/orderbook"
	"github.com/anthonychristian/crypto-arbitrage/trade"
	"github.com/anthonychristian/crypto-arbitrage/websocket"
	"github.com/joho/godotenv"
	"github.com/kataras/iris"
	irisWs "github.com/kataras/iris/websocket"
)

func init() {
	trade.InitExchanges()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	app := iris.New()
	app.Get("/", func(ctx iris.Context) {
		ctx.ServeFile("view/websockets.html", false)
	})

	// Using iris websocket to show orderbook updates (for testing purposes)
	// Open Localhost 8080 to start orderbook websocket
	setupWebsocket(app)
	app.Run(iris.Addr(":8080"))
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
		}
	}()
}
