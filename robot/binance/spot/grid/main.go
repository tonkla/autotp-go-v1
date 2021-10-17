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
	"github.com/tonkla/autotp/robot"
	"github.com/tonkla/autotp/strategy/grid"
	t "github.com/tonkla/autotp/types"
)

var rootCmd = &cobra.Command{
	Use:   "autotp",
	Short: "AutoTP: Auto Take Profit (Grid)",
	Long:  "AutoTP: Auto Trading Platform (Grid)",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var (
	configFile string
)

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
	upperPrice := viper.GetFloat64("upperPrice")
	lowerPrice := viper.GetFloat64("lowerPrice")
	startPrice := viper.GetFloat64("startPrice")
	gridSize := viper.GetFloat64("gridSize")
	gridTP := viper.GetFloat64("gridTP")
	openZones := viper.GetInt64("openZones")
	baseQty := viper.GetFloat64("baseQty")
	quoteQty := viper.GetFloat64("quoteQty")
	intervalSec := viper.GetInt64("intervalSec")
	autoTP := viper.GetBool("autoTP")
	orderType := viper.GetString("orderType")

	tpStop := viper.GetInt64("tpStop")
	tpLimit := viper.GetInt64("tpLimit")
	openLimit := viper.GetInt64("openLimit")

	if upperPrice <= lowerPrice {
		fmt.Fprintln(os.Stderr, "The upper price must be greater than the lower price")
		os.Exit(0)
	} else if gridSize < 2 {
		fmt.Fprintln(os.Stderr, "Grid size must be greater than 1")
		os.Exit(0)
	} else if baseQty == 0 && quoteQty == 0 {
		fmt.Fprintln(os.Stderr, "Quantity per grid must be greater than 0")
		os.Exit(0)
	}

	db := rds.Connect(dbName)

	exchange := binance.NewSpotClient(apiKey, secretKey)

	slim := t.StopLimit{
		TPStop:    tpStop,
		TPLimit:   tpLimit,
		OpenLimit: openLimit,
	}

	bp := t.BotParams{
		BotID:       botID,
		UpperPrice:  upperPrice,
		LowerPrice:  lowerPrice,
		GridSize:    gridSize,
		GridTP:      gridTP,
		OpenZones:   openZones,
		PriceDigits: priceDigits,
		QtyDigits:   qtyDigits,
		BaseQty:     baseQty,
		QuoteQty:    quoteQty,
		AutoTP:      autoTP,
		View:        "LONG",
		SLim:        slim,
	}

	qo := t.QueryOrder{
		BotID:    botID,
		Exchange: t.ExcBinance,
		Symbol:   symbol,
	}

	_params := robot.AppParams{
		DB:   *db,
		Repo: exchange,
		QO:   qo,
		BP:   bp,
	}

	if intervalSec == 0 {
		intervalSec = 3
	}

	h.Logf("{Exchange:BinanceSpot Symbol:%s BotID:%d}\n", symbol, botID)

	for range time.Tick(time.Duration(intervalSec) * time.Second) {
		ticker := exchange.GetTicker(symbol)
		if ticker == nil || ticker.Price <= 0 {
			continue
		}

		if startPrice > 0 && ticker.Price > startPrice && len(db.GetActiveOrders(qo)) == 0 {
			continue
		}

		tradeOrders := grid.OnTick(grid.OnTickParams{
			DB:        db,
			Ticker:    ticker,
			BotParams: bp,
		})
		if tradeOrders == nil {
			continue
		}

		_params.Ticker = *ticker
		_params.TO = *tradeOrders
		if orderType == t.OrderTypeLimit {
			placeAsMaker(&_params)
		} else if orderType == t.OrderTypeMarket {
			placeAsTaker(&_params)
		}
	}
}

func placeAsMaker(p *robot.AppParams) {
	openNewOrders(p)
	syncHighestNewOrder(p)
	syncLowestFilledOrder(p)
	syncLowestTPOrder(p)
}

func openNewOrders(p *robot.AppParams) {
	for _, o := range p.TO.OpenOrders {
		if o.OpenPrice < h.CalcStopBehindTicker(t.OrderSideBuy, p.Ticker.Price, float64(p.BP.SLim.OpenLimit), p.BP.PriceDigits) {
			return
		}

		exo, err := p.Repo.OpenLimitOrder(o)
		if err != nil || exo == nil {
			h.Log("OpenOrder")
			os.Exit(1)
		}

		o.RefID = exo.RefID
		o.Status = exo.Status
		o.OpenPrice = exo.OpenPrice
		o.OpenTime = exo.OpenTime
		err = p.DB.CreateOrder(o)
		if err != nil {
			h.Log("CreateOrder", err)
			continue
		}
		h.LogNew(&o)
	}
}

