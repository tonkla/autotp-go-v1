package types

const (
	EXC_BINANCE = "BINANCE"
	EXC_BITKUB  = "BITKUB"
	EXC_FTX     = "FTX"
	EXC_SATANG  = "SATANG"

	SIDE_BUY  = "BUY"
	SIDE_SELL = "SELL"

	ORDER_STATUS_LIMIT  = "LIMIT"
	ORDER_STATUS_OPEN   = "OPEN"
	ORDER_STATUS_CLOSED = "CLOSED"
)

type Ticker struct {
	Exchange Exchange
	Symbol   string
	Price    float64
	Qty      float64
	Time     int64
}

type HisPrice struct {
	Symbol string
	Time   int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
}

type Exchange struct {
	Name string
}

type Order struct {
	Time     int64
	Exchange Exchange
	Symbol   string
	Price    float64
	TP       float64
	Qty      float64
	Side     string
	Status   string
}

type OrderBook struct {
	Exchange Exchange
	Symbol   string
	Bids     []Order
	Asks     []Order
}

type GridParams struct {
	LowerPrice   float64
	UpperPrice   float64
	Grids        int64
	Qty          float64
	TriggerPrice float64
	View         string
}
