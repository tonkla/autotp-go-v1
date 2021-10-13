package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	rds "github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/exchange/binance"
	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/strategy/daily"
	t "github.com/tonkla/autotp/types"
)

var rootCmd = &cobra.Command{
	Use:   "autotp",
	Short: "AutoTP: Auto Take Profit (Daily)",
	Long:  "AutoTP: Auto Trading Platform (Daily)",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var (
	configFile string
)

type params struct {
	db          *rds.DB
	ticker      *t.Ticker
	tradeOrders *t.TradeOrders
	exchange    *binance.Client
	queryOrder  *t.Order
	symbol      string
	priceDigits int64
	qtyDigits   int64
	baseQty     float64
	quoteQty    float64
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "configFile", "c", "", "Configuration File (required)")
	rootCmd.MarkFlagRequired("configFile")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	} else if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	} else if ext := path.Ext(configFile); ext != ".yml" && ext != ".yaml" {
		fmt.Fprintln(os.Stderr, "Accept only YAML file")
		os.Exit(0)
	}

	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	}

	apiKey := viper.GetString("apiKey")
	secretKey := viper.GetString("secretKey")
	dbName := viper.GetString("dbName")
	symbol := viper.GetString("symbol")
	botID := viper.GetInt64("botID")
	priceDigits := viper.GetInt64("priceDigits")
	qtyDigits := viper.GetInt64("qtyDigits")
	startPrice := viper.GetFloat64("startPrice")
	baseQty := viper.GetFloat64("baseQty")
	quoteQty := viper.GetFloat64("quoteQty")
	intervalSec := viper.GetInt64("intervalSec")
	maTimeframe := viper.GetString("maTimeframe")
	maPeriod := viper.GetInt64("maPeriod")
	autoTP := viper.GetBool("autoTP")
	mos := viper.GetFloat64("mos")
	atrTP := viper.GetFloat64("atrTP")
	orderType := viper.GetString("orderType")

	tpStop := viper.GetInt64("tpStop")
	tpLimit := viper.GetInt64("tpLimit")
	openLimit := viper.GetInt64("openLimit")

	db := rds.Connect(dbName)

	exchange := binance.NewSpotClient(apiKey, secretKey)

	bp := t.BotParams{
		BotID:       botID,
		PriceDigits: priceDigits,
		QtyDigits:   qtyDigits,
		BaseQty:     baseQty,
		QuoteQty:    quoteQty,
		MATimeframe: maTimeframe,
		MAPeriod:    maPeriod,
		AutoTP:      autoTP,
		AtrTP:       atrTP,
		MoS:         mos,
		View:        "LONG",
	}
	if orderType == t.OrderTypeLimit {
		bp.StopLimit = t.StopLimit{
			TPStop:    tpStop,
			TPLimit:   tpLimit,
			OpenLimit: openLimit,
		}
	}

	queryOrder := t.Order{
		BotID:    botID,
		Exchange: t.ExcBinance,
		Symbol:   symbol,
	}

	_params := params{
		db:          db,
		exchange:    &exchange,
		queryOrder:  &queryOrder,
		symbol:      symbol,
		priceDigits: priceDigits,
		qtyDigits:   qtyDigits,
		baseQty:     baseQty,
		quoteQty:    quoteQty,
	}

	if intervalSec == 0 {
		intervalSec = 3
	}

	h.Logf("{Exchange:BinanceSpot Symbol:%s BotID:%d Strategy:Daily}\n", symbol, botID)

	for range time.Tick(time.Duration(intervalSec) * time.Second) {
		ticker := exchange.GetTicker(symbol)
		if ticker == nil || ticker.Price <= 0 {
			continue
		}

		if startPrice > 0 && ticker.Price > startPrice && len(db.GetActiveOrders(queryOrder)) == 0 {
			continue
		}

		hprices := exchange.GetHistoricalPrices(symbol, maTimeframe, 50)
		if len(hprices) == 0 || hprices[0].Open == 0 {
			continue
		}

		tradeOrders := daily.OnTick(daily.OnTickParams{
			DB:        *db,
			Ticker:    *ticker,
			BotParams: bp,
			HPrices:   hprices,
		})
		if tradeOrders == nil {
			continue
		}

		_params.ticker = ticker
		_params.tradeOrders = tradeOrders
		if orderType == t.OrderTypeLimit {
			placeAsMaker(&_params)
		} else if orderType == t.OrderTypeMarket {
			placeAsTaker(&_params)
		}
	}
}

