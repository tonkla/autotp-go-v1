package main

import (
	"fmt"
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
	botID := viper.GetInt64("botID")
	symbol := viper.GetString("symbol")
	lowerPrice := viper.GetFloat64("lowerPrice")
	upperPrice := viper.GetFloat64("upperPrice")
	grids := viper.GetFloat64("grids")
	qty := viper.GetFloat64("qty")
	view := viper.GetString("view")
	sl := viper.GetFloat64("gridSL")
	tp := viper.GetFloat64("gridTP")
	triggerPrice := viper.GetFloat64("triggerPrice")
	slippage := viper.GetFloat64("slippage")
	intervalSec := viper.GetInt64("intervalSec")
	maTimeframe := viper.GetString("maTimeframe")
	maPeriod := viper.GetInt64("maPeriod")

	if upperPrice <= lowerPrice {
		fmt.Fprintln(os.Stderr, "The upper price must be greater than the lower price")
		os.Exit(1)
	} else if grids < 2 {
		fmt.Fprintln(os.Stderr, "Size of the grids must be greater than 1")
		os.Exit(1)
	} else if qty <= 0 {
		fmt.Fprintln(os.Stderr, "Quantity per grid must be greater than 0")
		os.Exit(1)
	}

	h.Logf("I'm a bot ID [%d], trading [%s] on a Binance's Spot\n", botID, symbol)

	params := t.BotParams{
		BotID:        botID,
		LowerPrice:   lowerPrice,
		UpperPrice:   upperPrice,
		Grids:        grids,
		Qty:          qty,
		View:         view,
		SL:           sl,
		TP:           tp,
		TriggerPrice: triggerPrice,
		Slippage:     slippage,
		MATimeframe:  maTimeframe,
		MAPeriod:     maPeriod,
	}

	db := db.Connect()

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

		// Close orders by using late SL/TP -----------------------------------------

		for _, _o := range tradeOrders.CloseOrders {
			err := db.UpdateOrder(_o)
			if err != nil {
				h.Log(err)
			}
		}

		// Open a new order --------------------------------------------------------

		for _, _o := range tradeOrders.OpenOrders {
			o := exchange.PlaceOrder(_o)
			if o == nil {
				continue
			}
			err := db.CreateOrder(*o)
			if err != nil {
				h.Log(err)
				continue
			}
			h.Logf("Open {side: %s, qty: %.4f, price: %.2f}\n", o.Side, o.Qty, o.OpenPrice)
		}

		// Place SL/TP with the order type STOP_LOSS_LIMIT / TAKE_PROFIT_LIMIT

		qo := t.Order{
			BotID:    botID,
			Exchange: t.ExcBinance,
			Symbol:   symbol,
		}
		for _, o := range db.GetNewOrders(qo) {
			exo := exchange.GetOrder(o)
			if exo == nil || exo.Status == t.OrderStatusNew {
				continue
			}

			if exo.Status == t.OrderStatusCanceled || !exo.IsWorking {
				o.CloseTime = h.Now13()
				o.Status = t.OrderStatusCanceled
				if exo.Status != t.OrderStatusCanceled && !exo.IsWorking {
					o.Status = t.OrderStatusClosed
				}
				err := db.UpdateOrder(o)
				if err != nil {
					h.Log(err)
				}
				continue
			}

			// Swap a side before sending to Binance, because SL/TP needs an opposite side
			if o.Side == t.OrderSideBuy {
				o.Side = t.OrderSideSell
			} else if o.Side == t.OrderSideSell {
				o.Side = t.OrderSideBuy
			}

			if o.SL > 0 {
				o.Type = t.OrderTypeSL
				o.ClosePrice = o.SL
				if nil == exchange.PlaceOrder(o) {
					continue
				}
			}

			if o.TP > 0 {
				o.Type = t.OrderTypeTP
				o.ClosePrice = o.TP
				if nil == exchange.PlaceOrder(o) {
					continue
				}
			}

			// Swap back a side to the original one before update to DB
			if o.Side == t.OrderSideBuy {
				o.Side = t.OrderSideSell
			} else if o.Side == t.OrderSideSell {
				o.Side = t.OrderSideBuy
			}

			o.Status = exo.Status
			err := db.UpdateOrder(o)
			if err != nil {
				h.Log(err)
			}
		}

		// Update closed orders in the local DB

		for _, o := range db.GetFilledOrders(qo) {
			exo := exchange.GetOrder(o)
			if exo == nil || exo.IsWorking {
				continue
			}

			o.Status = t.OrderStatusClosed
			o.CloseTime = h.Now13()
			err := db.UpdateOrder(o)
			if err != nil {
				h.Log(err)
				continue
			}
			h.Logf("Close {side: %s, qty: %.4f, price: %.2f}\n", o.Side, o.Qty, o.ClosePrice)
		}
	}
}
