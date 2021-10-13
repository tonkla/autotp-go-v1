package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	rds "github.com/tonkla/autotp/db"
	binance "github.com/tonkla/autotp/exchange/binance/futures"
	h "github.com/tonkla/autotp/helper"
	strategy "github.com/tonkla/autotp/strategy/trend"
	t "github.com/tonkla/autotp/types"
)

var rootCmd = &cobra.Command{
	Use:   "autotp",
	Short: "AutoTP: Auto Take Profit (Trend)",
	Long:  "AutoTP: Auto Trading Platform (Trend)",
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
	autoSL := viper.GetBool("autoSL")
	autoTP := viper.GetBool("autoTP")
	quoteSL := viper.GetFloat64("quoteSL")
	quoteTP := viper.GetFloat64("quoteTP")
	atrSL := viper.GetFloat64("atrSL")
	atrTP := viper.GetFloat64("atrTP")
	minGap := viper.GetFloat64("minGap")
	orderType := viper.GetString("orderType")
	view := viper.GetString("view")

	slStop := viper.GetInt64("slStop")
	slLimit := viper.GetInt64("slLimit")
	tpStop := viper.GetInt64("tpStop")
	tpLimit := viper.GetInt64("tpLimit")
	openLimit := viper.GetInt64("openLimit")

	db := rds.Connect(dbName)

	exchange := binance.NewFuturesClient(apiKey, secretKey)

	bp := t.BotParams{
		BotID:       botID,
		PriceDigits: priceDigits,
		QtyDigits:   qtyDigits,
		BaseQty:     baseQty,
		QuoteQty:    quoteQty,
		MATimeframe: maTimeframe,
		MAPeriod:    maPeriod,
		AutoSL:      autoSL,
		AutoTP:      autoTP,
		QuoteSL:     quoteSL,
		QuoteTP:     quoteTP,
		AtrSL:       atrSL,
		AtrTP:       atrTP,
		MinGap:      minGap,
		View:        view,
	}
	if orderType == t.OrderTypeLimit {
		bp.StopLimit = t.StopLimit{
			SLStop:    slStop,
			SLLimit:   slLimit,
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

	h.Logf("{Exchange:BiFu Symbol:%s BotID:%d Strategy:TF}\n", symbol, botID)

	for range time.Tick(time.Duration(intervalSec) * time.Second) {
		ticker := exchange.GetTicker(symbol)
		if ticker == nil || ticker.Price <= 0 {
			continue
		}

		if startPrice > 0 && ticker.Price > startPrice && len(db.GetActiveOrders(queryOrder)) == 0 {
			continue
		}

		var _period int64 = 50
		const n int64 = 4
		if maPeriod*n > _period {
			_period *= (n - 1)
		}
		hprices := exchange.GetHistoricalPrices(symbol, maTimeframe, int(_period))
		if len(hprices) == 0 || hprices[0].Open == 0 {
			continue
		}

		tradeOrders := strategy.OnTick(strategy.OnTickParams{
			DB:        *db,
			Ticker:    *ticker,
			BotParams: bp,
			HPrices:   hprices,
			IsFutures: true,
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
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		o.RefID = exo.RefID
		o.OpenTime = exo.OpenTime
		err = p.db.CreateOrder(o)
		if err != nil {
			h.Log(err)
			return
		}

		h.LogNewF(&o)
	}

	for _, o := range p.tradeOrders.OpenOrders {
		exo, err := p.exchange.PlaceLimitOrder(o)
		if err != nil || exo == nil {
			fmt.Fprintln(os.Stderr, err)
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

		h.LogNewF(&o)
	}

	syncSLLongOrders(p)
	syncSLShortOrders(p)
	syncTPLongOrders(p)
	syncTPShortOrders(p)
	syncLimitLongOrders(p)
	syncLimitShortOrders(p)
}

func syncSLLongOrders(p *params) {
	slo := p.db.GetHighestSLLongOrder(*p.queryOrder)
	if slo == nil {
		return
	}
	exo, err := p.exchange.GetOrder(*slo)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if slo.Status != exo.Status {
		slo.Status = exo.Status
		slo.UpdateTime = exo.UpdateTime
		err := p.db.UpdateOrder(*slo)
		if err != nil {
			h.Log(err)
			return
		}

		if exo.Status == t.OrderStatusFilled {
			h.LogFilledF(slo)
		}

		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(slo)
		}
	}

	if p.ticker.Price < slo.OpenPrice && slo.CloseTime == 0 {
		o := p.db.GetOrderByID(slo.OpenOrderID)
		if o == nil {
			slo.CloseTime = h.Now13()
			err = p.db.UpdateOrder(*slo)
			if err != nil {
				h.Log(err)
			}
			return
		}

		o.CloseOrderID = slo.ID
		o.ClosePrice = slo.OpenPrice
		o.CloseTime = h.Now13()
		o.PL = h.NormalizeDouble(((o.ClosePrice-o.OpenPrice)*slo.Qty)-o.Commission-slo.Commission, p.priceDigits)
		err := p.db.UpdateOrder(*o)
		if err != nil {
			h.Log(err)
			return
		}

		slo.CloseTime = o.CloseTime
		err = p.db.UpdateOrder(*slo)
		if err != nil {
			h.Log(err)
			return
		}

		h.LogClosedF(o, slo)
	}
}

func syncSLShortOrders(p *params) {
	slo := p.db.GetLowestSLShortOrder(*p.queryOrder)
	if slo == nil {
		return
	}
	exo, err := p.exchange.GetOrder(*slo)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if slo.Status != exo.Status {
		slo.Status = exo.Status
		slo.UpdateTime = exo.UpdateTime
		err := p.db.UpdateOrder(*slo)
		if err != nil {
			h.Log(err)
			return
		}

		if exo.Status == t.OrderStatusFilled {
			h.LogFilledF(slo)
		}

		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(slo)
		}
	}

	if p.ticker.Price > slo.OpenPrice && slo.CloseTime == 0 {
		o := p.db.GetOrderByID(slo.OpenOrderID)
		if o == nil {
			slo.CloseTime = h.Now13()
			err = p.db.UpdateOrder(*slo)
			if err != nil {
				h.Log(err)
			}
			return
		}

		o.CloseOrderID = slo.ID
		o.ClosePrice = slo.OpenPrice
		o.CloseTime = h.Now13()
		o.PL = h.NormalizeDouble(((o.ClosePrice-o.OpenPrice)*slo.Qty)-o.Commission-slo.Commission, p.priceDigits)
		err := p.db.UpdateOrder(*o)
		if err != nil {
			h.Log(err)
			return
		}

		slo.CloseTime = o.CloseTime
		err = p.db.UpdateOrder(*slo)
		if err != nil {
			h.Log(err)
			return
		}

		h.LogClosedF(o, slo)
	}
}

func syncTPLongOrders(p *params) {
	tpo := p.db.GetLowestTPLongOrder(*p.queryOrder)
	if tpo == nil {
		return
	}
	exo, err := p.exchange.GetOrder(*tpo)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
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
			h.LogFilledF(tpo)
		}

		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(tpo)
		}
	}

	if p.ticker.Price > tpo.OpenPrice && tpo.CloseTime == 0 {
		o := p.db.GetOrderByID(tpo.OpenOrderID)
		if o == nil {
			tpo.CloseTime = h.Now13()
			err = p.db.UpdateOrder(*tpo)
			if err != nil {
				h.Log(err)
			}
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

		h.LogClosedF(o, tpo)
	}
}

func syncTPShortOrders(p *params) {
	tpo := p.db.GetHighestTPShortOrder(*p.queryOrder)
	if tpo == nil {
		return
	}
	exo, err := p.exchange.GetOrder(*tpo)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
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
			h.LogFilledF(tpo)
		}

		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(tpo)
		}
	}

	if p.ticker.Price < tpo.OpenPrice && tpo.CloseTime == 0 {
		o := p.db.GetOrderByID(tpo.OpenOrderID)
		if o == nil {
			tpo.CloseTime = h.Now13()
			err = p.db.UpdateOrder(*tpo)
			if err != nil {
				h.Log(err)
			}
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

		h.LogClosedF(o, tpo)
	}
}

func syncLimitLongOrders(p *params) {
	o := p.db.GetHighestNewLongOrder(*p.queryOrder)
	if o == nil {
		return
	}
	exo, err := p.exchange.GetOrder(*o)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
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
			h.LogFilledF(o)
		}
		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(o)
		}
	}
}

func syncLimitShortOrders(p *params) {
	o := p.db.GetLowestNewShortOrder(*p.queryOrder)
	if o == nil {
		return
	}
	exo, err := p.exchange.GetOrder(*o)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
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
			h.LogFilledF(o)
		}
		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(o)
		}
	}
}

func placeAsTaker(p *params) {
}
