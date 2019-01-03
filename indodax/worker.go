package indodax

// Worker is the main engine for making order decisions
// and continuing arbitrage loop ETH -> IDR -> USDT
type Worker struct {
	depth chan Depth
	halt  bool
}

var WorkerInstance *Worker

// InitWorker instances
func InitWorker() *Worker {
	newWorker := &Worker{
		depth: make(chan Depth),
	}
	return newWorker
}

// Halt to halt the worker from doing actions
func (w *Worker) Halt() {
	w.halt = true
}

// Start to start the worker to do actions
func (w *Worker) Start() {
	w.halt = false
}

func (w *Worker) work() {
	// infinite loop to keep doing actions
	for {
		select {
		case d := <-w.depth:
			// add depth to orderbook
			updateDepth(d)
		}
	}
}

func updateDepth(d Depth) {
	// TODO update the indodax's orderbook
}
