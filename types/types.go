package types

const (
	EXC_BINANCE = "BINANCE"
	EXC_BITKUB  = "BITKUB"
	EXC_FTX     = "FTX"
	EXC_SATANG  = "SATANG"

	ORDER_STATUS_LIMIT  = "LIMIT"
	ORDER_STATUS_OPEN   = "OPEN"
	ORDER_STATUS_CLOSED = "CLOSED"

	SIDE_BUY  = "BUY"
	SIDE_SELL = "SELL"

	TREND_NO     = 0
	TREND_UP_1   = 1
	TREND_UP_2   = 2
	TREND_UP_3   = 3
	TREND_UP_4   = 4
	TREND_UP_5   = 5
	TREND_DOWN_1 = -1
	TREND_DOWN_2 = -2
	TREND_DOWN_3 = -3
	TREND_DOWN_4 = -4
	TREND_DOWN_5 = -5

	VIEW_NEUTRAL = "NEUTRAL"
	VIEW_LONG    = "LONG"
	VIEW_SHORT   = "SHORT"
)

type Ticker struct {
	Exchange string
	Symbol   string
	Price    float64
	Qty      float64
	Time     int64
}

type HistoricalPrice struct {
	Symbol string
	Time   int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
}

type Order struct {
	ID         uint   `gorm:"primaryKey"`
	CreatedAt  int64  `gorm:"autoCreateTime"`
	UpdatedAt  int64  `gorm:"autoUpdateTime"`
	BotID      int64  `gorm:"index"`
	Exchange   string `gorm:"index"`
	Symbol     string `gorm:"index"`
	OpenPrice  float64
	ClosePrice float64
	OpenTime   int64
	CloseTime  int64
	SL         float64
	TP         float64
	Qty        float64
	Side       string `gorm:"index"`
	Status     string `gorm:"index"`
}

type TradeOrders struct {
	OpenOrders  []Order
	CloseOrders []Order
}

type ExOrder struct {
	Symbol string
	Price  float64
	Qty    float64
	Side   string
}

type OrderBook struct {
	Exchange string
	Symbol   string
	Bids     []ExOrder
	Asks     []ExOrder
}

type BotParams struct {
	BotID        int64
	LowerPrice   float64
	UpperPrice   float64
	Grids        float64
	Qty          float64
	View         string
	SL           float64
	TP           float64
	TriggerPrice float64
	Slippage     float64
	MATimeframe  string
	MAPeriod     int64
}
