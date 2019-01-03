package indodax

type Depth struct {
	Buy  []interface{} `json:"buy"`
	Sell []interface{} `json:"sell"`
}

type DepthPair struct {
	Price string
	Qty   string
}

// IsEmpty is a utility function to check if the depth
// is empty.
func (d *Depth) IsEmpty() bool {
	return len(d.Buy) == 0 && len(d.Sell) == 0
}
