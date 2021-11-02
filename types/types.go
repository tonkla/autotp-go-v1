package types

const (
	ExcBinance = "BINANCE"
	ExcBitkub  = "BITKUB"
	ExcFTX     = "FTX"
	ExcSatang  = "SATANG"

	ProductSpot    = "SPOT"
	ProductFutures = "FUTURES"

	StrategyDaily    = "DAILY"
	StrategyGrid     = "GRID"
	StrategyScalping = "SCALPING"
	StrategyTrend    = "TREND"

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
	Exchange string `gorm:"index"`
	Symbol   string `gorm:"index"`
	BotID    int64  `gorm:"index"`
	Side     string `gorm:"index"`
	PosSide  string `gorm:"index"`
	Type     string `gorm:"index"`
	Status   string `gorm:"index"`

	Qty        float64
	ClosePrice float64
	OpenPrice  float64
	ZonePrice  float64
	StopPrice  float64 `gorm:"-"`
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

type QueryOrder struct {
	ID        string
	RefID     string
	Exchange  string
	Symbol    string
	BotID     int64
	Side      string
	PosSide   string
	Type      string
	Status    string
	ZonePrice float64
	OpenPrice float64
	OpenTime  int64
	Qty       float64
}

type LogOpenOrder struct {
	Action string
	Type   string
	Qty    float64
	Open   float64
	Zone   float64
}

type LogOpenFOrder struct {
	Action  string
	Type    string
	PosSide string
	Qty     float64
	Open    float64
}

type LogCloseOrder struct {
	Action string
	Type   string
	Qty    float64
	Open   float64
	Zone   float64
	Close  float64
	Profit float64
}

type LogCloseFOrder struct {
	Action  string
	Type    string
	PosSide string
	Qty     float64
	Open    float64
	Close   float64
	Profit  float64
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
	ApiKey    string
	SecretKey string
	DbName    string
	OrderType string
	View      string

	IntervalSec int64

	Exchange    string
	Symbol      string
	BotID       int64
	Product     string
	Strategy    string
	PriceDigits int64
	QtyDigits   int64
	BaseQty     float64
	QuoteQty    float64

	StartPrice float64
	UpperPrice float64
	LowerPrice float64
	GridSize   float64
	GridTP     float64
	OpenZones  int64
	ApplyTA    bool
	Slippage   float64

	MATf1st     string
	MAPeriod1st int64
	MATf2nd     string
	MAPeriod2nd int64
	MATf3rd     string
	MAPeriod3rd int64
	OrderGap    float64
	MoS         float64

	AutoSL  bool
	AutoTP  bool
	QuoteSL float64
	QuoteTP float64
	AtrSL   float64
	AtrTP   float64

	CloseLong  bool
	CloseShort bool

	SLim StopLimit
}

type StopLimit struct {
	SLStop    int64
	SLLimit   int64
	TPStop    int64
	TPLimit   int64
	OpenLimit int64
}

type APIClient struct {
	BaseURL   string
	APIKey    string
	SecretKey string
}
