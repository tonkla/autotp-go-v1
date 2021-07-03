package types

const (
	ExcBinance = "BINANCE"
	ExcBitkub  = "BITKUB"
	ExcFTX     = "FTX"
	ExcSatang  = "SATANG"

	OrderStatusNew    = "NEW"
	OrderStatusFilled = "FILLED"
	OrderStatusClosed = "CLOSED"

	OrderSideBuy  = "BUY"
	OrderSideSell = "SELL"

	OrderTypeLimit  = "LIMIT"
	OrderTypeMarket = "MARKET"
	OrderTypeSL     = "STOP_LOSS"
	OrderTypeTP     = "TAKE_PROFIT"

	TrendNo    = 0
	TrendUp1   = 1
	TrendUp2   = 2
	TrendUp3   = 3
	TrendUp4   = 4
	TrendUp5   = 5
	TrendDown1 = -1
	TrendDown2 = -2
	TrendDown3 = -3
	TrendDown4 = -4
	TrendDown5 = -5

	ViewNeutral = "NEUTRAL"
	ViewLong    = "LONG"
	ViewShort   = "SHORT"
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
	RefID      int64  `gorm:"index"`
	OpenTime   int64
	CloseTime  int64
	OpenPrice  float64
	ClosePrice float64
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
