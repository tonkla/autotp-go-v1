package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	rdb "github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/exchange"
	"github.com/tonkla/autotp/strategy"
	s "github.com/tonkla/autotp/strategy/common"

	"github.com/tonkla/autotp/app"
	h "github.com/tonkla/autotp/helper"
	t "github.com/tonkla/autotp/types"
)

type AppParams struct {
	EX exchange.Repository
	ST s.Repository
	BP *t.BotParams
	TK *t.Ticker
	TO *t.TradeOrders
	QO *t.QueryOrder
}

var rootCmd = &cobra.Command{
	Use:   "autotp",
	Short: "AutoTP: Auto Take Profit (Trend)",
	Long:  "AutoTP: Auto Trading Platform (Trend)",
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

	bp := t.BotParams{
		ApiKey:    viper.GetString("apiKey"),
		SecretKey: viper.GetString("secretKey"),
		DbName:    viper.GetString("dbName"),
		OrderType: viper.GetString("orderType"),
		View:      viper.GetString("view"),

		IntervalSec: viper.GetInt64("intervalSec"),

		BotID:       viper.GetInt64("botID"),
		Exchange:    viper.GetString("exchange"),
		Product:     viper.GetString("product"),
		Symbol:      viper.GetString("symbol"),
		Strategy:    viper.GetString("strategy"),
		PriceDigits: viper.GetInt64("priceDigits"),
		QtyDigits:   viper.GetInt64("qtyDigits"),
		BaseQty:     viper.GetFloat64("baseQty"),
		QuoteQty:    viper.GetFloat64("quoteQty"),

		StartPrice: viper.GetFloat64("startPrice"),
		UpperPrice: viper.GetFloat64("upperPrice"),
		LowerPrice: viper.GetFloat64("lowerPrice"),
		GridSize:   viper.GetFloat64("gridSize"),
		GridTP:     viper.GetFloat64("gridTP"),
		OpenZones:  viper.GetInt64("openZones"),
		ApplyTrend: viper.GetBool("applyTrend"),

		OrderGap: viper.GetFloat64("orderGap"),
		MoS:      viper.GetFloat64("mos"),
		Slippage: viper.GetFloat64("slippage"),

		MATimeframe: viper.GetString("maTimeframe"),
		MAPeriod:    viper.GetInt64("maPeriod"),
		AutoSL:      viper.GetBool("autoSL"),
		AutoTP:      viper.GetBool("autoTP"),
		QuoteSL:     viper.GetFloat64("quoteSL"),
		QuoteTP:     viper.GetFloat64("quoteTP"),
		AtrSL:       viper.GetFloat64("atrSL"),
		AtrTP:       viper.GetFloat64("atrTP"),

		SLim: t.StopLimit{
			SLStop:    viper.GetInt64("slStop"),
			SLLimit:   viper.GetInt64("slLimit"),
			TPStop:    viper.GetInt64("tpStop"),
			TPLimit:   viper.GetInt64("tpLimit"),
			OpenLimit: viper.GetInt64("openLimit"),
		},
	}

	db := rdb.Connect(bp.DbName)

	ex, err := exchange.New(&bp)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	st, err := strategy.New(*db, bp)
	if err != nil {

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	qo := t.QueryOrder{
		BotID:    bp.BotID,
		Exchange: t.ExcBinance,
		Symbol:   bp.Symbol,
	}

	ap := app.AppParams{
		EX: ex,
		ST: st,
		BP: &bp,
		DB: db,
		QO: qo,
	}

	h.Logf("{Exchange:%s Product:%s Symbol:%s Strategy:%s BotID:%d}\n", bp.Exchange, bp.Product, bp.Symbol, bp.Strategy, bp.BotID)

	for range time.Tick(time.Duration(bp.IntervalSec) * time.Second) {
		ticker := ex.GetTicker(bp.Symbol)
		if ticker == nil || ticker.Price <= 0 {
			continue
		}

		if bp.StartPrice > 0 && ticker.Price > bp.StartPrice && len(db.GetActiveOrders(qo)) == 0 {
			continue
		}

		var _period int64 = 50
		const n int64 = 4
		if bp.MAPeriod*n > _period {
			_period *= (n - 1)
		}
		hprices := ex.GetHistoricalPrices(bp.Symbol, bp.MATimeframe, int(_period))
		if len(hprices) == 0 || hprices[0].Open == 0 {
			continue
		}

		bp.HPrices = hprices

		tradeOrders := ap.ST.OnTick(*ticker)

		ap.TK = ticker
		ap.TO = &tradeOrders

		if bp.OrderType == t.OrderTypeLimit {
			// PlaceAsMaker(&ap)
		} else if bp.OrderType == t.OrderTypeMarket {
			// PlaceAsTaker(&ap)
		}
	}
}
