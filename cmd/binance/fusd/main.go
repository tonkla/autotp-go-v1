package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

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
	symbol       string
	lowerPrice   float64
	upperPrice   float64
	grids        int64
	qty          float64
	triggerPrice float64
)

func init() {
	rootCmd.Flags().StringVarP(&symbol, "symbol", "s", "", "Symbol (required)")
	rootCmd.MarkFlagRequired("symbol")

	rootCmd.Flags().Float64VarP(&lowerPrice, "lowerPrice", "l", 0, "Lower Price (required)")
	rootCmd.MarkFlagRequired("lowerPrice")

	rootCmd.Flags().Float64VarP(&upperPrice, "upperPrice", "u", 0, "Upper Price (required)")
	rootCmd.MarkFlagRequired("upperPrice")

	rootCmd.Flags().Int64VarP(&grids, "grids", "g", 0, "Grids (required)")
	rootCmd.MarkFlagRequired("grids")

	rootCmd.Flags().Float64VarP(&qty, "qty", "q", 0, "Quantity per order (required)")
	rootCmd.MarkFlagRequired("qty")

	rootCmd.Flags().Float64VarP(&triggerPrice, "triggerPrice", "t", 0, "Trigger Price")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if upperPrice <= lowerPrice {
		fmt.Fprintln(os.Stderr, "An upper price must be greater than a lower price")
		os.Exit(1)
	} else if grids < 2 {
		fmt.Fprintln(os.Stderr, "Size of the grids must be greater than 1")
		os.Exit(1)
	} else if qty <= 0 {
		fmt.Fprintln(os.Stderr, "Quantity per grid must be greater than 0")
		os.Exit(1)
	}

	params := types.GridParams{
		LowerPrice: lowerPrice,
		UpperPrice: upperPrice,
		Grids:      grids,
		Qty:        qty,
	}

	ticker := binance.GetTicker(symbol)

	orders := strategy.OnTick(ticker, params)
	for _, order := range orders {
		result := binance.Trade(order)
		if result == nil {
			continue
		}

		err := db.CreateRecord(*result)
		if err != nil {
			log.Println(err)
		}
	}
}
