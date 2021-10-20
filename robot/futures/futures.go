package futures

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tonkla/autotp/exchange"
	binance "github.com/tonkla/autotp/exchange/binance/futures"
	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/rdb"
	strategy "github.com/tonkla/autotp/strategy/trend"
	t "github.com/tonkla/autotp/types"
)

type AppParams struct {
	DB     rdb.DB
	Repo   exchange.Repository
	Ticker t.Ticker
	TO     t.TradeOrders
	QO     t.QueryOrder
	BP     t.BotParams
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

func CommonFutures() {
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
	startPrice := viper.GetFloat64("startPrice")
	baseQty := viper.GetFloat64("baseQty")
	quoteQty := viper.GetFloat64("quoteQty")
	intervalSec := viper.GetInt64("intervalSec")
	maTimeframe := viper.GetString("maTimeframe")
	maPeriod := viper.GetInt64("maPeriod")
	autoSL := viper.GetBool("autoSL")
	autoTP := viper.GetBool("autoTP")
	quoteSL := viper.GetFloat64("quoteSL")
	quoteTP := viper.GetFloat64("quoteTP")
	atrSL := viper.GetFloat64("atrSL")
	atrTP := viper.GetFloat64("atrTP")
	orderGap := viper.GetFloat64("orderGap")
	orderType := viper.GetString("orderType")
	view := viper.GetString("view")

	slStop := viper.GetInt64("slStop")
	slLimit := viper.GetInt64("slLimit")
	tpStop := viper.GetInt64("tpStop")
	tpLimit := viper.GetInt64("tpLimit")
	openLimit := viper.GetInt64("openLimit")

	db := rdb.Connect(dbName)

	ex := binance.NewFuturesClient(apiKey, secretKey)

	bp := t.BotParams{
		BotID:       botID,
		PriceDigits: priceDigits,
		QtyDigits:   qtyDigits,
		BaseQty:     baseQty,
		QuoteQty:    quoteQty,
		MATimeframe: maTimeframe,
		MAPeriod:    maPeriod,
		AutoSL:      autoSL,
		AutoTP:      autoTP,
		QuoteSL:     quoteSL,
		QuoteTP:     quoteTP,
		AtrSL:       atrSL,
		AtrTP:       atrTP,
		OrderGap:    orderGap,
		View:        view,
	}
	if orderType == t.OrderTypeLimit {
		bp.SLim = t.StopLimit{
			SLStop:    slStop,
			SLLimit:   slLimit,
			TPStop:    tpStop,
			TPLimit:   tpLimit,
			OpenLimit: openLimit,
		}
	}

	qo := t.QueryOrder{
		BotID:    botID,
		Exchange: t.ExcBinance,
		Symbol:   symbol,
	}

	p := AppParams{
		DB:   *db,
		Repo: ex,
		QO:   qo,
		BP:   bp,
	}

	if intervalSec == 0 {
		intervalSec = 3
	}

	h.Logf("{Exchange:BinanceFutures Symbol:%s BotID:%d Strategy:TF}\n", symbol, botID)

	for range time.Tick(time.Duration(intervalSec) * time.Second) {
		ticker := ex.GetTicker(symbol)
		if ticker == nil || ticker.Price <= 0 {
			continue
		}

		if startPrice > 0 && ticker.Price > startPrice && len(db.GetActiveOrders(qo)) == 0 {
			continue
		}

		var _period int64 = 50
		const n int64 = 4
		if maPeriod*n > _period {
			_period *= (n - 1)
		}
		hprices := ex.GetHistoricalPrices(symbol, maTimeframe, int(_period))
		if len(hprices) == 0 || hprices[0].Open == 0 {
			continue
		}

		tradeOrder := strategy.OnTick(strategy.OnTickParams{
			DB:        *db,
			Ticker:    *ticker,
			BotParams: bp,
			HPrices:   hprices,
			IsFutures: true,
		})
		if tradeOrder == nil {
			continue
		}

		p.Ticker = *ticker
		p.TO = *tradeOrder
		if orderType == t.OrderTypeLimit {
			// robot.PlaceAsMaker(&p)
		} else if orderType == t.OrderTypeMarket {
			// robot.PlaceAsTaker(&p)
		}
	}
}
