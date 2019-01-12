package main

import (
	"fmt"
	"os"
	"time"

	"github.com/anthonychristian/crypto-arbitrage/indodax"
	"github.com/anthonychristian/crypto-arbitrage/orderbook"
	"github.com/anthonychristian/crypto-arbitrage/trade"
	"github.com/anthonychristian/crypto-arbitrage/websocket"
	"github.com/joho/godotenv"
	"github.com/kataras/iris"
	irisWs "github.com/kataras/iris/websocket"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	orderbook.InitExchanges()
	// initialize API gateway
	_ = indodax.InitIndodax(os.Getenv("IDX_API_KEY"), os.Getenv("IDX_SECRET_KEY"))
	initOrderbookWebsocket()

	initArbitrageWorker()
}

func main() {
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
	websocket.InitBinanceHandler()
	indodax.InitAllWorkers()
}

func initArbitrageWorker() {
	trade.InitEthUsdtIdr()
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
	binOrderBook := orderbook.Exchanges[orderbook.Binance].Books[orderbook.ETH_USDT]
	idxOrderBook := orderbook.Exchanges[orderbook.Indodax].Books[orderbook.ETH_IDR]
	idxOrderBookUSDT := orderbook.Exchanges[orderbook.Indodax].Books[orderbook.USDT_IDR]
	go func() {
		for range ticker.C {
			if !binOrderBook.Empty() {
				c.Emit("bin_orderbook_buy", binOrderBook.TopPriceBuySide())
				c.Emit("bin_orderbook_sell", binOrderBook.LowPriceSellSide())
			}
			if !idxOrderBook.Empty() {
				c.Emit("idx_orderbook_buy", idxOrderBook.TopPriceBuySide())
				c.Emit("idx_orderbook_sell", idxOrderBook.LowPriceSellSide())
			}
			if !idxOrderBookUSDT.Empty() {
				c.Emit("idx_orderbook_buy_usdt", idxOrderBookUSDT.TopPriceBuySide())
				c.Emit("idx_orderbook_sell_usdt", idxOrderBookUSDT.LowPriceSellSide())
			}
		}
	}()

	go func() {
		for {
			select {
			case x := <-trade.TradeUpdate:
				c.Emit("arbit_update", x)
			case x := <-trade.TradeHedge:
				c.Emit("arbit_hedge", x)
			}
		}
	}()
}
