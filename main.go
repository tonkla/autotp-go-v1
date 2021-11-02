package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tonkla/autotp/app"
	"github.com/tonkla/autotp/exchange"
	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/rdb"
	"github.com/tonkla/autotp/robot"
	"github.com/tonkla/autotp/strategy"
	t "github.com/tonkla/autotp/types"
)

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

		Exchange:    viper.GetString("exchange"),
		Symbol:      viper.GetString("symbol"),
		BotID:       viper.GetInt64("botID"),
		Product:     viper.GetString("product"),
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
		ApplyTA:    viper.GetBool("applyTA"),
		Slippage:   viper.GetFloat64("slippage"),

		MATf1st:     viper.GetString("maTf1st"),
		MAPeriod1st: viper.GetInt64("maPeriod1st"),
		MATf2nd:     viper.GetString("maTf2nd"),
		MAPeriod2nd: viper.GetInt64("maPeriod2nd"),
		MATf3rd:     viper.GetString("maTf3rd"),
		MAPeriod3rd: viper.GetInt64("maPeriod3rd"),
		OrderGap:    viper.GetFloat64("orderGap"),
		MoS:         viper.GetFloat64("mos"),

		AutoSL:  viper.GetBool("autoSL"),
		AutoTP:  viper.GetBool("autoTP"),
		QuoteSL: viper.GetFloat64("quoteSL"),
		QuoteTP: viper.GetFloat64("quoteTP"),
		AtrSL:   viper.GetFloat64("atrSL"),
		AtrTP:   viper.GetFloat64("atrTP"),

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

	st, err := strategy.New(db, &bp, ex)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	qo := t.QueryOrder{
		BotID:    bp.BotID,
		Exchange: bp.Exchange,
		Symbol:   bp.Symbol,
	}

	ap := app.AppParams{
		EX: ex,
		ST: st,
		DB: db,
		BP: &bp,
		QO: qo,
	}

	h.Logf("{Exchange:%s Product:%s Symbol:%s Strategy:%s BotID:%d}\n",
		bp.Exchange, bp.Product, bp.Symbol, bp.Strategy, bp.BotID)

	for range time.Tick(time.Duration(bp.IntervalSec) * time.Second) {
		ticker := ex.GetTicker(bp.Symbol)
		if ticker == nil || ticker.Price <= 0 {
			continue
		}
		ap.TK = *ticker
		tradeOrders := ap.ST.OnTick(*ticker)
		if tradeOrders != nil {
			ap.TO = *tradeOrders
			robot.Trade(&ap)
		}
	}
}
