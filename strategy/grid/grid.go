package grid

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

var rootCmd = &cobra.Command{
	Use:   "autotp-grid",
	Short: "AutoTP: Grid Strategy",
	Long:  "AutoTP: Grid Strategy",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var (
	lowerPrice   float64
	upperPrice   float64
	grids        int64
	qty          float64
	triggerPrice float64
)

func init() {
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

func OnTick(ticker types.Ticker) []types.Order {
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

	buyPrice, sellPrice, gridWidth := helper.GetGridRange(ticker.Price, lowerPrice, upperPrice, float64(grids))

	var orders []types.Order

	// Has already bought at this price?
	record := db.GetRecordByPrice(buyPrice, types.SIDE_BUY)
	if record == nil {
		orders = append(orders, types.Order{
			Symbol: ticker.Symbol,
			Side:   types.SIDE_BUY,
			Price:  buyPrice,
			Qty:    qty,
			TP:     buyPrice + (gridWidth * 2),
		})
	}

	// Has already sold at this price?
	record = db.GetRecordByPrice(sellPrice, types.SIDE_SELL)
	if record == nil {
		orders = append(orders, types.Order{
			Symbol: ticker.Symbol,
			Side:   types.SIDE_SELL,
			Price:  sellPrice,
			Qty:    qty,
			TP:     sellPrice - (gridWidth * 2),
		})
	}

	return orders
}
