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
	applyTrend := viper.GetBool("applyTrend")
	openAll := viper.GetBool("openAllZones")
	baseQty := viper.GetFloat64("baseQty")
	quoteQty := viper.GetFloat64("quoteQty")
	intervalSec := viper.GetInt64("intervalSec")
	autoTP := viper.GetBool("autoTP")

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

	h.Logf("{Exchange:BinanceSpot Symbol:%s BotID:%d}\n", symbol, botID)

	params := t.BotParams{
		BotID:      botID,
		UpperPrice: upperPrice,
		LowerPrice: lowerPrice,
		GridSize:   gridSize,
		GridTP:     gridTP,
		ApplyTrend: applyTrend,
		OpenAll:    openAll,
		Qty:        baseQty,
		View:       "LONG",
		AutoTP:     autoTP,
	}

	db := db.Connect(dbName)

	exchange := binance.NewSpotClient(apiKey, secretKey)

	if intervalSec == 0 {
		intervalSec = 3
	}

	for range time.Tick(time.Duration(intervalSec) * time.Second) {
		ticker := exchange.GetTicker(symbol)
		if ticker == nil || ticker.Price <= 0 {
			continue
		}

		qo := t.Order{
			BotID:    botID,
			Exchange: t.ExcBinance,
			Symbol:   symbol,
		}

		if ticker.Price > startPrice && len(db.GetLimitOrders(qo)) == 0 {
			continue
		}

		p := grid.OnTickParams{
			Ticker:    *ticker,
			BotParams: params,
			D1HPrices: exchange.Get1dHistoricalPrices(symbol, 1),
			H1HPrices: exchange.Get1hHistoricalPrices(symbol, 1),
			DB:        *db,
		}

		tradeOrders := grid.OnTick(p)
		if tradeOrders == nil {
			continue
		}

		// Open new orders ---------------------------------------------------------

		for _, o := range tradeOrders.OpenOrders {
			o.ID = h.GenID()
			if o.Qty == 0 {
				o.Qty = h.RoundToDigits(quoteQty/o.OpenPrice, qtyDigits)
				if o.Qty <= 0 {
					h.Log("Quantity must be greater than zero")
					continue
				}
			}
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
			o.OpenPrice = exo.OpenPrice
			o.OpenTime = exo.OpenTime
			err = db.CreateOrder(o)
			if err != nil {
				h.Log("CreateOrder", err)
				continue
			}
			log := t.LogOrder{
				Action: "New",
				Qty:    o.Qty,
				Open:   o.OpenPrice,
				Zone:   o.ZonePrice,
			}
			h.Log(log)
		}

		// Synchronize order status / Take Profit ----------------------------------

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
					h.Log("UpdateOrder", err)
					continue
				}
				if exo.Status == t.OrderStatusFilled {
					log := t.LogOrder{
						Action: "Filled",
						Qty:    o.Qty,
						Open:   o.OpenPrice,
						Zone:   o.ZonePrice,
					}
					h.Log(log)
				}
			}
			if exo.Status == t.OrderStatusCanceled {
				continue
			}

			// Close the order at the market price
			if o.TPPrice > 0 && o.CloseOrderID == "" && db.GetTPOrder(o.ID) == nil {
				book := exchange.GetOrderBook(symbol, 5)
				if book == nil || len(book.Bids) == 0 {
					continue
				}
				sellPrice := book.Bids[0].Price
				if o.TPPrice > sellPrice || sellPrice == 0 {
					continue
				}

				tpo := t.Order{
					BotID:       o.BotID,
					Exchange:    o.Exchange,
					Symbol:      o.Symbol,
					ID:          h.GenID(),
					OpenOrderID: o.ID,
					Qty:         o.Qty,
					Side:        h.Reverse(o.Side),
					Type:        t.OrderTypeMarket, // Place with a MARKET type
					Status:      t.OrderStatusNew,
				}
				exo, err := exchange.PlaceMarketOrder(tpo)
				if err != nil {
					h.Log("PlaceMarketOrder", tpo)
					os.Exit(1)
				}
				if exo == nil {
					continue
				}

				tpo.Type = t.OrderTypeTP // Save to local DB with a TAKE_PROFIT_LIMIT type
				tpo.RefID = exo.RefID
				tpo.OpenPrice = exo.OpenPrice
				tpo.Qty = exo.Qty
				tpo.Commission = exo.Commission
				tpo.OpenTime = exo.OpenTime
				tpo.Status = exo.Status
				err = db.CreateOrder(tpo)
				if err != nil {
					h.Log("CreateTPOrder", err)
					continue
				}

				o.CloseOrderID = tpo.ID
				o.ClosePrice = tpo.OpenPrice
				o.CloseTime = tpo.OpenTime
				o.PL = h.RoundToDigits(((o.ClosePrice-o.OpenPrice)*tpo.Qty)-tpo.Commission, priceDigits)
				err = db.UpdateOrder(o)
				if err != nil {
					h.Log("UpdateOrder", err)
					continue
				}
				log := t.LogOrder{
					Action: "TP",
					Qty:    tpo.Qty,
					Close:  o.ClosePrice,
					Open:   o.OpenPrice,
					Zone:   o.ZonePrice,
					Profit: o.PL,
				}
				h.Log(log)
			}
		}
	}
}
