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
	digits := viper.GetInt64("digits")
	lowerPrice := viper.GetFloat64("lowerPrice")
	upperPrice := viper.GetFloat64("upperPrice")
	gridSize := viper.GetFloat64("gridSize")
	gridTP := viper.GetFloat64("gridTP")
	followTrend := viper.GetBool("followTrend")
	openAll := viper.GetBool("openAllZones")
	qty := viper.GetFloat64("qty")
	view := viper.GetString("view")
	slippage := viper.GetFloat64("slippage")
	intervalSec := viper.GetInt64("intervalSec")
	maTimeframe := viper.GetString("maTimeframe")
	maPeriod := viper.GetInt64("maPeriod")
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
		FollowTrend: followTrend,
		OpenAll:     openAll,
		Qty:         qty,
		View:        view,
		Slippage:    slippage,
		MATimeframe: maTimeframe,
		MAPeriod:    maPeriod,
		AutoTP:      autoTP,
	}

	db := db.Connect(dbName)

	exchange := binance.NewSpotClient(apiKey, secretKey)

	if intervalSec == 0 {
		intervalSec = 5
	}

	for range time.Tick(time.Duration(intervalSec) * time.Second) {
		ticker := exchange.GetTicker(symbol)
		if ticker == nil || ticker.Price <= 0 {
			continue
		}

		hprices := exchange.GetHistoricalPrices(ticker.Symbol, maTimeframe, 50)
		if len(hprices) == 0 {
			continue
		}

		p := grid.OnTickParams{
			Ticker:    *ticker,
			BotParams: params,
			HPrices:   hprices,
			DB:        *db,
		}

		tradeOrders := grid.OnTick(p)
		if tradeOrders == nil {
			continue
		}

		// Open new orders ---------------------------------------------------------

		for _, o := range tradeOrders.OpenOrders {
			o.ID = h.GenID()
			exo, err := exchange.PlaceLimitOrder(o)
			if err != nil {
				h.Log("OpenOrder")
				os.Exit(1)
			}
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
			h.Logf("{Action: Open, Side: %s, Qty: %.4f, Zone: %.2f, Price: %.2f}\n", o.Side, o.Qty, o.ZonePrice, o.OpenPrice)
		}

		// Synchronize order status ------------------------------------------------

		qo := t.Order{
			BotID:    botID,
			Exchange: t.ExcBinance,
			Symbol:   symbol,
		}
		for _, o := range db.GetLimitOrders(qo) {
			exo, err := exchange.GetOrder(o)
			if err != nil {
				h.Log("GetLimitOrders", err)
				os.Exit(1)
			}
			if exo == nil || exo.Status == t.OrderStatusNew {
				continue
			}

			// Synchronize FILLED/CANCELED order
			if o.Status != exo.Status {
				o.Status = exo.Status
				o.UpdateTime = exo.UpdateTime
				err := db.UpdateOrder(o)
				if err != nil {
					h.Log(err)
					continue
				}
			}
			if exo.Status == t.OrderStatusCanceled {
				continue
			}

			// Place a new Take Profit order
			const maxNumAlgoOrders = 5
			if o.TPPrice > 0 && db.GetTPOrder(o.ID) == nil {
				// Remove the highest price TP order, because of 'MAX_NUM_ALGO_ORDERS=5'
				tpOrders := db.GetNewTPOrders(qo)
				if len(tpOrders) == maxNumAlgoOrders {
					_tpo := tpOrders[len(tpOrders)-1]
					// Ignore the far TP price, keep calm and waiting
					if _tpo.OpenPrice < o.TPPrice {
						continue
					}
					exo, err := exchange.CancelOrder(_tpo)
					if err != nil {
						h.Log("CancelOrder", err)
						continue
					}
					if exo == nil {
						continue
					}
					_tpo.Status = exo.Status
					_tpo.UpdateTime = exo.UpdateTime
					err = db.UpdateOrder(_tpo)
					if err != nil {
						h.Log(err)
						continue
					}
					h.Logf("{Action: CancelTP, Side: %s, Qty: %.2f, Price: %.2f}\n",
						_tpo.Side, _tpo.Qty, _tpo.OpenPrice)
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
					OpenPrice:   o.TPPrice,
					StopPrice:   h.CalcTPStop(o.Side, o.TPPrice, 0, digits),
				}
				exo, err := exchange.PlaceStopOrder(tpo)
				if err != nil {
					h.Log("PlaceTPOrder", tpo)
					os.Exit(1)
				}
				if exo == nil {
					continue
				}
				tpo.RefID = exo.RefID
				tpo.OpenTime = exo.OpenTime
				err = db.CreateOrder(tpo)
				if err != nil {
					h.Log(err)
					continue
				}
				h.Logf("{Action: NewTP, Side: %s, Qty: %.2f, Zone: %.2f, Price: %.2f, OpenPrice: %.2f",
					tpo.Side, tpo.Qty, o.ZonePrice, tpo.OpenPrice, o.OpenPrice)
			}
		}

		// Synchronize TP order status ---------------------------------------------

		for _, tpo := range db.GetTPOrders(qo) {
			exo, err := exchange.GetOrder(tpo)
			if err != nil {
				h.Log("GetTPOrders")
				os.Exit(1)
			}
			if exo == nil || exo.Status == t.OrderStatusNew {
				continue
			}

			if tpo.Status != exo.Status {
				tpo.Status = exo.Status
				tpo.UpdateTime = exo.UpdateTime
				err := db.UpdateOrder(tpo)
				if err != nil {
					h.Log(err)
					continue
				}
			}
			if exo.Status == t.OrderStatusCanceled {
				continue
			}

			oo := db.GetOrderByID(tpo.OpenOrderID)
			if (oo.Side == t.OrderSideBuy && ticker.Price > tpo.OpenPrice && oo.CloseOrderID == "") ||
				(oo.Side == t.OrderSideSell && ticker.Price < tpo.OpenPrice && oo.CloseOrderID == "") {
				oo.PL = (tpo.OpenPrice - oo.OpenPrice) * oo.Qty
				oo.CloseOrderID = tpo.ID
				oo.ClosePrice = tpo.OpenPrice
				oo.CloseTime = h.Now13()
				err := db.UpdateOrder(*oo)
				if err != nil {
					h.Log(err)
					continue
				}
				h.Logf("{Action: TP, Side: %s, Qty: %.2f, Zone: %.2f, Price: %.2f, OpenPrice: %.2f, Profit: %.2f}\n",
					tpo.Side, tpo.Qty, oo.ZonePrice, tpo.OpenPrice, oo.OpenPrice, math.Abs(tpo.OpenPrice-oo.OpenPrice)*tpo.Qty)
			}
		}
	}
}
