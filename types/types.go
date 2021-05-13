package types

const (
	EXC_BINANCE = "BINANCE"
	EXC_BITKUB  = "BITKUB"
	EXC_SATANG  = "SATANG"
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
	Side  string
	Price float64
	Qty   float64
}

type OrderBook struct {
	Exchange Exchange
	Symbol   string
	Bids     []Order
	Asks     []Order
}

type Advice struct {
	Symbol string
	Side   string
	Price  float64
	Qty    float64
}

type TradeResult struct {
	Time   int64
	Symbol string
	Side   string
	Price  float64
	Qty    float64
}
