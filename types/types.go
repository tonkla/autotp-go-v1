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

	OrderPosSideLong  = "LONG"
	OrderPosSideShort = "SHORT"

	OrderTypeLimit  = "LIMIT"
	OrderTypeMarket = "MARKET"
	OrderTypeSL     = "STOP_LOSS_LIMIT"
	OrderTypeTP     = "TAKE_PROFIT_LIMIT"
	OrderTypeFSL    = "STOP"
	OrderTypeFTP    = "TAKE_PROFIT"

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
	PosSide  string `gorm:"index"`
	Type     string `gorm:"index"`
	Status   string `gorm:"index"`

	Qty        float64
	ClosePrice float64
	OpenPrice  float64
	ZonePrice  float64
	StopPrice  float64 `gorm:"-"`
	TPPrice    float64
	SLPrice    float64
	PL         float64
	Commission float64

	OpenOrderID  string `gorm:"index"`
	CloseOrderID string `gorm:"index"`

	CloseTime  int64 `gorm:"index"`
	OpenTime   int64
	UpdateTime int64

	// OpenOrder *Order `gorm:"references:OpenOrderID"`
	// CloseOrder  *Order  `gorm:"foreignKey:CloseOrderID"`
	// CloseOrders []Order `gorm:"foreignKey:OpenOrderID"`
}

type LogOpenOrder struct {
	Action string
	Qty    float64
	Open   float64
	Zone   float64
	TP     float64
}

type LogOpenFOrder struct {
	Action  string
	PosSide string
	Qty     float64
	Open    float64
}

type LogTPOrder struct {
	Action string
	Qty    float64
	Stop   float64
	Close  float64
	Open   float64
	Zone   float64
	Profit float64
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
	UpperPrice  float64
	LowerPrice  float64
	GridSize    float64
	GridTP      float64
	ApplyTrend  bool
	OpenZones   int64
	PriceDigits int64
	QtyDigits   int64
	BaseQty     float64
	QuoteQty    float64
	View        string
	Slippage    float64
	MATimeframe string
	MAPeriod    int64
	MoS         float64
	AutoSL      bool
	AutoTP      bool
	QuoteSL     float64
	QuoteTP     float64
	AtrSL       float64
	AtrTP       float64
	MinGap      float64
	StopLimit   StopLimit
}

type StopLimit struct {
	SLStop    int64
	SLLimit   int64
	TPStop    int64
	TPLimit   int64
	OpenLimit int64
}
