package exchange

import (
	"strings"

	binance "github.com/tonkla/autotp/exchange/binance/spot"
	"github.com/tonkla/autotp/types"
)

type Exchangeable interface {
	GetName() string
	GetTicker(symbol string) types.Ticker
	GetHistoricalPrices(symbol string, interval string, limit int) []types.HisPrice
	OpenOrder(types.Order) *types.Order
	CloseOrder(types.Order) *types.Order
	CloseOrderByID(string) *types.Order
}

func New(name string) Exchangeable {
	var ex Exchangeable
	_name := strings.ToUpper(name)
	if _name == types.EXC_BINANCE {
		ex = binance.New()
	}
	//  else if _name == types.EXC_BITKUB {
	// 		ex = bitkub.New()
	// 	} else if _name == types.EXC_SATANG {
	// 		ex = satang.New()
	// 	}
	return ex
}
