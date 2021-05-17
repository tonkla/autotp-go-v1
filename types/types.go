package types

const (
	EXC_BINANCE = "BINANCE"
	EXC_BITKUB  = "BITKUB"
	EXC_SATANG  = "SATANG"

	SIDE_BUY  = "Buy"
	SIDE_SELL = "Sell"
)

type Ticker struct {
	Exchange Exchange
	Symbol   string
	Price    float64
	Qty      float64
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
	Symbol string
	Side   string
	Price  float64
	Qty    float64
	TP     float64
}

type OrderBook struct {
	Exchange Exchange
	Symbol   string
	Bids     []Order
	Asks     []Order
}

type Record struct {
}

type TradeResult struct {
	Time  int64
	RefID string
}
