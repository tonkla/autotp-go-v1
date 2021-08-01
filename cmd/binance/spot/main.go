package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/exchange/binance"
	h "github.com/tonkla/autotp/helper"
	strategy "github.com/tonkla/autotp/strategy/gridtrend"
	t "github.com/tonkla/autotp/types"
)

var rootCmd = &cobra.Command{
	Use:   "autotp-grid",
	Short: "AutoTP: Grid Strategy",
	Long:  "AutoTP: Grid Strategy",
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
		os.Exit(1)
	} else if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if ext := path.Ext(configFile); ext != ".yml" && ext != ".yaml" {
		fmt.Fprintln(os.Stderr, "Accept only YAML file")
		os.Exit(1)
	}

	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	apiKey := viper.GetString("apiKey")
	secretKey := viper.GetString("secretKey")
	dbName := viper.GetString("dbName")
	botID := viper.GetInt64("botID")
	symbol := viper.GetString("symbol")
	lowerPrice := viper.GetFloat64("lowerPrice")
	upperPrice := viper.GetFloat64("upperPrice")
	gridSize := viper.GetFloat64("gridSize")
	gridTP := viper.GetFloat64("gridTP")
	qty := viper.GetFloat64("qty")
	view := viper.GetString("view")
	slippage := viper.GetFloat64("slippage")
	intervalSec := viper.GetInt64("intervalSec")
	maTimeframe := viper.GetString("maTimeframe")
	maPeriod := viper.GetInt64("maPeriod")
	autoSL := viper.GetBool("autoSL")
	autoTP := viper.GetBool("autoTP")

	if upperPrice <= lowerPrice {
		fmt.Fprintln(os.Stderr, "The upper price must be greater than the lower price")
		os.Exit(1)
	} else if gridSize < 2 {
		fmt.Fprintln(os.Stderr, "Grid size must be greater than 1")
		os.Exit(1)
	} else if qty <= 0 {
		fmt.Fprintln(os.Stderr, "Quantity per grid must be greater than 0")
		os.Exit(1)
	}

	h.Logf("{Exchange: 'Binance Spot', BotID: %d, Symbol: %s}\n", botID, symbol)

	params := t.BotParams{
		BotID:       botID,
		LowerPrice:  lowerPrice,
		UpperPrice:  upperPrice,
		GridSize:    gridSize,
		GridTP:      gridTP,
		Qty:         qty,
		View:        view,
		Slippage:    slippage,
		MATimeframe: maTimeframe,
		MAPeriod:    maPeriod,
		AutoSL:      autoSL,
		AutoTP:      autoTP,
	}

	db := db.Connect(dbName)

	if intervalSec == 0 {
		intervalSec = 5
	}

	exchange := binance.NewSpotClient(apiKey, secretKey)

	for range time.Tick(time.Duration(intervalSec) * time.Second) {
		ticker := exchange.GetTicker(symbol)
		if ticker == nil || ticker.Price <= 0 {
			continue
		}

		orderBook := exchange.GetOrderBook(symbol, 5)
		if orderBook == nil {
			continue
		}

		hprices := exchange.GetHistoricalPrices(ticker.Symbol, maTimeframe, 100)
		if len(hprices) == 0 {
			continue
		}

		p := strategy.OnTickParams{
			Ticker:    *ticker,
			OrderBook: *orderBook,
			BotParams: params,
			HPrices:   hprices,
			DB:        *db,
		}

		tradeOrders := strategy.OnTick(p)
		if tradeOrders == nil {
			continue
		}

		// Open new orders ---------------------------------------------------------

		for _, o := range tradeOrders.OpenOrders {
			id, err := h.GenID(0)
			if err != nil {
				continue
			}
			o.ID = id
			exo := exchange.PlaceLimitOrder(o)
			if exo == nil {
				continue
			}
			o.RefID = exo.RefID
			o.Status = exo.Status
			o.OpenTime = exo.OpenTime
			o.OpenPrice = exo.OpenPrice
			err = db.CreateOrder(o)
			if err != nil {
				h.Log(err)
				continue
			}
			h.Logf("{Action: Open, Side: %s, Qty: %.4f, Price: %.2f}\n", o.Side, o.Qty, o.OpenPrice)
		}

		// Close orders ------------------------------------------------------------

		for _, o := range tradeOrders.CloseOrders {
			id, err := h.GenID(0)
			if err != nil {
				continue
			}
			o.ID = id
			exo := exchange.PlaceLimitOrder(o)
			if exo == nil {
				continue
			}
			o.RefID = exo.RefID
			o.Status = exo.Status
			o.OpenTime = exo.OpenTime
			o.OpenPrice = exo.OpenPrice
			err = db.CreateOrder(o)
			if err != nil {
				h.Log(err)
				continue
			}
			h.Logf("{Action: Close, Side: %s, Qty: %.4f, Price: %.2f}\n", o.Side, o.Qty, o.OpenPrice)
		}

		// Stop Loss / Take Profit -------------------------------------------------

		qo := t.Order{
			BotID:    botID,
			Exchange: t.ExcBinance,
			Symbol:   symbol,
		}
		for _, o := range db.GetActiveOrders(qo) {
			exo := exchange.GetOrder(o)
			if exo == nil || exo.Status == t.OrderStatusNew {
				continue
			}

			// Synchronize canceled order
			if exo.Status == t.OrderStatusCanceled {
				o.Status = t.OrderStatusCanceled
				err := db.UpdateOrder(o)
				if err != nil {
					h.Log(err)
				}
				continue
			}

			// Place a new Stop Loss order
			if o.SLPrice > 0 && db.GetSLOrder(o.ID) == nil {
				id, err := h.GenID(0)
				if err != nil {
					continue
				}

				slo := t.Order{
					BotID:       o.BotID,
					Exchange:    o.Exchange,
					Symbol:      o.Symbol,
					ID:          id,
					OpenOrderID: o.ID,
					Qty:         o.Qty,
					Side:        h.Reverse(o.Side),
					Type:        t.OrderTypeSL,
					OpenPrice:   o.SLPrice,
					StopPrice:   h.CalcSLStop(o.Side, o.SLPrice, 0),
				}
				exo := exchange.PlaceStopOrder(slo)
				if exo == nil {
					continue
				}
				slo.RefID = exo.RefID
				slo.Status = exo.Status
				slo.OpenTime = exo.OpenTime
				err = db.CreateOrder(slo)
				if err != nil {
					h.Log(err)
					continue
				}
				h.Logf("{Action: SL, Side: %s, Qty: %.4f, Price: %.2f, OpenPrice: %.2f, Loss: -%.2f}\n",
					slo.Side, slo.Qty, slo.OpenPrice, o.OpenPrice, math.Abs(slo.OpenPrice-o.OpenPrice)*slo.Qty)
			}

			// Place a new Take Profit order
			if o.TPPrice > 0 && db.GetTPOrder(o.ID) == nil {
				id, err := h.GenID(0)
				if err != nil {
					continue
				}

				tpo := t.Order{
					BotID:       o.BotID,
					Exchange:    o.Exchange,
					Symbol:      o.Symbol,
					ID:          id,
					OpenOrderID: o.ID,
					Qty:         o.Qty,
					Side:        h.Reverse(o.Side),
					Type:        t.OrderTypeTP,
					OpenPrice:   o.TPPrice,
					StopPrice:   h.CalcTPStop(o.Side, o.TPPrice, 0),
				}
				exo := exchange.PlaceStopOrder(tpo)
				if exo == nil {
					continue
				}
				tpo.RefID = exo.RefID
				tpo.Status = exo.Status
				tpo.OpenTime = exo.OpenTime
				err = db.CreateOrder(tpo)
				if err != nil {
					h.Log(err)
					continue
				}
				h.Logf("{Action: TP, Side: %s, Qty: %.4f, Price: %.2f, OpenPrice: %.2f, Profit: %.2f}\n",
					tpo.Side, tpo.Qty, tpo.OpenPrice, o.OpenPrice, math.Abs(tpo.OpenPrice-o.OpenPrice)*tpo.Qty)
			}
		}
	}
}
