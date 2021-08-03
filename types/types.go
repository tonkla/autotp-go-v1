package types

const (
	ExcBinance = "BINANCE"
	ExcBitkub  = "BITKUB"
	ExcFTX     = "FTX"
	ExcSatang  = "SATANG"

	OrderStatusNew      = "NEW"
	OrderStatusFilled   = "FILLED"
	OrderStatusCanceled = "CANCELED"

	OrderSideBuy  = "BUY"
	OrderSideSell = "SELL"

	OrderTypeLimit  = "LIMIT"
	OrderTypeMarket = "MARKET"
	OrderTypeSL     = "STOP_LOSS_LIMIT"
	OrderTypeTP     = "TAKE_PROFIT_LIMIT"

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
	ID       string `gorm:"index"`
	RefID    string `gorm:"index"`
	BotID    int64  `gorm:"index"`
	Exchange string `gorm:"index"`
	Symbol   string `gorm:"index"`
	Side     string `gorm:"index"`
	Type     string `gorm:"index"`
	Status   string `gorm:"index"`
	Qty      float64

	OpenPrice  float64
	StopPrice  float64 `gorm:"-"`
	SLPrice    float64
	TPPrice    float64
	OpenTime   int64
	UpdateTime int64

	OpenOrderID string `gorm:"index"`
	OpenOrder   *Order `gorm:"foreignKey:OpenOrderID"`

	CloseOrderID string `gorm:"index"`
	CloseOrder   *Order `gorm:"foreignKey:CloseOrderID"`
	ClosePrice   float64
	CloseTime    int64

	CloseOrders []Order `gorm:"foreignKey:OpenOrderID"`
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
	Symbol string
	Bids   []ExOrder
	Asks   []ExOrder
}

type BotParams struct {
	BotID       int64
	LowerPrice  float64
	UpperPrice  float64
	GridSize    float64
	GridTP      float64
	FollowTrend bool
	OpenAll     bool
	Qty         float64
	View        string
	Slippage    float64
	MATimeframe string
	MAPeriod    int64
	AutoSL      bool
	AutoTP      bool
}
