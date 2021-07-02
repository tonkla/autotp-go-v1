package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tonkla/autotp/db"
	binance "github.com/tonkla/autotp/exchange/binance/fusd"
	"github.com/tonkla/autotp/strategy/trend"
	"github.com/tonkla/autotp/types"
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

	log.Printf("I'm a bot ID %d, working on Binance's USDⓈ-M Futures\n", botID)

	params := types.BotParams{
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

	for range time.Tick(time.Duration(intervalSec) * time.Second) {
		ticker := binance.GetTicker(symbol)
		if ticker == nil || ticker.Price <= 0 {
			continue
		}

		orderBook := binance.GetOrderBook(symbol, 5)
		if orderBook == nil {
			continue
		}

		hprices := binance.GetHistoricalPrices(ticker.Symbol, maTimeframe, 100)
		if len(hprices) == 0 {
			continue
		}

		p := trend.OnTickParams{
			Ticker:    *ticker,
			OrderBook: *orderBook,
			BotParams: params,
			HPrices:   hprices,
			DB:        *db,
		}

		tradeOrders := trend.OnTick(p)
		if tradeOrders == nil {
			continue
		}

		for _, order := range tradeOrders.CloseOrders {
			if binance.NewOrder(order, apiKey, secretKey) == nil {
				continue
			}
			order.Status = types.ORDER_STATUS_CLOSED
			err := db.UpdateOrder(order)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("\t%s %.4f of %s at $%.2f (%s)\n",
				order.Side, order.Qty, order.Symbol, order.OpenPrice, order.Exchange)
		}

		for _, order := range tradeOrders.OpenOrders {
			if binance.NewOrder(order, apiKey, secretKey) == nil {
				continue
			}
			err := db.CreateOrder(order)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("\t%s %.4f of %s at $%.2f (%s)\n",
				order.Side, order.Qty, order.Symbol, order.OpenPrice, order.Exchange)
		}
	}
}
