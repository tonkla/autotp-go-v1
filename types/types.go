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
	Exchange string
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

type Order struct {
	ID        uint  `gorm:"primaryKey"`
	CreatedAt int64 `gorm:"autoCreateTime"`
	UpdatedAt int64 `gorm:"autoUpdateTime"`
	Time      int64
	Exchange  string `gorm:"index"`
	Symbol    string `gorm:"index"`
	Price     float64
	SL        float64
	TP        float64
	Qty       float64
	Side      string `gorm:"index"`
	Status    string `gorm:"index"`
}

type OrderBook struct {
	Exchange string
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
	SL           float64
	TP           float64
}

type Helper interface {
	DoesOrderExists(*Order) bool
}
