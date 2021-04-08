package exchange

import (
	"strings"

	"github.com/tonkla/autotp/common"
	"github.com/tonkla/autotp/exchange/binance"
	"github.com/tonkla/autotp/exchange/bitkub"
	"github.com/tonkla/autotp/exchange/satang"
)

type Exchangeable interface {
	GetName() string
	GetTicker(symbol string) common.Ticker
	GetHistoricalPrices(symbol string, interval string, limit int) []common.HisPrice
}

func New(name string) Exchangeable {
	var ex Exchangeable
	_name := strings.ToUpper(name)
	if _name == common.EXC_BINANCE {
		ex = binance.New()
	} else if _name == common.EXC_BITKUB {
		ex = bitkub.New()
	} else if _name == common.EXC_SATANG {
		ex = satang.New()
	}
	return ex
}