func syncHighestNewOrder(p *robot.AppParams) {
	// Synchronize order status
	o := p.DB.GetHighestNewBuyOrder(p.QO)
	if o == nil {
		return
	}
	exo, err := p.Repo.GetOrder(*o)
	if err != nil || exo == nil {
		h.Log("GetOrder")
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	// Synchronize FILLED/CANCELED order
	if o.Status != exo.Status {
		o.Status = exo.Status
		o.UpdateTime = exo.UpdateTime
		err := p.DB.UpdateOrder(*o)
		if err != nil {
			h.Log("UpdateOrder", err)
			return
		}
		if exo.Status == t.OrderStatusFilled {
			h.LogFilled(o)
		}
		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceled(o)
		}
	}
}

func syncLowestFilledOrder(p *robot.AppParams) {
	// Place a new Take Profit order
	o := p.DB.GetLowestFilledBuyOrder(p.QO)
	if o != nil && o.TPPrice > 0 && p.DB.GetTPOrder(o.ID) == nil {
		// Keep only one TP order, because Binance has a 'MAX_NUM_ALGO_ORDERS=5'
		tpOrders := p.DB.GetNewTPOrders(p.QO)
		if len(tpOrders) > 0 {
			tpo := tpOrders[0]
			// Ignore when the order TP price is so far, keep calm and waiting
			if tpo.OpenPrice < o.TPPrice {
				return
			}
			exo, err := p.Repo.CancelOrder(tpo)
			if err != nil || exo == nil {
				h.Log("CancelOrder")
				os.Exit(1)
			}

			tpo.Status = exo.Status
			tpo.UpdateTime = exo.UpdateTime
			err = p.DB.UpdateOrder(tpo)
			if err != nil {
				h.Log(err)
				return
			}
			h.LogCanceled(&tpo)
		}

		if p.Ticker.Price < h.CalcTPStop(o.Side, o.TPPrice, float64(p.BP.SLim.TPStop), p.BP.PriceDigits) {
			return
		}
		stopPrice := h.CalcTPStop(o.Side, o.TPPrice, float64(p.BP.SLim.TPLimit), p.BP.PriceDigits)

		// The price moves so fast
		if p.Ticker.Price > stopPrice && o.CloseOrderID == "" {
			o.CloseOrderID = "0"
			o.ClosePrice = o.TPPrice
			o.CloseTime = h.Now13()
			o.PL = h.NormalizeDouble(((o.ClosePrice - o.OpenPrice) * o.Qty), p.BP.PriceDigits)
			err := p.DB.UpdateOrder(*o)
			if err != nil {
				h.Log(err)
			}
			return
		}

		tpo := t.Order{
			BotID:       o.BotID,
			Exchange:    o.Exchange,
			Symbol:      o.Symbol,
			ID:          h.GenID(),
			OpenOrderID: o.ID,
			Qty:         o.Qty,
			Side:        h.Reverse(o.Side),
			Type:        t.OrderTypeTP,
			Status:      t.OrderStatusNew,
			StopPrice:   stopPrice,
			OpenPrice:   o.TPPrice,
		}
		exo, err := p.Repo.OpenStopOrder(tpo)
		if err != nil || exo == nil {
			h.Log("PlaceTPOrder")
			os.Exit(1)
		}

		tpo.RefID = exo.RefID
		tpo.OpenTime = exo.OpenTime
		err = p.DB.CreateOrder(tpo)
		if err != nil {
			h.Log(err)
			return
		}
		h.LogNew(&tpo)
	}
}

