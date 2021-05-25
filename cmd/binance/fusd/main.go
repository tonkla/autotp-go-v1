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
	strategy "github.com/tonkla/autotp/strategy/grid"
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

	symbol := viper.GetString("symbol")
	lowerPrice := viper.GetFloat64("lowerPrice")
	upperPrice := viper.GetFloat64("upperPrice")
	grids := viper.GetInt64("grids")
	qty := viper.GetFloat64("qty")
	view := viper.GetString("view")
	sl := viper.GetFloat64("gridSL")
	tp := viper.GetFloat64("gridTP")
	triggerPrice := viper.GetFloat64("triggerPrice")
	intervalSec := viper.GetInt64("intervalSec")

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

	params := types.GridParams{
		LowerPrice:   lowerPrice,
		UpperPrice:   upperPrice,
		Grids:        grids,
		Qty:          qty,
		View:         view,
		SL:           sl,
		TP:           tp,
		TriggerPrice: triggerPrice,
	}

	db := db.Connect()
	hp := helper{db: db}

	if intervalSec == 0 {
		intervalSec = 5
	}
	for range time.Tick(time.Duration(intervalSec) * time.Second) {
		ticker := binance.GetTicker(symbol)
		if ticker == nil {
			continue
		}
		orders := strategy.OnTick(ticker, params, hp)
		for _, ord := range orders {
			order := binance.Trade(ord)
			if order == nil {
				continue
			}
			err := db.CreateOrder(order)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("%s %.4f of %s at $%.2f (%s)\n", order.Side, order.Qty, order.Symbol, order.Price, order.Exchange)
		}
	}
}

type helper struct {
	db *db.DB
}

func (h helper) DoesOrderExists(order *types.Order) bool {
	return h.db.DoesOrderExists(order)
}