func placeAsMaker(p *params) {
	for _, o := range p.tradeOrders.CloseOrders {
		exo, err := p.exchange.PlaceStopOrder(o)
		if err != nil || exo == nil {
			os.Exit(1)
		}

		o.RefID = exo.RefID
		o.OpenTime = exo.OpenTime
		err = p.db.CreateOrder(o)
		if err != nil {
			h.Log(err)
			return
		}

		log := t.LogCloseOrder{
			Action: "NEW_TP",
			Qty:    o.Qty,
			Close:  o.OpenPrice,
		}
		h.Log(log)
	}

	for _, o := range p.tradeOrders.OpenOrders {
		exo, err := p.exchange.PlaceLimitOrder(o)
		if err != nil || exo == nil {
			os.Exit(1)
		}

		o.RefID = exo.RefID
		o.Status = exo.Status
		o.OpenTime = exo.OpenTime
		o.OpenPrice = exo.OpenPrice
		err = p.db.CreateOrder(o)
		if err != nil {
			h.Log(err)
			return
		}

		log := t.LogOpenOrder{
			Action: "NEW",
			Qty:    o.Qty,
			Open:   o.OpenPrice,
		}
		h.Log(log)
	}

	syncTPOrder(p)
	syncLimitOrder(p)
}

func syncTPOrder(p *params) {
	tpo := p.db.GetLowestTPOrder(*p.queryOrder)
	if tpo == nil {
		return
	}
	exo, err := p.exchange.GetOrder(*tpo)
	if err != nil || exo == nil {
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if tpo.Status != exo.Status {
		tpo.Status = exo.Status
		tpo.UpdateTime = exo.UpdateTime
		err := p.db.UpdateOrder(*tpo)
		if err != nil {
			h.Log(err)
			return
		}

		if exo.Status == t.OrderStatusFilled {
			log := t.LogCloseOrder{
				Action: "FILLED_TP",
				Qty:    tpo.Qty,
				Open:   tpo.OpenPrice,
			}
			h.Log(log)
		}

		if exo.Status == t.OrderStatusCanceled {
			log := t.LogCloseOrder{
				Action: "CANCELED_TP",
				Qty:    tpo.Qty,
				Open:   tpo.OpenPrice,
			}
			h.Log(log)
		}
	}

	if p.ticker.Price > tpo.OpenPrice && tpo.CloseTime == 0 {
		o := p.db.GetOrderByID(tpo.OpenOrderID)
		if o == nil {
			return
		}

		o.CloseOrderID = tpo.ID
		o.ClosePrice = tpo.OpenPrice
		o.CloseTime = h.Now13()
		o.PL = h.NormalizeDouble(((o.ClosePrice-o.OpenPrice)*tpo.Qty)-o.Commission-tpo.Commission, p.priceDigits)
		err := p.db.UpdateOrder(*o)
		if err != nil {
			h.Log(err)
			return
		}

		tpo.CloseTime = o.CloseTime
		err = p.db.UpdateOrder(*tpo)
		if err != nil {
			h.Log(err)
			return
		}

		log := t.LogCloseOrder{
			Action: "CLOSED_TP",
			Qty:    tpo.Qty,
			Close:  o.ClosePrice,
			Open:   o.OpenPrice,
			Profit: o.PL,
		}
		h.Log(log)
	}
}

func syncLimitOrder(p *params) {
	o := p.db.GetHighestNewBuyOrder(*p.queryOrder)
	if o == nil {
		return
	}
	exo, err := p.exchange.GetOrder(*o)
	if err != nil || exo == nil {
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if o.Status != exo.Status {
		o.Status = exo.Status
		o.UpdateTime = exo.UpdateTime
		err := p.db.UpdateOrder(*o)
		if err != nil {
			h.Log(err)
			return
		}
		if exo.Status == t.OrderStatusFilled {
			log := t.LogOpenOrder{
				Action: "FILLED",
				Qty:    o.Qty,
				Open:   o.OpenPrice,
			}
			h.Log(log)
		}
		if exo.Status == t.OrderStatusCanceled {
			log := t.LogOpenOrder{
				Action: "CANCELED",
				Qty:    o.Qty,
				Open:   o.OpenPrice,
			}
			h.Log(log)
		}
	}
}

func placeAsTaker(p *params) {
}