func syncLowestTPOrder(p *robot.AppParams) {
	tpo := p.DB.GetLowestTPOrder(p.QO)
	if tpo == nil {
		return
	}
	exo, err := p.Repo.GetOrder(*tpo)
	if err != nil || exo == nil {
		h.Log("GetTPOrder")
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if tpo.Status != exo.Status {
		tpo.Status = exo.Status
		tpo.UpdateTime = exo.UpdateTime
		err := p.DB.UpdateOrder(*tpo)
		if err != nil {
			h.Log(err)
			return
		}
		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceled(tpo)
			return
		}
	}
	if exo.Status == t.OrderStatusCanceled {
		return
	}

	oo := p.DB.GetOrderByID(tpo.OpenOrderID)
	if oo != nil && oo.CloseOrderID == "" && p.Ticker.Price > tpo.OpenPrice {
		oo.CloseOrderID = tpo.ID
		oo.ClosePrice = tpo.OpenPrice
		oo.CloseTime = h.Now13()
		oo.PL = h.NormalizeDouble(((oo.ClosePrice-oo.OpenPrice)*tpo.Qty)-oo.Commission-tpo.Commission, p.BP.PriceDigits)
		err := p.DB.UpdateOrder(*oo)
		if err != nil {
			h.Log(err)
			return
		}
		h.LogClosed(oo, tpo)

		tpo.CloseTime = oo.CloseTime
		err = p.DB.UpdateOrder(*tpo)
		if err != nil {
			h.Log("UpdateTPOrder", err)
		}
	}
}

func placeAsTaker(p *robot.AppParams) {
	// Open new orders -----------------------------------------------------------
	for _, o := range p.TO.OpenOrders {
		book := p.Repo.GetOrderBook(p.Ticker.Symbol, 5)
		if book == nil || len(book.Asks) == 0 {
			continue
		}
		buyPrice := book.Asks[0].Price
		if buyPrice > o.ZonePrice || buyPrice == 0 {
			continue
		}

		_qty := h.NormalizeDouble(p.BP.QuoteQty/buyPrice, p.BP.QtyDigits)
		if _qty > o.Qty {
			o.Qty = _qty
		}
		o.Type = t.OrderTypeMarket
		exo, err := p.Repo.OpenMarketOrder(o)
		if err != nil || exo == nil {
			h.Log("OpenOrder")
			os.Exit(1)
		}

		o.RefID = exo.RefID
		o.Status = exo.Status
		o.OpenTime = exo.OpenTime
		o.OpenPrice = exo.OpenPrice
		o.Qty = exo.Qty
		o.Commission = exo.Commission
		err = p.DB.CreateOrder(o)
		if err != nil {
			h.Log("CreateOrder", err)
			continue
		}
		h.LogFilled(&o)
	}

	// Take Profit ---------------------------------------------------------------
	o := p.DB.GetLowestFilledBuyOrder(p.QO)
	if o != nil && o.TPPrice > 0 && p.DB.GetTPOrder(o.ID) == nil {
		book := p.Repo.GetOrderBook(p.Ticker.Symbol, 5)
		if book == nil || len(book.Bids) == 0 {
			return
		}
		sellPrice := book.Bids[0].Price
		if o.TPPrice > sellPrice || sellPrice == 0 {
			return
		}

		tpo := t.Order{
			BotID:       o.BotID,
			Exchange:    o.Exchange,
			Symbol:      o.Symbol,
			ID:          h.GenID(),
			OpenOrderID: o.ID,
			Qty:         o.Qty,
			Side:        h.Reverse(o.Side),
			Status:      t.OrderStatusNew,
			Type:        t.OrderTypeMarket,
		}
		exo, err := p.Repo.OpenMarketOrder(tpo)
		if err != nil || exo == nil {
			h.Log("TakeProfit")
			os.Exit(1)
		}

		tpo.Type = t.OrderTypeTP // Save to the local DB as a TAKE_PROFIT_LIMIT type
		tpo.RefID = exo.RefID
		tpo.Status = exo.Status
		tpo.OpenTime = exo.OpenTime
		tpo.OpenPrice = exo.OpenPrice
		tpo.Qty = exo.Qty
		tpo.Commission = exo.Commission
		tpo.CloseTime = h.Now13()
		err = p.DB.CreateOrder(tpo)
		if err != nil {
			h.Log("CreateTPOrder", err)
			return
		}

		o.CloseOrderID = tpo.ID
		o.ClosePrice = tpo.OpenPrice
		o.CloseTime = tpo.OpenTime
		o.PL = h.NormalizeDouble(((o.ClosePrice-o.OpenPrice)*tpo.Qty)-o.Commission-tpo.Commission, p.BP.PriceDigits)
		err = p.DB.UpdateOrder(*o)
		if err != nil {
			h.Log("UpdateOrder", err)
			return
		}
		h.LogClosed(o, &tpo)
	}
}
